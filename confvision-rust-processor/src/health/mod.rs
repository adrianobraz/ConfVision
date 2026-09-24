use std::sync::Arc;

use axum::extract::State;
use axum::http::StatusCode;
use axum::Json;
use serde::Serialize;
use tokio::sync::RwLock;

use crate::camera::{CameraRuntimeState, CameraStatus};
use crate::config::Config;
use crate::decode::AccelerationRuntime;
use crate::metrics::{MetricsSnapshot, SharedMetrics};

#[derive(Clone, Serialize)]
pub struct RuntimeIdentity {
    pub processor_id: String,
    pub worker_id: String,
    pub worker_tipo: String,
    pub shard_mode: String,
    pub worker_shard_index: i32,
    pub worker_shard_total: u32,
    pub mediamtx_node_id: u32,
    pub max_cameras: usize,
}

impl RuntimeIdentity {
    pub fn from_config(cfg: &Config) -> Self {
        Self {
            processor_id: cfg.processor_id.clone(),
            worker_id: cfg.worker_id.clone(),
            worker_tipo: cfg.worker_tipo.clone(),
            shard_mode: cfg.shard_mode.as_str().to_string(),
            worker_shard_index: cfg.worker_shard_index,
            worker_shard_total: cfg.worker_shard_total,
            mediamtx_node_id: cfg.mediamtx_node_id,
            max_cameras: cfg.max_cameras,
        }
    }
}

#[derive(Clone)]
pub struct AppState {
    pub metrics: SharedMetrics,
    pub acceleration: Arc<AccelerationRuntime>,
    pub camera_states:
        Arc<RwLock<std::collections::HashMap<i64, crate::camera::SharedCameraState>>>,
    pub api_ready: Arc<std::sync::atomic::AtomicBool>,
    pub identity: RuntimeIdentity,
}

#[derive(Serialize)]
pub struct HealthResponse {
    pub status: &'static str,
    pub processor_id: String,
    pub worker_id: String,
    pub worker_tipo: String,
    pub shard_mode: String,
    pub worker_shard_index: i32,
    pub worker_shard_total: u32,
    pub mediamtx_node_id: u32,
    pub max_cameras: usize,
    pub uptime_secs: u64,
    pub cameras_total: usize,
    pub cameras_online: usize,
    pub cameras_offline: usize,
    pub cameras_reconnecting: usize,
    pub cameras_starting: usize,
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
    let summary = summarize_cameras(&st.camera_states).await;
    Json(HealthResponse {
        status: "ok",
        processor_id: st.identity.processor_id.clone(),
        worker_id: st.identity.worker_id.clone(),
        worker_tipo: st.identity.worker_tipo.clone(),
        shard_mode: st.identity.shard_mode.clone(),
        worker_shard_index: st.identity.worker_shard_index,
        worker_shard_total: st.identity.worker_shard_total,
        mediamtx_node_id: st.identity.mediamtx_node_id,
        max_cameras: st.identity.max_cameras,
        uptime_secs: snap.uptime_secs,
        cameras_total: summary.total,
        cameras_online: summary.online,
        cameras_offline: summary.offline,
        cameras_reconnecting: summary.reconnecting,
        cameras_starting: summary.starting,
        frames_received: snap.frames_received,
        fps_total: summary.fps_total,
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
    let summary = summarize_cameras(&st.camera_states).await;
    let cameras = snapshot_camera_states(&st.camera_states).await;
    Json(MetricsBody {
        identity: st.identity.clone(),
        metrics: snap,
        cameras_total: summary.total,
        cameras_online: summary.online,
        cameras_offline: summary.offline,
        cameras_reconnecting: summary.reconnecting,
        fps_total: summary.fps_total,
        queue_depth: summary.queue_depth,
        processing_latency_ms: frame_latency_ms,
        cameras,
        note: "CPU/RAM/GPU — PRECISA SER MEDIDO no host",
    })
}

#[derive(Serialize, Clone)]
pub struct MetricsBody {
    pub identity: RuntimeIdentity,
    pub metrics: MetricsSnapshot,
    pub cameras_total: usize,
    pub cameras_online: usize,
    pub cameras_offline: usize,
    pub cameras_reconnecting: usize,
    pub fps_total: f64,
    pub queue_depth: u64,
    pub processing_latency_ms: u64,
    /// Métricas por câmera (mesmos campos que `CameraRuntimeState`).
    pub cameras: Vec<CameraRuntimeState>,
    pub note: &'static str,
}

struct CameraSummary {
    total: usize,
    online: usize,
    offline: usize,
    reconnecting: usize,
    starting: usize,
    fps_total: f64,
    queue_depth: u64,
}

async fn summarize_cameras(
    states: &Arc<RwLock<std::collections::HashMap<i64, crate::camera::SharedCameraState>>>,
) -> CameraSummary {
    let map = states.read().await;
    let total = map.len();
    let mut online = 0usize;
    let mut offline = 0usize;
    let mut reconnecting = 0usize;
    let mut starting = 0usize;
    let mut fps_total = 0.0f64;
    let mut queue_depth = 0u64;
    for cell in map.values() {
        let s = cell.read().await;
        fps_total += s.fps;
        queue_depth += s.buffer_size;
        match s.status {
            CameraStatus::Online => online += 1,
            CameraStatus::Offline | CameraStatus::Error => offline += 1,
            CameraStatus::Reconnecting => reconnecting += 1,
            CameraStatus::Starting => starting += 1,
            CameraStatus::Stopped => {}
        }
    }
    CameraSummary {
        total,
        online,
        offline,
        reconnecting,
        starting,
        fps_total,
        queue_depth,
    }
}

async fn snapshot_camera_states(
    states: &Arc<RwLock<std::collections::HashMap<i64, crate::camera::SharedCameraState>>>,
) -> Vec<CameraRuntimeState> {
    let map = states.read().await;
    let mut out = Vec::with_capacity(map.len());
    for cell in map.values() {
        out.push(cell.read().await.clone());
    }
    out.sort_by_key(|c| c.camera_id);
    out
}
