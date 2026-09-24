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

/// Registra um frame de vídeo recebido (e opcionalmente descartado pelo buffer).
/// Deve ser chamado uma vez por VideoFrame efetivo no demux.
pub async fn record_video_frame(ctx: &mut LiveCaptureContext<'_>, dropped: bool) {
    let dropped_n = u64::from(dropped);
    ctx.metrics.add_frames(1, dropped_n);
    ctx.global_frames.fetch_add(1, Ordering::Relaxed);

    let fps_update = ctx.fps_est.record_frame();

    let mut map = ctx.states.write().await;
    if let Some(s) = map.get_mut(&ctx.camera_id) {
        s.frames_received += 1;
        s.frames_dropped += dropped_n;
        s.last_frame_at = Some(Utc::now());
        s.status = CameraStatus::Online;
        if let Some(fps) = fps_update {
            s.fps = fps;
        }
    }
}
