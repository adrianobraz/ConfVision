mod policy;
mod types;

pub use policy::{allow_new_camera, evaluate_load};
pub use types::{LoadAdvisory, LoadAdvisoryResult, LoadPolicyConfig, LoadPolicyMode};

use std::sync::Arc;

use crate::capacity::CapacityEngine;

/// Consulta admission local (Fase 6.2) — apenas para novas câmeras.
pub struct LoadAdmissionGate {
    capacity: Arc<CapacityEngine>,
    config: LoadPolicyConfig,
}

impl LoadAdmissionGate {
    pub fn new(capacity: Arc<CapacityEngine>, config: LoadPolicyConfig) -> Arc<Self> {
        Arc::new(Self { capacity, config })
    }

    pub fn config(&self) -> &LoadPolicyConfig {
        &self.config
    }

    pub async fn snapshot_and_evaluate(&self) -> LoadAdvisoryResult {
        let snap = self.capacity.snapshot().await;
        evaluate_load(&snap, &self.config)
    }

    pub async fn allow_new_camera(&self) -> bool {
        let snap = self.capacity.snapshot().await;
        allow_new_camera(&snap, &self.config)
    }
}

#[cfg(test)]
mod admission_tests {
    use super::*;
    use crate::capacity::{
        CapacitySnapshot, CapacityState, LimitingResource, NetworkSnapshot, ResourceMetric,
        ResourceScope, StorageSnapshot,
    };
    use crate::config::{CapacityMode, Config};
    fn metric() -> ResourceMetric {
        ResourceMetric {
            percent: None,
            target_percent: 80.0,
            headroom_percent: None,
            scope: ResourceScope::Unavailable,
        }
    }

    fn critical_snapshot() -> CapacitySnapshot {
        let m = metric();
        CapacitySnapshot {
            mode: CapacityMode::Dynamic,
            state: CapacityState::Critical,
            worker_id: "w".into(),
            processor_id: "p".into(),
            cpu: m.clone(),
            memory: m.clone(),
            gpu: m.clone(),
            vram: m,
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
            current_cameras: 2,
            current_online_cameras: 2,
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
            estimated_available_cameras: Some(0),
            estimated_available_fps: None,
            capacity_used_percent: Some(99.0),
            limiting_resource: LimitingResource::Cpu,
            max_cameras_safety_limit: None,
            cameras: vec![],
            observation_ready: true,
            samples_in_window: 1,
        }
    }

    fn test_cfg() -> Config {
        let mut cfg = Config {
            confvision_api_url: "http://localhost".into(),
            vis_worker_api_key: String::new(),
            mediamtx_rtsp_base: "rtsp://x".into(),
            rtmp_publish_secret: None,
            processor_id: "proc".into(),
            processor_hostname: "h".into(),
            processor_version: "0.1.0".into(),
            worker_id: "worker".into(),
            worker_tipo: "rust_processor".into(),
            shard_mode: crate::config::ShardMode::Auto,
            worker_shard_index: -1,
            worker_shard_total: 0,
            mediamtx_node_id: 0,
            redis_url: None,
            s3_endpoint: None,
            s3_bucket: None,
            http_host: "127.0.0.1".into(),
            http_port: 8090,
            log_level: "info".into(),
            max_cameras: 10,
            sync_interval: std::time::Duration::from_secs(60),
            ping_interval: std::time::Duration::from_secs(30),
            rtsp_connect_timeout: std::time::Duration::from_secs(5),
            rtsp_reconnect_base: std::time::Duration::from_secs(10),
            rtsp_frame_timeout: std::time::Duration::from_secs(30),
            frame_buffer_max: 2,
            queue_backend: "none".into(),
            capacity_mode: CapacityMode::Dynamic,
            capacity_cpu_target_percent: 80.0,
            capacity_memory_target_percent: 80.0,
            capacity_gpu_target_percent: 80.0,
            capacity_vram_target_percent: 80.0,
            capacity_min_sample_sec: 30,
            capacity_safety_factor: 0.80,
            capacity_history_size: 120,
            capacity_sample_interval_sec: 5,
            decode_runtime_fallback: true,
            decode_hw_error_threshold: 10,
            load_policy_mode: LoadPolicyMode::Admission,
            load_admission_enabled: true,
            motion_analysis_max_fps: 0.0,
            motion_frame_stride: 1,
            decode_frame_stride: 1,
            analysis_only_on_motion: false,
            motion_gate_probe_max_fps: 0.5,
            motion_gate_miss_frames: 10,
            motion_probe_keyframe_only: false,
            rtsp_idle_suspend: false,
            motion_pixel_diff_threshold: 8,
            motion_percent_threshold: 5,
        };
        let _ = &mut cfg;
        cfg
    }

    #[tokio::test]
    async fn admission_blocks_new_when_critical() {
        let cfg = test_cfg();
        let engine = CapacityEngine::new_with_initial_snapshot(&cfg, critical_snapshot());
        let gate = LoadAdmissionGate::new(
            engine,
            LoadPolicyConfig {
                mode: LoadPolicyMode::Admission,
                admission_enabled: true,
            },
        );
        assert!(!gate.allow_new_camera().await);
    }

    #[tokio::test]
    async fn advisory_with_admission_disabled_allows_new_cameras() {
        let cfg = test_cfg();
        let engine = CapacityEngine::new_with_initial_snapshot(&cfg, critical_snapshot());
        let gate = LoadAdmissionGate::new(
            engine,
            LoadPolicyConfig {
                mode: LoadPolicyMode::Advisory,
                admission_enabled: false,
            },
        );
        assert!(gate.allow_new_camera().await);
    }

    #[tokio::test]
    async fn advisory_with_load_admission_enabled_blocks_on_critical() {
        let cfg = test_cfg();
        let engine = CapacityEngine::new_with_initial_snapshot(&cfg, critical_snapshot());
        let gate = LoadAdmissionGate::new(
            engine,
            LoadPolicyConfig {
                mode: LoadPolicyMode::Advisory,
                admission_enabled: true,
            },
        );
        assert!(!gate.allow_new_camera().await);
    }
}
