use std::sync::Arc;

use axum::extract::State;
use axum::http::StatusCode;
use axum::Json;
use serde::Serialize;
use tokio::sync::RwLock;

use crate::camera::{CameraRuntimeState, CameraStatus};
use crate::capacity::{CapacityEngine, CapacitySnapshot, CapacityState, LimitingResource};
use crate::config::Config;
use crate::decode::{AccelerationRuntime, DecodePolicyCoordinator};
use crate::load::LoadAdmissionGate;
use crate::metrics::{MetricsSnapshot, SharedMetrics};
use crate::rtsp_hotpath::RtspHotpathMetrics;

mod capacity_report;
mod runtime_phase62;
pub use capacity_report::capacity_report_handler;
pub use runtime_phase62::{build_phase62_view, Phase62RuntimeView};

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
    pub capacity: Arc<CapacityEngine>,
    pub decode_policy: Arc<DecodePolicyCoordinator>,
    pub load_admission: Arc<LoadAdmissionGate>,
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
    pub capacity_state: CapacityState,
    pub capacity_used_percent: Option<f64>,
    pub estimated_capacity_cameras: Option<u32>,
    pub estimated_available_cameras: Option<i32>,
    pub limiting_resource: LimitingResource,
    pub gpu_decode_state: crate::decode::GpuDecodeState,
    pub hw_stack_grade: crate::decode::HwStackGrade,
    pub decode_backend_effective: String,
    pub load_advisory: crate::load::LoadAdvisory,
    pub load_advisory_reason: String,
}

#[derive(Serialize)]
pub struct ReadyResponse {
    pub ready: bool,
    pub confvision_api_configured: bool,
}

pub async fn health_handler(State(st): State<AppState>) -> Json<HealthResponse> {
    let snap = st.metrics.snapshot();
    let summary = summarize_cameras(&st.camera_states).await;
    let cap = st.capacity.snapshot().await;
    let phase62 = build_phase62_view(
        st.acceleration.as_ref(),
        st.decode_policy.as_ref(),
        &cap,
        st.load_admission.config(),
    );
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
        capacity_state: cap.state,
        capacity_used_percent: cap.capacity_used_percent,
        estimated_capacity_cameras: cap.estimated_capacity_cameras,
        estimated_available_cameras: cap.estimated_available_cameras,
        limiting_resource: cap.limiting_resource,
        gpu_decode_state: phase62.gpu_decode_state,
        hw_stack_grade: phase62.hw_stack_grade,
        decode_backend_effective: phase62.decode_backend_effective,
        load_advisory: phase62.load_advisory,
        load_advisory_reason: phase62.load_advisory_reason,
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
    let capacity = st.capacity.snapshot().await;
    let runtime_phase62 = build_phase62_view(
        st.acceleration.as_ref(),
        st.decode_policy.as_ref(),
        &capacity,
        st.load_admission.config(),
    );
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
        capacity,
        gpu_decode_state: runtime_phase62.gpu_decode_state,
        decode_backend_effective: runtime_phase62.decode_backend_effective,
        hw_stack_grade: runtime_phase62.hw_stack_grade,
        load_advisory: runtime_phase62.load_advisory,
        load_advisory_reason: runtime_phase62.load_advisory_reason,
        rtsp_hotpath: st.metrics.rtsp_hotpath.snapshot(),
        note: "capacity.mode=dynamic: MAX_CAMERAS é apenas hard safety limit",
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
    pub capacity: CapacitySnapshot,
    pub gpu_decode_state: crate::decode::GpuDecodeState,
    pub decode_backend_effective: String,
    pub hw_stack_grade: crate::decode::HwStackGrade,
    pub load_advisory: crate::load::LoadAdvisory,
    pub load_advisory_reason: String,
    pub rtsp_hotpath: RtspHotpathMetrics,
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

#[cfg(test)]
mod integration_tests {
    use super::*;
    use crate::capacity::CapacityEngine;
    use crate::config::{CapacityMode, Config, ShardMode};
    use crate::decode::{
        AccelerationPolicy, AccelerationRuntime, VideoAccelerationMode, VideoGpuBackend,
    };
    use crate::metrics::ProcessorMetrics;
    use std::sync::Arc;
    use std::time::Duration;

    fn test_config() -> Config {
        Config::test_stub()
    }

    #[tokio::test]
    async fn metrics_json_includes_capacity_section() {
        let cfg = test_config();
        let metrics = Arc::new(ProcessorMetrics::new("proc"));
        let acceleration = AccelerationRuntime::bootstrap(AccelerationPolicy::from_parts(
            VideoAccelerationMode::Cpu,
            VideoGpuBackend::Auto,
        ))
        .unwrap();
        let capacity = CapacityEngine::new(&cfg);
        let body = MetricsBody {
            identity: RuntimeIdentity::from_config(&cfg),
            metrics: metrics.snapshot(),
            cameras_total: 0,
            cameras_online: 0,
            cameras_offline: 0,
            cameras_reconnecting: 0,
            fps_total: 0.0,
            queue_depth: 0,
            processing_latency_ms: 0,
            cameras: vec![],
            capacity: capacity.snapshot().await,
            gpu_decode_state: crate::decode::GpuDecodeState::Unavailable,
            decode_backend_effective: "cpu".into(),
            hw_stack_grade: crate::decode::HwStackGrade::None,
            load_advisory: crate::load::LoadAdvisory::Normal,
            load_advisory_reason: "test".into(),
            rtsp_hotpath: metrics.rtsp_hotpath.snapshot(),
            note: "test",
        };
        let v = serde_json::to_value(&body).unwrap();
        assert!(v.get("capacity").is_some());
        assert_eq!(v["capacity"]["mode"], "dynamic");
    }

    #[tokio::test]
    async fn health_json_includes_capacity_summary_fields() {
        let cfg = test_config();
        let cap = CapacityEngine::new(&cfg);
        let snap = cap.snapshot().await;
        let health = HealthResponse {
            status: "ok",
            processor_id: cfg.processor_id.clone(),
            worker_id: cfg.worker_id.clone(),
            worker_tipo: cfg.worker_tipo.clone(),
            shard_mode: cfg.shard_mode.as_str().to_string(),
            worker_shard_index: cfg.worker_shard_index,
            worker_shard_total: cfg.worker_shard_total,
            mediamtx_node_id: cfg.mediamtx_node_id,
            max_cameras: cfg.max_cameras,
            uptime_secs: 0,
            cameras_total: 0,
            cameras_online: 0,
            cameras_offline: 0,
            cameras_reconnecting: 0,
            cameras_starting: 0,
            frames_received: 0,
            fps_total: 0.0,
            reconnects: 0,
            errors: 0,
            capacity_state: snap.state,
            capacity_used_percent: snap.capacity_used_percent,
            estimated_capacity_cameras: snap.estimated_capacity_cameras,
            estimated_available_cameras: snap.estimated_available_cameras,
            limiting_resource: snap.limiting_resource,
            gpu_decode_state: crate::decode::GpuDecodeState::Unavailable,
            hw_stack_grade: crate::decode::HwStackGrade::None,
            decode_backend_effective: "cpu".into(),
            load_advisory: crate::load::LoadAdvisory::Normal,
            load_advisory_reason: "test".into(),
        };
        let v = serde_json::to_value(&health).unwrap();
        assert!(v.get("capacity_state").is_some());
        assert!(v.get("limiting_resource").is_some());
    }
}
