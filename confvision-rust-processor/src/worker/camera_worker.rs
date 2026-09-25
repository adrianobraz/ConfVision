use std::sync::atomic::AtomicU64;
use std::sync::Arc;

use tokio::sync::watch;
use tracing::{info, warn};

use crate::camera::{
    redact_rtsp_url, CameraCancel, CameraRuntimeState, CameraStatus, FpsEstimator,
    LiveCaptureContext, ReconnectBackoff, SharedCameraState,
};
use crate::config::Config;
use crate::decode::{AccelerationRuntime, DecodePolicyCoordinator, SessionDecodeContext};
use crate::metrics::ProcessorMetrics;
use crate::pipeline::{run_frame_consumer, FramePipeline};
use crate::rtsp::{run_rtsp_frame_loop, simulate_frame_loop};

async fn run_session_with_pipeline(
    camera_id: i64,
    rtsp_url: &str,
    cfg: &Config,
    metrics: Arc<ProcessorMetrics>,
    acceleration: Arc<AccelerationRuntime>,
    decode_policy: Arc<DecodePolicyCoordinator>,
    state: SharedCameraState,
    cancel: CameraCancel,
    global_frames: Arc<AtomicU64>,
    simulate: bool,
) -> Result<crate::rtsp::RtspLoopStats, crate::error::AppError> {
    let (session_shutdown_tx, session_shutdown_rx) = watch::channel(false);
    let pipeline = FramePipeline::new(cfg.frame_buffer_max, metrics.clone(), state.clone());

    let decode_ctx = SessionDecodeContext::new();
    let consumer_pipeline = pipeline.clone();
    let global_shutdown = cancel.global();
    let decode_for_consumer = decode_ctx.clone();
    let acceleration_for_consumer = acceleration.clone();
    let decode_policy_for_consumer = decode_policy.clone();
    let consumer = tokio::spawn(async move {
        run_frame_consumer(
            consumer_pipeline,
            session_shutdown_rx,
            global_shutdown,
            decode_for_consumer,
            acceleration_for_consumer,
            decode_policy_for_consumer,
            !simulate,
        )
        .await;
    });

    let mut fps_est = FpsEstimator::new();
    let mut live = LiveCaptureContext {
        metrics: metrics.as_ref(),
        global_frames: global_frames.as_ref(),
        state: &state,
        fps_est: &mut fps_est,
    };

    let stats = if simulate {
        simulate_frame_loop(camera_id, 5.0, cancel.clone(), Some(&mut live), &pipeline).await
    } else {
        run_rtsp_frame_loop(
            camera_id,
            rtsp_url,
            cfg.rtsp_connect_timeout,
            cfg.rtsp_frame_timeout,
            cancel,
            Some(&mut live),
            &pipeline,
            decode_ctx.as_ref(),
        )
        .await?
    };

    pipeline.close();
    let _ = session_shutdown_tx.send(true);
    let _ = tokio::time::timeout(std::time::Duration::from_secs(5), consumer).await;

    Ok(stats)
}

pub async fn run_camera_worker(
    camera_id: i64,
    rtsp_url: String,
    cfg: Config,
    metrics: Arc<ProcessorMetrics>,
    acceleration: Arc<AccelerationRuntime>,
    decode_policy: Arc<DecodePolicyCoordinator>,
    state: SharedCameraState,
    shutdown_rx: &mut watch::Receiver<bool>,
    camera_stop_rx: watch::Receiver<bool>,
    global_frames: Arc<AtomicU64>,
) {
    let redacted = redact_rtsp_url(&rtsp_url);
    let mut backoff = ReconnectBackoff::new(cfg.rtsp_reconnect_base);

    let simulate = std::env::var("RTSP_SIMULATE")
        .map(|v| matches!(v.as_str(), "1" | "true" | "yes"))
        .unwrap_or(false);

    loop {
        if *shutdown_rx.borrow() || *camera_stop_rx.borrow() {
            break;
        }

        update_state(&state, |s| {
            s.status = CameraStatus::Reconnecting;
            s.fps = 0.0;
            s.last_error = None;
        })
        .await;

        if simulate {
            let cancel = CameraCancel::new(shutdown_rx.clone(), camera_stop_rx.clone());
            let _ = run_session_with_pipeline(
                camera_id,
                &rtsp_url,
                &cfg,
                metrics.clone(),
                acceleration.clone(),
                decode_policy.clone(),
                state.clone(),
                cancel,
                global_frames.clone(),
                true,
            )
            .await;
            break;
        }

        let cancel = CameraCancel::new(shutdown_rx.clone(), camera_stop_rx.clone());
        match run_session_with_pipeline(
            camera_id,
            &rtsp_url,
            &cfg,
            metrics.clone(),
            acceleration.clone(),
            decode_policy.clone(),
            state.clone(),
            cancel,
            global_frames.clone(),
            false,
        )
        .await
        {
            Ok(_stats) => {
                if *shutdown_rx.borrow() || *camera_stop_rx.borrow() {
                    break;
                }
                backoff.reset();
                info!(camera_id, processor_id = %cfg.processor_id, "rtsp session ended — reconnecting");
                continue;
            }
            Err(e) => {
                if *camera_stop_rx.borrow() {
                    break;
                }
                metrics.record_reconnect();
                metrics.record_rtsp_error();
                let delay = backoff.next_delay();
                warn!(
                    camera_id,
                    processor_id = %cfg.processor_id,
                    error = %e,
                    delay_secs = delay.as_secs(),
                    "rtsp reconnect"
                );
                update_state(&state, |s| {
                    s.status = CameraStatus::Offline;
                    s.fps = 0.0;
                    s.last_error = Some(e.to_string());
                    s.reconnect_count += 1;
                    s.rtsp_errors += 1;
                })
                .await;

                let mut local_stop = camera_stop_rx.clone();
                tokio::select! {
                    _ = tokio::time::sleep(delay) => {}
                    _ = shutdown_rx.changed() => { break; }
                    _ = local_stop.changed() => { break; }
                }
            }
        }
    }

    update_state(&state, |s| {
        s.status = CameraStatus::Stopped;
        s.fps = 0.0;
    })
    .await;

    info!(
        camera_id,
        processor_id = %cfg.processor_id,
        url = %redacted,
        "worker stopped"
    );
}

async fn update_state<F>(state: &SharedCameraState, f: F)
where
    F: FnOnce(&mut CameraRuntimeState),
{
    if let Ok(mut guard) = state.try_write() {
        f(&mut guard);
    } else {
        let mut guard = state.write().await;
        f(&mut guard);
    }
}
