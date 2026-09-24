use std::sync::atomic::AtomicU64;
use std::sync::Arc;

use tokio::sync::watch;
use tracing::{info, warn};

use crate::camera::{
    redact_rtsp_url, CameraRuntimeState, CameraStatus, FpsEstimator, LiveCaptureContext,
    ReconnectBackoff,
};
use crate::config::Config;
use crate::decode::{AccelerationRuntime, SessionDecodeContext};
use crate::metrics::ProcessorMetrics;
use crate::pipeline::{run_frame_consumer, FramePipeline};
use crate::rtsp::{run_rtsp_frame_loop, simulate_frame_loop};

async fn run_session_with_pipeline(
    camera_id: i64,
    rtsp_url: &str,
    cfg: &Config,
    metrics: Arc<ProcessorMetrics>,
    acceleration: Arc<AccelerationRuntime>,
    states: Arc<tokio::sync::RwLock<std::collections::HashMap<i64, CameraRuntimeState>>>,
    shutdown_rx: watch::Receiver<bool>,
    global_frames: Arc<AtomicU64>,
    simulate: bool,
) -> Result<crate::rtsp::RtspLoopStats, crate::error::AppError> {
    let (session_shutdown_tx, session_shutdown_rx) = watch::channel(false);
    let pipeline = FramePipeline::new(
        cfg.frame_buffer_max,
        metrics.clone(),
        states.clone(),
        camera_id,
    );

    let decode_ctx = SessionDecodeContext::new();
    let consumer_pipeline = pipeline.clone();
    let global_shutdown = shutdown_rx.clone();
    let decode_for_consumer = decode_ctx.clone();
    let acceleration_for_consumer = acceleration.clone();
    let consumer = tokio::spawn(async move {
        run_frame_consumer(
            consumer_pipeline,
            session_shutdown_rx,
            global_shutdown,
            decode_for_consumer,
            acceleration_for_consumer,
            !simulate,
        )
        .await;
    });

    let mut fps_est = FpsEstimator::new();
    let mut live = LiveCaptureContext {
        metrics: metrics.as_ref(),
        global_frames: global_frames.as_ref(),
        states: &states,
        camera_id,
        fps_est: &mut fps_est,
    };

    let stats = if simulate {
        simulate_frame_loop(
            camera_id,
            5.0,
            shutdown_rx.clone(),
            Some(&mut live),
            &pipeline,
        )
        .await
    } else {
        run_rtsp_frame_loop(
            camera_id,
            rtsp_url,
            cfg.rtsp_connect_timeout,
            cfg.rtsp_frame_timeout,
            shutdown_rx.clone(),
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
    states: Arc<tokio::sync::RwLock<std::collections::HashMap<i64, CameraRuntimeState>>>,
    shutdown_rx: &mut watch::Receiver<bool>,
    global_frames: Arc<AtomicU64>,
) {
    let redacted = redact_rtsp_url(&rtsp_url);
    let mut backoff = ReconnectBackoff::new(cfg.rtsp_reconnect_base);

    let simulate = std::env::var("RTSP_SIMULATE")
        .map(|v| matches!(v.as_str(), "1" | "true" | "yes"))
        .unwrap_or(false);

    loop {
        if *shutdown_rx.borrow() {
            break;
        }

        update_state(&states, camera_id, |s| {
            s.status = CameraStatus::Reconnecting;
            s.fps = 0.0;
        })
        .await;

        if simulate {
            let cancel = shutdown_rx.clone();
            let _ = run_session_with_pipeline(
                camera_id,
                &rtsp_url,
                &cfg,
                metrics.clone(),
                acceleration.clone(),
                states.clone(),
                cancel,
                global_frames.clone(),
                true,
            )
            .await;
            break;
        }

        let cancel = shutdown_rx.clone();
        match run_session_with_pipeline(
            camera_id,
            &rtsp_url,
            &cfg,
            metrics.clone(),
            acceleration.clone(),
            states.clone(),
            cancel,
            global_frames.clone(),
            false,
        )
        .await
        {
            Ok(_stats) => {
                backoff.reset();
                info!(camera_id, processor_id = %cfg.processor_id, "rtsp session ended — reconnecting");
                continue;
            }
            Err(e) => {
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
                update_state(&states, camera_id, |s| {
                    s.status = CameraStatus::Offline;
                    s.fps = 0.0;
                    s.last_error = Some(e.to_string());
                    s.reconnect_count += 1;
                })
                .await;

                tokio::select! {
                    _ = tokio::time::sleep(delay) => {}
                    _ = shutdown_rx.changed() => { break; }
                }
            }
        }
    }

    update_state(&states, camera_id, |s| {
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

async fn update_state<F>(
    states: &Arc<tokio::sync::RwLock<std::collections::HashMap<i64, CameraRuntimeState>>>,
    camera_id: i64,
    f: F,
) where
    F: FnOnce(&mut CameraRuntimeState),
{
    if let Some(s) = states.write().await.get_mut(&camera_id) {
        f(s);
    }
}
