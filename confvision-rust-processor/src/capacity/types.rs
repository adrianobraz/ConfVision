use serde::Serialize;

use crate::config::CapacityMode;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum ResourceScope {
    Host,
    Container,
    Process,
    #[default]
    Unavailable,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum CapacityState {
    Healthy,
    Warning,
    Critical,
    Unknown,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum LimitingResource {
    Cpu,
    Memory,
    Gpu,
    Vram,
    SafetyLimit,
    None,
}

#[derive(Debug, Clone, Serialize)]
pub struct ResourceMetric {
    pub percent: Option<f64>,
    pub target_percent: f64,
    pub headroom_percent: Option<f64>,
    pub scope: ResourceScope,
}

#[derive(Debug, Clone, Serialize)]
pub struct NetworkSnapshot {
    pub rx_bytes_total: Option<u64>,
    pub tx_bytes_total: Option<u64>,
    pub rx_bps: Option<f64>,
    pub tx_bps: Option<f64>,
    pub scope: ResourceScope,
}

#[derive(Debug, Clone, Serialize)]
pub struct StorageSnapshot {
    pub disk_total_bytes: Option<u64>,
    pub disk_used_bytes: Option<u64>,
    pub disk_free_bytes: Option<u64>,
    pub disk_used_percent: Option<f64>,
    pub scope: ResourceScope,
}

#[derive(Debug, Clone, Serialize)]
pub struct ResourceSnapshot {
    pub cpu_usage_percent: Option<f64>,
    pub memory_usage_percent: Option<f64>,
    pub load_1m: Option<f64>,
    pub gpu_usage_percent: Option<f64>,
    pub vram_usage_percent: Option<f64>,
    pub cpu: ResourceMetric,
    pub memory: ResourceMetric,
    pub gpu: ResourceMetric,
    pub vram: ResourceMetric,
    pub network: NetworkSnapshot,
    pub storage: StorageSnapshot,
    pub uptime_secs: u64,
}

#[derive(Debug, Clone, Serialize)]
pub struct CameraCapacityView {
    pub camera_id: i64,
    pub status: String,
    pub fps: f64,
    pub camera_fps_share: Option<f64>,
    pub camera_drop_rate: Option<f64>,
    pub frames_received: u64,
    pub frames_processed: u64,
    pub frames_dropped: u64,
    pub frames_decoded: u64,
    pub processing_latency_ms: u64,
}

#[derive(Debug, Clone, Serialize)]
pub struct CapacitySnapshot {
    pub mode: CapacityMode,
    pub state: CapacityState,
    pub worker_id: String,
    pub processor_id: String,
    pub cpu: ResourceMetric,
    pub memory: ResourceMetric,
    pub gpu: ResourceMetric,
    pub vram: ResourceMetric,
    pub network: NetworkSnapshot,
    pub storage: StorageSnapshot,
    pub current_cameras: usize,
    pub current_online_cameras: usize,
    pub current_fps: f64,
    pub average_fps_per_camera: Option<f64>,
    pub current_frames_received: u64,
    pub current_frames_processed: u64,
    pub current_frames_dropped: u64,
    pub drop_rate_percent: Option<f64>,
    pub queue_depth: u64,
    pub processing_latency_ms: u64,
    pub estimated_capacity_cameras: Option<u32>,
    pub estimated_capacity_fps: Option<f64>,
    pub estimated_available_cameras: Option<i32>,
    pub estimated_available_fps: Option<f64>,
    pub capacity_used_percent: Option<f64>,
    pub limiting_resource: LimitingResource,
    pub max_cameras_safety_limit: Option<usize>,
    pub cameras: Vec<CameraCapacityView>,
    pub observation_ready: bool,
    pub samples_in_window: usize,
}

#[derive(Debug, Clone, Serialize)]
pub struct CapacityEstimate {
    pub worker_id: String,
    pub processor_id: String,
    pub state: CapacityState,
    pub current_load: CapacityLoadSummary,
    pub capacity: CapacityHeadroomSummary,
    pub available_capacity: CapacityAvailableSummary,
    pub limiting_resource: LimitingResource,
}

#[derive(Debug, Clone, Serialize)]
pub struct CapacityLoadSummary {
    pub online_cameras: usize,
    pub fps_total: f64,
    pub capacity_used_percent: Option<f64>,
}

#[derive(Debug, Clone, Serialize)]
pub struct CapacityHeadroomSummary {
    pub estimated_capacity_cameras: Option<u32>,
    pub estimated_capacity_fps: Option<f64>,
}

#[derive(Debug, Clone, Serialize)]
pub struct CapacityAvailableSummary {
    pub estimated_available_cameras: Option<i32>,
    pub estimated_available_fps: Option<f64>,
}

#[derive(Debug, Clone)]
pub struct HistorySample {
    pub cpu_percent: Option<f64>,
    pub memory_percent: Option<f64>,
    pub gpu_percent: Option<f64>,
    pub vram_percent: Option<f64>,
    pub fps: f64,
    pub online_cameras: usize,
    pub total_cameras: usize,
    pub queue_depth: u64,
    pub frames_dropped_delta: u64,
    pub frames_received_delta: u64,
    pub processing_latency_ms: u64,
}
