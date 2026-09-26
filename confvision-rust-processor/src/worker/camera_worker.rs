use std::sync::atomic::AtomicU64;
use std::sync::Arc;
use std::time::Duration;

use chrono::{DateTime, Utc};
use tokio::sync::{watch, RwLock};
use tracing::{debug, info, warn};

use crate::stream_policy::{
    classify_rtsp_error, delay_after_failure, StreamFailureClass, StreamHealthAction,
    StreamPolicyState,
};
use crate::worker::stream_health_queue::{enqueue_stream_action, enqueue_stream_failure};

use crate::camera::{
    redact_rtsp_url, CameraCancel, CameraRuntimeState, CameraStatus, FpsEstimator,
    LiveCaptureContext, ReconnectBackoff, SharedCameraState,
};
use crate::config::Config;
use crate::decode::{AccelerationRuntime, DecodePolicyCoordinator, SessionDecodeContext};
use crate::metrics::ProcessorMetrics;
use crate::motion::{MotionEnqueueGate, MotionGatedSession, MotionSensitivity};
use crate::pipeline::{run_frame_consumer, FramePipeline};
use crate::rtsp::{
    connect_rtsp_demuxed, run_rtsp_demux_loop, simulate_frame_loop, RtspDemuxOutcome, RtspLoopStats,
};

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
    let pipeline = FramePipeline::new(cfg.frame_buffer_max, metrics.clone(), state.clone());
    let decode_ctx = SessionDecodeContext::new();
    let motion_gate = if cfg.analysis_only_on_motion {
        Some(MotionGatedSession::new(cfg.motion_gate_miss_frames))
    } else {
        None
    };
    let mut enqueue_gate = MotionEnqueueGate::from_config(cfg, motion_gate.clone());
    let motion_sensitivity = MotionSensitivity::from_config(cfg);

    let mut fps_est = FpsEstimator::new();
    let mut live = LiveCaptureContext {
        metrics: metrics.as_ref(),
        global_frames: global_frames.as_ref(),
        state: &state,
        fps_est: &mut fps_est,
        pending_received: 0,
        pending_fps_frames: 0,
    };

    let stats = if simulate {
        let (session_shutdown_tx, session_shutdown_rx) = watch::channel(false);
        let consumer_pipeline = pipeline.clone();
        let global_shutdown = cancel.global();
        let decode_for_consumer = decode_ctx.clone();
        let acceleration_for_consumer = acceleration.clone();
        let decode_policy_for_consumer = decode_policy.clone();
        let motion_gate_for_consumer = motion_gate.clone();
        let consumer = tokio::spawn(async move {
            run_frame_consumer(
                consumer_pipeline,
                session_shutdown_rx,
                global_shutdown,
                decode_for_consumer,
                acceleration_for_consumer,
                decode_policy_for_consumer,
                false,
                motion_gate_for_consumer,
                motion_sensitivity,
            )
            .await;
        });
        let stats = simulate_frame_loop(camera_id, 5.0, cancel, Some(&mut live), &pipeline).await;
        pipeline.close();
        let _ = session_shutdown_tx.send(true);
        let _ = tokio::time::timeout(Duration::from_secs(5), consumer).await;
        stats
    } else {
        let (session_shutdown_tx, session_shutdown_rx) = watch::channel(false);
        let consumer_pipeline = pipeline.clone();
        let global_shutdown = cancel.global();
        let decode_for_consumer = decode_ctx.clone();
        let acceleration_for_consumer = acceleration.clone();
        let decode_policy_for_consumer = decode_policy.clone();
        let motion_gate_for_consumer = motion_gate.clone();
        let consumer = tokio::spawn(async move {
            run_frame_consumer(
                consumer_pipeline,
                session_shutdown_rx,
                global_shutdown,
                decode_for_consumer,
                acceleration_for_consumer,
                decode_policy_for_consumer,
                true,
                motion_gate_for_consumer,
                motion_sensitivity,
            )
            .await;
        });

        let mut session_stats = RtspLoopStats {
            frames_received: 0,
            frames_dropped: 0,
        };

        loop {
            if cancel.is_cancelled() {
                break;
            }

            let (session, video_index) =
                connect_rtsp_demuxed(rtsp_url, cfg.rtsp_connect_timeout, decode_ctx.as_ref())
                    .await?;

            match run_rtsp_demux_loop(
                camera_id,
                session,
                video_index,
                cfg.rtsp_frame_timeout,
                cancel.clone(),
                Some(&mut live),
                &pipeline,
                decode_ctx.as_ref(),
                &mut enqueue_gate,
                metrics.rtsp_hotpath.as_ref(),
            )
            .await?
            {
                RtspDemuxOutcome::Finished(s) => {
                    session_stats.frames_received += s.frames_received;
                    session_stats.frames_dropped += s.frames_dropped;
                    break;
                }
                RtspDemuxOutcome::IdleSuspend(s) => {
                    session_stats.frames_received += s.frames_received;
                    session_stats.frames_dropped += s.frames_dropped;
                    if cancel.is_cancelled() {
                        break;
                    }
                    let sleep_for = enqueue_gate.rtsp_suspend_sleep();
                    debug!(
                        camera_id,
                        sleep_ms = sleep_for.as_millis(),
                        "rtsp idle suspend until next motion probe"
                    );
                    tokio::select! {
                        _ = tokio::time::sleep(sleep_for) => {}
                        _ = cancel.wait_until_cancelled() => break,
                    }
                }
            }
        }

        pipeline.close();
        let _ = session_shutdown_tx.send(true);
        let _ = tokio::time::timeout(Duration::from_secs(5), consumer).await;
        session_stats
    };

    Ok(stats)
}

