mod collector;
mod estimator;
mod policy;
mod types;

pub use crate::config::CapacityMode;
pub use policy::CapacityPolicy;
pub use types::{
    CapacityEstimate, CapacitySnapshot, CapacityState, LimitingResource, NetworkSnapshot,
    ResourceMetric, ResourceScope, StorageSnapshot,
};

use std::collections::VecDeque;
use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::Arc;
use std::time::Instant;

use tokio::sync::RwLock;
use tracing::debug;

use crate::camera::{CameraRuntimeState, CameraStatus, SharedCameraState};
use crate::config::Config;
use crate::decode::AccelerationRuntime;
use crate::metrics::{MetricsSnapshot, ProcessorMetrics};

use collector::ResourceCollector;
use estimator::{average_history, compute_estimate, EstimateInput};
use types::{
    CameraCapacityView, CapacityAvailableSummary, CapacityHeadroomSummary, CapacityLoadSummary,
};

pub struct CapacityEngine {
    policy: CapacityPolicy,
    worker_id: String,
    processor_id: String,
    started_at: Instant,
    collector: RwLock<ResourceCollector>,
    history: RwLock<VecDeque<types::HistorySample>>,
    snapshot: RwLock<CapacitySnapshot>,
    last_frames_received: AtomicU64,
    last_frames_dropped: AtomicU64,
}

impl CapacityEngine {
    pub fn new(cfg: &Config) -> Arc<Self> {
        let policy = CapacityPolicy::from_config(cfg);
        let disabled = policy.mode == CapacityMode::Disabled;
        let initial_state = if disabled {
            CapacityState::Unknown
        } else {
            CapacityState::Unknown
        };
        let snap = empty_snapshot(cfg, initial_state, &policy);
        Self::from_parts(cfg, policy, snap)
    }

    #[cfg(test)]
    pub fn new_with_initial_snapshot(cfg: &Config, snap: CapacitySnapshot) -> Arc<Self> {
        let policy = CapacityPolicy::from_config(cfg);
        Self::from_parts(cfg, policy, snap)
    }

    fn from_parts(cfg: &Config, policy: CapacityPolicy, snap: CapacitySnapshot) -> Arc<Self> {
        Arc::new(Self {
            policy,
            worker_id: cfg.worker_id.clone(),
            processor_id: cfg.processor_id.clone(),
            started_at: Instant::now(),
            collector: RwLock::new(ResourceCollector::new()),
            history: RwLock::new(VecDeque::new()),
            snapshot: RwLock::new(snap),
            last_frames_received: AtomicU64::new(0),
            last_frames_dropped: AtomicU64::new(0),
        })
    }

    pub async fn snapshot(&self) -> CapacitySnapshot {
        self.snapshot.read().await.clone()
    }

    #[cfg(test)]
    pub async fn replace_snapshot_for_test(&self, snap: CapacitySnapshot) {
        *self.snapshot.write().await = snap;
    }

    pub async fn estimate(&self) -> CapacityEstimate {
        self.current().await
    }

    pub async fn current(&self) -> CapacityEstimate {
        let snap = self.snapshot.read().await.clone();
        CapacityEstimate {
            worker_id: snap.worker_id.clone(),
            processor_id: snap.processor_id.clone(),
            state: snap.state,
            current_load: CapacityLoadSummary {
                online_cameras: snap.current_online_cameras,
                fps_total: snap.current_fps,
                capacity_used_percent: snap.capacity_used_percent,
            },
            capacity: CapacityHeadroomSummary {
                estimated_capacity_cameras: snap.estimated_capacity_cameras,
                estimated_capacity_fps: snap.estimated_capacity_fps,
            },
            available_capacity: CapacityAvailableSummary {
                estimated_available_cameras: snap.estimated_available_cameras,
                estimated_available_fps: snap.estimated_available_fps,
            },
            limiting_resource: snap.limiting_resource,
        }
    }

