use std::sync::atomic::{AtomicU64, Ordering};

use chrono::Utc;

use crate::camera::{CameraStatus, FpsEstimator, SharedCameraState};
use crate::metrics::ProcessorMetrics;

/// Sincroniza contadores locais → estado da câmera a cada N AUs throttled.
const THROTTLED_STATE_SYNC_EVERY: u32 = 16;

/// Referências para atualização incremental durante captura RTSP/simulada (por câmera).
pub struct LiveCaptureContext<'a> {
    pub metrics: &'a ProcessorMetrics,
    pub global_frames: &'a AtomicU64,
    pub state: &'a SharedCameraState,
    pub fps_est: &'a mut FpsEstimator,
    /// AUs recebidos no RTSP ainda não refletidos em `CameraRuntimeState`.
    pub pending_received: u32,
    /// AUs throttled acumulados para `FpsEstimator::record_frames`.
    pub pending_fps_frames: u32,
}

fn bump_global_receive_counters(ctx: &mut LiveCaptureContext<'_>) {
    ctx.metrics.add_frames_received(1);
    ctx.global_frames.fetch_add(1, Ordering::Relaxed);
    ctx.pending_received = ctx.pending_received.saturating_add(1);
}

fn sync_camera_state_if_due(ctx: &mut LiveCaptureContext<'_>, force: bool, fps_batch: u32) {
    let fps_update = if fps_batch > 0 {
        ctx.fps_est.record_frames(fps_batch)
    } else {
        None
    };
    if !force && fps_update.is_none() && ctx.pending_received < THROTTLED_STATE_SYNC_EVERY {
        return;
    }
    if let Ok(mut s) = ctx.state.try_write() {
        if ctx.pending_received > 0 {
            s.frames_received += ctx.pending_received as u64;
            ctx.pending_received = 0;
        }
        s.last_frame_at = Some(Utc::now());
        s.status = CameraStatus::Online;
        if let Some(fps) = fps_update {
            s.fps = fps;
        }
    }
}

/// AU recebido no RTSP mas descartado antes do enqueue (hot path leve).
pub fn record_rtsp_au_throttled(ctx: &mut LiveCaptureContext<'_>) {
    ctx.metrics.add_frames_received(1);
    ctx.global_frames.fetch_add(1, Ordering::Relaxed);
    ctx.pending_received = ctx.pending_received.saturating_add(1);
    ctx.pending_fps_frames = ctx.pending_fps_frames.saturating_add(1);
    if ctx.pending_received >= THROTTLED_STATE_SYNC_EVERY {
        let batch = ctx.pending_fps_frames;
        ctx.pending_fps_frames = 0;
        sync_camera_state_if_due(ctx, false, batch);
    }
}

/// AU enfileirado para decode — atualiza estado da câmera imediatamente.
pub fn record_frame_enqueued(ctx: &mut LiveCaptureContext<'_>) {
    bump_global_receive_counters(ctx);
    let batch = ctx.pending_fps_frames.saturating_add(1);
    ctx.pending_fps_frames = 0;
    sync_camera_state_if_due(ctx, true, batch);
}

/// RTSP recebeu um VideoFrame (simulação / caminho sem gate).
pub fn record_frame_received(ctx: &mut LiveCaptureContext<'_>) {
    record_frame_enqueued(ctx);
}

pub fn flush_pending_camera_state(ctx: &mut LiveCaptureContext<'_>) {
    if ctx.pending_received == 0 && ctx.pending_fps_frames == 0 {
        return;
    }
    let batch = ctx.pending_fps_frames;
    ctx.pending_fps_frames = 0;
    if batch > 0 {
        let _ = ctx.fps_est.record_frames(batch);
    }
    if ctx.pending_received == 0 {
        return;
    }
    if let Ok(mut s) = ctx.state.try_write() {
        s.frames_received += ctx.pending_received as u64;
        ctx.pending_received = 0;
        s.last_frame_at = Some(Utc::now());
        s.status = CameraStatus::Online;
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::sync::atomic::AtomicU64;
    use std::sync::Arc;

    use crate::camera::{CameraRuntimeState, FpsEstimator};

    #[tokio::test]
    async fn throttled_path_batches_fps_estimator() {
        let metrics = Arc::new(ProcessorMetrics::new("t"));
        let global = AtomicU64::new(0);
        let state = Arc::new(tokio::sync::RwLock::new(CameraRuntimeState::new(
            1,
            "x".into(),
        )));
        let mut fps_est = FpsEstimator::new();
        let mut ctx = LiveCaptureContext {
            metrics: metrics.as_ref(),
            global_frames: &global,
            state: &state,
            fps_est: &mut fps_est,
            pending_received: 0,
            pending_fps_frames: 0,
        };
        for _ in 0..THROTTLED_STATE_SYNC_EVERY {
            record_rtsp_au_throttled(&mut ctx);
        }
        assert_eq!(ctx.pending_fps_frames, 0);
        assert_eq!(
            global.load(Ordering::Relaxed),
            THROTTLED_STATE_SYNC_EVERY as u64
        );
    }
}
