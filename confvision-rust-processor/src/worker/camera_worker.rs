use std::sync::atomic::AtomicU64;
use std::sync::Arc;

use tokio::sync::watch;
use tracing::{info, warn};

use crate::camera::{
    redact_rtsp_url, CameraRuntimeState, CameraStatus, FpsEstimator, LiveCaptureContext,
    ReconnectBackoff,
};
use crate::config::Config;
use crate::metrics::ProcessorMetrics;
use crate::rtsp::{run_rtsp_frame_loop, simulate_frame_loop};

pub async fn run_camera_worker(
    camera_id: i64,
    rtsp_url: String,
    cfg: Config,
    metrics: Arc<ProcessorMetrics>,
    states: Arc<tokio::sync::RwLock<std::collections::HashMap<i64, CameraRuntimeState>>>,
    shutdown_rx: &mut watch::Receiver<bool>,
    global_frames: Arc<AtomicU64>,
) {
    let redacted = redact_rtsp_url(&rtsp_url);
    let mut backoff = ReconnectBackoff::new(cfg.rtsp_reconnect_base);
    let mut fps_est = FpsEstimator::new();

    loop {
        if *shutdown_rx.borrow() {
            break;
        }

        update_state(&states, camera_id, |s| {
            s.status = CameraStatus::Reconnecting;
            s.fps = 0.0;
        })
        .await;

        let mut live = LiveCaptureContext {
            metrics: metrics.as_ref(),
            global_frames: global_frames.as_ref(),
            states: &states,
            camera_id,
            fps_est: &mut fps_est,
        };

        if std::env::var("RTSP_SIMULATE")
            .map(|v| matches!(v.as_str(), "1" | "true" | "yes"))
            .unwrap_or(false)
        {
            let cancel = shutdown_rx.clone();
            simulate_frame_loop(
                camera_id,
                5.0,
                cancel,
                cfg.frame_buffer_max,
                Some(&mut live),
            )
            .await;
            break;
        }

        let cancel = shutdown_rx.clone();
        match run_rtsp_frame_loop(
            camera_id,
            &rtsp_url,
            cfg.rtsp_connect_timeout,
            cfg.rtsp_frame_timeout,
            cfg.frame_buffer_max,
            cancel,
            Some(&mut live),
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