    pub async fn tick(
        &self,
        metrics: &ProcessorMetrics,
        camera_states: &Arc<tokio::sync::RwLock<std::collections::HashMap<i64, SharedCameraState>>>,
        acceleration: &AccelerationRuntime,
    ) {
        if self.policy.mode == CapacityMode::Disabled {
            let mut snap = self.snapshot.write().await;
            snap.mode = CapacityMode::Disabled;
            snap.state = CapacityState::Unknown;
            return;
        }

        let m = metrics.snapshot_with_acceleration(Some(acceleration));
        let (summary, cameras) = gather_cameras(camera_states).await;
        let mut collector = self.collector.write().await;
        let raw = collector.sample();
        let (rx_bps, tx_bps) = collector.network_rates(raw.network_rx_bytes, raw.network_tx_bytes);

        let history_empty = self.history.read().await.is_empty();
        let prev_recv = self.last_frames_received.load(Ordering::Relaxed);
        let prev_drop = self.last_frames_dropped.load(Ordering::Relaxed);
        let recv_delta = if history_empty {
            0
        } else {
            m.frames_received.saturating_sub(prev_recv)
        };
        let drop_delta = if history_empty {
            0
        } else {
            m.frames_dropped.saturating_sub(prev_drop)
        };
        self.last_frames_received
            .store(m.frames_received, Ordering::Relaxed);
        self.last_frames_dropped
            .store(m.frames_dropped, Ordering::Relaxed);

        let sample = types::HistorySample {
            cpu_percent: raw.cpu_usage_percent,
            memory_percent: raw.memory_usage_percent,
            gpu_percent: gpu_load_proxy(acceleration, &m),
            vram_percent: None,
            fps: summary.fps_total,
            online_cameras: summary.online,
            total_cameras: summary.total,
            queue_depth: summary.queue_depth,
            frames_dropped_delta: drop_delta,
            frames_received_delta: recv_delta,
            processing_latency_ms: m.frame_latency_ms,
        };

        {
            let mut hist = self.history.write().await;
            hist.push_back(sample);
            while hist.len() > self.policy.history_size {
                hist.pop_front();
            }
        }

        let hist: Vec<_> = self.history.read().await.iter().cloned().collect();
        let avg = average_history(&hist);
        let observation_ready = self.started_at.elapsed() >= self.policy.min_sample
            && avg.samples > 0
            && (avg.cpu_percent.is_some() || avg.memory_percent.is_some());

        let est = compute_estimate(EstimateInput {
            policy: &self.policy,
            signals: &avg,
            observation_ready,
            cpu_scope: raw.cpu_scope,
            memory_scope: raw.memory_scope,
            gpu_scope: if acceleration.gpu_detected() {
                ResourceScope::Process
            } else {
                ResourceScope::Unavailable
            },
            vram_scope: ResourceScope::Unavailable,
            gpu_detected: acceleration.gpu_detected(),
        });

        let storage = storage_from_raw(&raw);
        let network = NetworkSnapshot {
            rx_bytes_total: raw.network_rx_bytes,
            tx_bytes_total: raw.network_tx_bytes,
            rx_bps,
            tx_bps,
            scope: raw.network_scope,
        };

        let camera_views = build_camera_views(&cameras, summary.fps_total);

        let mut snap = self.snapshot.write().await;
        snap.mode = CapacityMode::Dynamic;
        snap.state = est.state;
        snap.cpu = est.cpu_metric;
        snap.memory = est.memory_metric;
        snap.gpu = est.gpu_metric;
        snap.vram = est.vram_metric;
        snap.network = network;
        snap.storage = storage;
        snap.current_cameras = summary.total;
        snap.current_online_cameras = summary.online;
        snap.current_fps = summary.fps_total;
        snap.average_fps_per_camera = est.avg_fps_per_camera;
        snap.current_frames_received = m.frames_received;
        snap.current_frames_processed = m.frames_processed;
        snap.current_frames_dropped = m.frames_dropped;
        snap.drop_rate_percent = avg.drop_rate_percent;
        snap.queue_depth = summary.queue_depth;
        snap.processing_latency_ms = m.frame_latency_ms;
        snap.estimated_capacity_cameras = est.estimated_capacity_cameras;
        snap.estimated_capacity_fps = est.estimated_capacity_fps;
        snap.estimated_available_cameras = est.estimated_available_cameras;
        snap.estimated_available_fps = est.estimated_available_fps;
        snap.capacity_used_percent = est.capacity_used_percent;
        snap.limiting_resource = est.limiting_resource;
        snap.max_cameras_safety_limit = self.policy.max_cameras_safety;
        snap.cameras = camera_views;
        snap.observation_ready = observation_ready;
        snap.samples_in_window = avg.samples;

        debug!(
            state = ?snap.state,
            limiting = ?snap.limiting_resource,
            est_cameras = ?snap.estimated_capacity_cameras,
            "capacity tick"
        );
    }

    pub fn sample_interval(&self) -> std::time::Duration {
        self.policy.sample_interval
    }
}

fn gpu_load_proxy(acceleration: &AccelerationRuntime, m: &MetricsSnapshot) -> Option<f64> {
    if !acceleration.gpu_detected() {
        return None;
    }
    let hw = m.frames_hw_decoded;
    let cpu = m.frames_cpu_decoded;
    let total = hw + cpu;
    if total == 0 {
        return None;
    }
    Some((hw as f64 / total as f64) * 100.0)
}

fn storage_from_raw(raw: &collector::RawHostSample) -> StorageSnapshot {
    let used_pct = match (raw.disk_total, raw.disk_used) {
        (Some(t), Some(u)) if t > 0 => Some((u as f64 / t as f64) * 100.0),
        _ => None,
    };
    StorageSnapshot {
        disk_total_bytes: raw.disk_total,
        disk_used_bytes: raw.disk_used,
        disk_free_bytes: raw.disk_free,
        disk_used_percent: used_pct,
        scope: raw.disk_scope,
    }
}