/// `true` = encerrar o worker (shutdown/stop da câmera).
async fn wait_reconnect_delay(
    delay: Duration,
    shutdown_rx: &mut watch::Receiver<bool>,
    camera_stop_rx: watch::Receiver<bool>,
) -> bool {
    let mut local_stop = camera_stop_rx;
    tokio::select! {
        _ = tokio::time::sleep(delay) => false,
        _ = shutdown_rx.changed() => true,
        _ = local_stop.changed() => true,
    }
}

async fn mirror_policy_to_state(state: &SharedCameraState, pol: &StreamPolicyState) {
    update_state(state, |s| {
        s.stream_failures_consecutive = pol.failures_consecutive;
        s.stream_hourly_attempts = pol.hourly_attempts;
        s.stream_next_probe_at = pol.next_probe_at;
        s.stream_local_paused = pol.local_paused;
    })
    .await;
}

pub async fn run_camera_worker(
    camera_id: i64,
    rtsp_url: String,
    cfg: Config,
    metrics: Arc<ProcessorMetrics>,
    acceleration: Arc<AccelerationRuntime>,
    decode_policy: Arc<DecodePolicyCoordinator>,
    state: SharedCameraState,
    stream_policy: Arc<RwLock<StreamPolicyState>>,
    camera_created_at: Option<DateTime<Utc>>,
    health_queue: Arc<RwLock<Vec<crate::api::CameraStreamHealthReport>>>,
    shutdown_rx: &mut watch::Receiver<bool>,
    camera_stop_rx: watch::Receiver<bool>,
    global_frames: Arc<AtomicU64>,
) {
    let redacted = redact_rtsp_url(&rtsp_url);
    let mut backoff = ReconnectBackoff::new(cfg.rtsp_reconnect_base);
    let stream_cfg = cfg.stream_retry;

    let simulate = std::env::var("RTSP_SIMULATE")
        .map(|v| matches!(v.as_str(), "1" | "true" | "yes"))
        .unwrap_or(false);

    loop {
        if *shutdown_rx.borrow() || *camera_stop_rx.borrow() {
            break;
        }

        {
            let pol = stream_policy.read().await;
            if pol.local_paused {
                update_state(&state, |s| {
                    s.status = CameraStatus::Offline;
                    s.stream_local_paused = true;
                })
                .await;
                break;
            }
            if let Some(wait) = pol.wait_before_probe() {
                drop(pol);
                update_state(&state, |s| s.status = CameraStatus::Offline).await;
                if wait_reconnect_delay(wait, shutdown_rx, camera_stop_rx.clone()).await {
                    break;
                }
                continue;
            }
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
                {
                    let mut pol = stream_policy.write().await;
                    let action = pol.on_success(&stream_cfg);
                    mirror_policy_to_state(&state, &pol).await;
                    enqueue_stream_action(&health_queue, camera_id, action, None).await;
                }
                info!(camera_id, processor_id = %cfg.processor_id, "rtsp session ended — reconnecting");
                if wait_reconnect_delay(
                    backoff.base_interval(),
                    shutdown_rx,
                    camera_stop_rx.clone(),
                )
                .await
                {
                    break;
                }
            }
            Err(e) => {
                if *camera_stop_rx.borrow() {
                    break;
                }
                metrics.record_reconnect();
                metrics.record_rtsp_error();
                let err_text = e.to_string();
                let failure_class = classify_rtsp_error(&err_text);
                let delay = if stream_cfg.enabled && failure_class != StreamFailureClass::Transient {
                    let mut pol = stream_policy.write().await;
                    let action = pol.on_failure(&stream_cfg, failure_class, camera_created_at);
                    mirror_policy_to_state(&state, &pol).await;
                    if failure_class == StreamFailureClass::PathAbsent {
                        enqueue_stream_failure(
                            &health_queue,
                            camera_id,
                            pol.failures_consecutive,
                            pol.hourly_attempts,
                            Some(err_text.clone()),
                        )
                        .await;
                    }
                    match action {
                        StreamHealthAction::PauseAnalytic { reason } => {
                            enqueue_stream_action(
                                &health_queue,
                                camera_id,
                                StreamHealthAction::PauseAnalytic { reason },
                                Some(err_text.clone()),
                            )
                            .await;
                            Duration::ZERO
                        }
                        StreamHealthAction::RetryAfter(d) => d,
                        StreamHealthAction::ReportStreamOk => backoff.next_delay(),
                        StreamHealthAction::ReportFailure { .. } => delay_after_failure(
                            pol.failures_consecutive,
                            &stream_cfg,
                        ),
                    }
                } else {
                    let d = backoff.next_delay();
                    update_state(&state, |s| {
                        s.reconnect_count += 1;
                        s.rtsp_errors += 1;
                    })
                    .await;
                    d
                };
                warn!(
                    camera_id,
                    processor_id = %cfg.processor_id,
                    error = %err_text,
                    delay_secs = delay.as_secs(),
                    "rtsp reconnect"
                );
                update_state(&state, |s| {
                    s.status = CameraStatus::Offline;
                    s.fps = 0.0;
                    s.last_error = Some(err_text);
                    s.reconnect_count += 1;
                    s.rtsp_errors += 1;
                })
                .await;

                if delay.is_zero() {
                    break;
                }
                if wait_reconnect_delay(delay, shutdown_rx, camera_stop_rx.clone()).await {
                    break;
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

#[cfg(test)]
mod tests {
    use super::*;
    use std::time::Instant;

    #[tokio::test]
    async fn reconnect_delay_waits_at_least_interval() {
        let (_tx, mut shutdown_rx) = watch::channel(false);
        let (_stop_tx, stop_rx) = watch::channel(false);
        let start = Instant::now();
        let stop =
            wait_reconnect_delay(Duration::from_millis(120), &mut shutdown_rx, stop_rx).await;
        assert!(!stop);
        assert!(start.elapsed() >= Duration::from_millis(100));
    }

    #[tokio::test]
    async fn reconnect_delay_exits_on_camera_stop() {
        let (_tx, mut shutdown_rx) = watch::channel(false);
        let (stop_tx, stop_rx) = watch::channel(false);
        stop_tx.send(true).unwrap();
        let stop = wait_reconnect_delay(Duration::from_secs(60), &mut shutdown_rx, stop_rx).await;
        assert!(stop);
    }
}
