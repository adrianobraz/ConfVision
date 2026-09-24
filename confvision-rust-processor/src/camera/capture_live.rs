use std::collections::HashMap;
use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::Arc;

use chrono::Utc;
use tokio::sync::RwLock;

use crate::camera::{CameraRuntimeState, CameraStatus, FpsEstimator};
use crate::metrics::ProcessorMetrics;

/// Referências compartilhadas para atualização incremental durante captura RTSP/simulada.
pub struct LiveCaptureContext<'a> {
    pub metrics: &'a ProcessorMetrics,
    pub global_frames: &'a AtomicU64,
    pub states: &'a Arc<RwLock<HashMap<i64, CameraRuntimeState>>>,
    pub camera_id: i64,
    pub fps_est: &'a mut FpsEstimator,
}

/// RTSP recebeu um VideoFrame (contagem live — não inclui drops da pipeline).
pub async fn record_frame_received(ctx: &mut LiveCaptureContext<'_>) {
    ctx.metrics.add_frames_received(1);
    ctx.global_frames.fetch_add(1, Ordering::Relaxed);

    let fps_update = ctx.fps_est.record_frame();

    let mut map = ctx.states.write().await;
    if let Some(s) = map.get_mut(&ctx.camera_id) {
        s.frames_received += 1;
        s.last_frame_at = Some(Utc::now());
        s.status = CameraStatus::Online;
        if let Some(fps) = fps_update {
            s.fps = fps;
        }
    }
}