struct CameraGatherSummary {
    total: usize,
    online: usize,
    fps_total: f64,
    queue_depth: u64,
}

async fn gather_cameras(
    states: &Arc<tokio::sync::RwLock<std::collections::HashMap<i64, SharedCameraState>>>,
) -> (CameraGatherSummary, Vec<CameraRuntimeState>) {
    let map = states.read().await;
    let mut total = 0usize;
    let mut online = 0usize;
    let mut fps_total = 0.0f64;
    let mut queue_depth = 0u64;
    let mut cameras = Vec::with_capacity(map.len());
    for cell in map.values() {
        let s = cell.read().await;
        total += 1;
        fps_total += s.fps;
        queue_depth += s.buffer_size;
        if s.status == CameraStatus::Online {
            online += 1;
        }
        cameras.push(s.clone());
    }
    cameras.sort_by_key(|c| c.camera_id);
    (
        CameraGatherSummary {
            total,
            online,
            fps_total,
            queue_depth,
        },
        cameras,
    )
}

fn build_camera_views(cameras: &[CameraRuntimeState], fps_total: f64) -> Vec<CameraCapacityView> {
    cameras
        .iter()
        .map(|c| {
            let drop_rate = if c.frames_received > 0 {
                Some((c.frames_dropped as f64 / c.frames_received as f64) * 100.0)
            } else {
                None
            };
            let share = if fps_total > 0.0 {
                Some((c.fps / fps_total) * 100.0)
            } else {
                None
            };
            CameraCapacityView {
                camera_id: c.camera_id,
                status: format!("{:?}", c.status).to_lowercase(),
                fps: c.fps,
                camera_fps_share: share,
                camera_drop_rate: drop_rate,
                frames_received: c.frames_received,
                frames_processed: c.frames_processed,
                frames_dropped: c.frames_dropped,
                frames_decoded: c.frames_decoded,
                processing_latency_ms: c.last_frame_latency_ms,
            }
        })
        .collect()
}

fn empty_snapshot(cfg: &Config, state: CapacityState, policy: &CapacityPolicy) -> CapacitySnapshot {
    let unavailable = types::ResourceMetric {
        percent: None,
        target_percent: policy.cpu_target_percent,
        headroom_percent: None,
        scope: ResourceScope::Unavailable,
    };
    CapacitySnapshot {
        mode: if cfg.capacity_mode == CapacityMode::Disabled {
            CapacityMode::Disabled
        } else {
            CapacityMode::Dynamic
        },
        state,
        worker_id: cfg.worker_id.clone(),
        processor_id: cfg.processor_id.clone(),
        cpu: unavailable.clone(),
        memory: types::ResourceMetric {
            target_percent: policy.memory_target_percent,
            ..unavailable.clone()
        },
        gpu: types::ResourceMetric {
            target_percent: policy.gpu_target_percent,
            ..unavailable.clone()
        },
        vram: types::ResourceMetric {
            target_percent: policy.vram_target_percent,
            ..unavailable
        },
        network: NetworkSnapshot {
            rx_bytes_total: None,
            tx_bytes_total: None,
            rx_bps: None,
            tx_bps: None,
            scope: ResourceScope::Unavailable,
        },
        storage: StorageSnapshot {
            disk_total_bytes: None,
            disk_used_bytes: None,
            disk_free_bytes: None,
            disk_used_percent: None,
            scope: ResourceScope::Unavailable,
        },
        current_cameras: 0,
        current_online_cameras: 0,
        current_fps: 0.0,
        average_fps_per_camera: None,
        current_frames_received: 0,
        current_frames_processed: 0,
        current_frames_dropped: 0,
        drop_rate_percent: None,
        queue_depth: 0,
        processing_latency_ms: 0,
        estimated_capacity_cameras: None,
        estimated_capacity_fps: None,
        estimated_available_cameras: None,
        estimated_available_fps: None,
        capacity_used_percent: None,
        limiting_resource: LimitingResource::None,
        max_cameras_safety_limit: policy.max_cameras_safety,
        cameras: Vec::new(),
        observation_ready: false,
        samples_in_window: 0,
    }
}

pub async fn run_capacity_sampler(
    engine: Arc<CapacityEngine>,
    metrics: Arc<ProcessorMetrics>,
    camera_states: Arc<tokio::sync::RwLock<std::collections::HashMap<i64, SharedCameraState>>>,
    acceleration: Arc<AccelerationRuntime>,
) {
    let interval = engine.sample_interval();
    loop {
        engine
            .tick(metrics.as_ref(), &camera_states, acceleration.as_ref())
            .await;
        tokio::time::sleep(interval).await;
    }
}
