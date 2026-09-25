use std::sync::atomic::{AtomicU64, Ordering};

use chrono::Utc;

use crate::camera::{CameraStatus, FpsEstimator, SharedCameraState};
use crate::metrics::ProcessorMetrics;

/// Referências para atualização incremental durante captura RTSP/simulada (por câmera).
pub struct LiveCaptureContext<'a> {
    pub metrics: &'a ProcessorMetrics,
    pub global_frames: &'a AtomicU64,
    pub state: &'a SharedCameraState,
    pub fps_est: &'a mut FpsEstimator,
}

/// RTSP recebeu um VideoFrame (contagem live — não inclui drops da pipeline).
pub fn record_frame_received(ctx: &mut LiveCaptureContext<'_>) {
    ctx.metrics.add_frames_received(1);
    ctx.global_frames.fetch_add(1, Ordering::Relaxed);

    let fps_update = ctx.fps_est.record_frame();

    if let Ok(mut s) = ctx.state.try_write() {
        s.frames_received += 1;
        s.last_frame_at = Some(Utc::now());
        s.status = CameraStatus::Online;
        if let Some(fps) = fps_update {
            s.fps = fps;
        }
    }
}
