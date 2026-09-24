use std::sync::Arc;

use axum::extract::State;
use axum::http::StatusCode;
use axum::Json;
use serde::Serialize;
use tokio::sync::RwLock;

use crate::camera::{CameraRuntimeState, CameraStatus};
use crate::decode::AccelerationRuntime;
use crate::metrics::{MetricsSnapshot, SharedMetrics};

#[derive(Clone)]
pub struct AppState {
    pub metrics: SharedMetrics,
    pub acceleration: Arc<AccelerationRuntime>,
    pub camera_states: Arc<RwLock<std::collections::HashMap<i64, CameraRuntimeState>>>,
    pub api_ready: Arc<std::sync::atomic::AtomicBool>,
}

#[derive(Serialize)]
pub struct HealthResponse {
    pub status: &'static str,
    pub processor_id: String,
    pub uptime_secs: u64,
    pub cameras_total: usize,
    pub cameras_online: usize,
    pub cameras_offline: usize,
    pub frames_received: u64,
    pub fps_total: f64,
    pub reconnects: u64,
    pub errors: u64,
}

#[derive(Serialize)]
pub struct ReadyResponse {
    pub ready: bool,
    pub confvision_api_configured: bool,
}

pub async fn health_handler(State(st): State<AppState>) -> Json<HealthResponse> {
    let snap = st.metrics.snapshot();
    let (total, online, offline, fps_total, _) = summarize_cameras(&st.camera_states).await;
    Json(HealthResponse {
        status: "ok",
        processor_id: snap.processor_id,
        uptime_secs: snap.uptime_secs,
        cameras_total: total,
        cameras_online: online,
        cameras_offline: offline,
        frames_received: snap.frames_received,
        fps_total,
        reconnects: snap.reconnects,
        errors: snap.errors + snap.rtsp_errors,
    })
}

pub async fn ready_handler(State(st): State<AppState>) -> (StatusCode, Json<ReadyResponse>) {
    let ready = st.api_ready.load(std::sync::atomic::Ordering::Relaxed);
    let code = if ready {
        StatusCode::OK
    } else {
        StatusCode::SERVICE_UNAVAILABLE
    };
    (
        code,
        Json(ReadyResponse {
            ready,
            confvision_api_configured: ready,
        }),
    )
}

pub async fn metrics_handler(State(st): State<AppState>) -> Json<MetricsBody> {
    let snap = st
        .metrics
        .snapshot_with_acceleration(Some(st.acceleration.as_ref()));
    let frame_latency_ms = snap.frame_latency_ms;
    let (total, online, offline, fps_total, queue_depth) =
        summarize_cameras(&st.camera_states).await;
    Json(MetricsBody {
        metrics: snap,
        cameras_total: total,
        cameras_online: online,
        cameras_offline: offline,
        fps_total,
        queue_depth,
        processing_latency_ms: frame_latency_ms,
        note: "CPU/RAM/GPU — PRECISA SER MEDIDO no host",
    })
}

#[derive(Serialize)]
pub struct MetricsBody {
    pub metrics: MetricsSnapshot,
    pub cameras_total: usize,
    pub cameras_online: usize,
    pub cameras_offline: usize,
    pub fps_total: f64,
    pub queue_depth: u64,
    pub processing_latency_ms: u64,
    pub note: &'static str,
}

async fn summarize_cameras(
    states: &Arc<RwLock<std::collections::HashMap<i64, CameraRuntimeState>>>,
) -> (usize, usize, usize, f64, u64) {
    let map = states.read().await;
    let total = map.len();
    let mut online = 0usize;
    let mut offline = 0usize;
    let mut fps_total = 0.0f64;
    let mut queue_depth = 0u64;
    for s in map.values() {
        fps_total += s.fps;
        queue_depth += s.buffer_size;
        match s.status {
            CameraStatus::Online => online += 1,
            CameraStatus::Offline | CameraStatus::Error | CameraStatus::Reconnecting => {
                offline += 1
            }
            _ => {}
        }
    }
    (total, online, offline, fps_total, queue_depth)
}
