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
        let mut cfg = Config::test_stub();
        cfg.load_policy_mode = LoadPolicyMode::Admission;
        cfg.load_admission_enabled = true;
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
