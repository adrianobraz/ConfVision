use std::sync::Arc;

use tokio::sync::{watch, RwLock};

use crate::api::CameraStreamHealthReport;
use crate::stream_policy::{StreamHealthAction, StreamPolicyState, STREAM_PAUSE_REASON_H264_NAL};
use crate::worker::stream_health_queue::{enqueue_stream_action, enqueue_stream_failure};

/// Contexto para pausa imediata por H.264/NAL corrupto (mesmo fluxo que RTSP 404).
#[derive(Clone)]
pub struct H264NalPauseContext {
    pub camera_id: i64,
    health_queue: Arc<RwLock<Vec<CameraStreamHealthReport>>>,
    stream_policy: Arc<RwLock<StreamPolicyState>>,
    session_shutdown: watch::Sender<bool>,
}

impl H264NalPauseContext {
    pub fn new(
        camera_id: i64,
        health_queue: Arc<RwLock<Vec<CameraStreamHealthReport>>>,
        stream_policy: Arc<RwLock<StreamPolicyState>>,
        session_shutdown: watch::Sender<bool>,
    ) -> Self {
        Self {
            camera_id,
            health_queue,
            stream_policy,
            session_shutdown,
        }
    }

    pub async fn trigger_immediate_pause(&self, detail: String) {
        let (failures, hourly) = {
            let mut pol = self.stream_policy.write().await;
            pol.local_paused = true;
            pol.failures_consecutive = pol.failures_consecutive.saturating_add(1);
            (pol.failures_consecutive, pol.hourly_attempts)
        };

        let msg = detail.trim().to_string();
        let last_error = if msg.is_empty() {
            Some("H.264 inválido (NAL corrupto no stream)".into())
        } else {
            Some(msg.clone())
        };

        enqueue_stream_failure(
            &self.health_queue,
            self.camera_id,
            failures,
            hourly,
            last_error.clone(),
            Some("h264_invalid_nal".into()),
        )
        .await;

        enqueue_stream_action(
            &self.health_queue,
            self.camera_id,
            StreamHealthAction::PauseAnalytic {
                reason: STREAM_PAUSE_REASON_H264_NAL.into(),
            },
            last_error,
            Some("h264_invalid_nal".into()),
        )
        .await;

        let _ = self.session_shutdown.send(true);
    }
}
