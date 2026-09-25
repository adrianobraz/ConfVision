use crate::capacity::{CapacitySnapshot, CapacityState, LimitingResource};

use super::types::{LoadAdvisory, LoadAdvisoryResult, LoadPolicyConfig, LoadPolicyMode};

pub fn evaluate_load(capacity: &CapacitySnapshot, config: &LoadPolicyConfig) -> LoadAdvisoryResult {
    if matches!(config.mode, LoadPolicyMode::Disabled) {
        return LoadAdvisoryResult {
            advisory: LoadAdvisory::Normal,
            reason: "load_policy_disabled".into(),
        };
    }

    let mut reasons = Vec::new();

    let advisory = match capacity.state {
        CapacityState::Unknown => {
            reasons.push("capacity_unknown");
            LoadAdvisory::Caution
        }
        CapacityState::Healthy => LoadAdvisory::Normal,
        CapacityState::Warning => {
            reasons.push("capacity_warning");
            if let Some(pct) = capacity.capacity_used_percent {
                if pct >= 85.0 {
                    reasons.push("high_capacity_used");
                }
            }
            LoadAdvisory::Caution
        }
        CapacityState::Critical => {
            reasons.push("capacity_critical");
            if config.admission_active() {
                reasons.push("admission_reject");
                LoadAdvisory::RejectAdmission
            } else {
                LoadAdvisory::Saturated
            }
        }
    };

    if capacity.limiting_resource != LimitingResource::None {
        reasons.push(match capacity.limiting_resource {
            LimitingResource::Cpu => "limiting_cpu",
            LimitingResource::Memory => "limiting_memory",
            LimitingResource::Gpu => "limiting_gpu",
            LimitingResource::Vram => "limiting_vram",
            LimitingResource::SafetyLimit => "limiting_safety",
            LimitingResource::None => "limiting_none",
        });
    }

    if let Some(drop) = capacity.drop_rate_percent {
        if drop > 1.0 {
            reasons.push("elevated_drop_rate");
        }
    }

    if capacity.processing_latency_ms > 50 {
        reasons.push("elevated_latency");
    }

    if let Some(avail) = capacity.estimated_available_cameras {
        if avail <= 0 {
            reasons.push("no_estimated_camera_headroom");
        }
    }

    let mut final_advisory = advisory;
    if !matches!(capacity.state, CapacityState::Critical)
        && final_advisory == LoadAdvisory::Normal
        && reasons
            .iter()
            .any(|r| *r == "elevated_drop_rate" || *r == "elevated_latency")
    {
        final_advisory = LoadAdvisory::Caution;
    }

    LoadAdvisoryResult {
        advisory: final_advisory,
        reason: reasons.join(","),
    }
}

pub fn allow_new_camera(capacity: &CapacitySnapshot, config: &LoadPolicyConfig) -> bool {
    !config.admission_blocks_new_cameras(capacity)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::capacity::{
        CapacitySnapshot, CapacityState, LimitingResource, NetworkSnapshot, ResourceMetric,
        ResourceScope, StorageSnapshot,
    };
    use crate::config::CapacityMode;

    fn metric() -> ResourceMetric {
        ResourceMetric {
            percent: None,
            target_percent: 80.0,
            headroom_percent: None,
            scope: ResourceScope::Unavailable,
        }
    }

    fn empty_snapshot(state: CapacityState) -> CapacitySnapshot {
        let m = metric();
        CapacitySnapshot {
            mode: CapacityMode::Dynamic,
            state,
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
            max_cameras_safety_limit: None,
            cameras: vec![],
            observation_ready: false,
            samples_in_window: 0,
        }
    }

    fn cfg(mode: LoadPolicyMode, admission: bool) -> LoadPolicyConfig {
        LoadPolicyConfig {
            mode,
            admission_enabled: admission,
        }
    }

    #[test]
    fn unknown_is_caution() {
        let r = evaluate_load(
            &empty_snapshot(CapacityState::Unknown),
            &cfg(LoadPolicyMode::Advisory, false),
        );
        assert_eq!(r.advisory, LoadAdvisory::Caution);
    }

    #[test]
    fn healthy_is_normal() {
        let r = evaluate_load(
            &empty_snapshot(CapacityState::Healthy),
            &cfg(LoadPolicyMode::Advisory, false),
        );
        assert_eq!(r.advisory, LoadAdvisory::Normal);
    }

    #[test]
    fn warning_is_caution() {
        let r = evaluate_load(
            &empty_snapshot(CapacityState::Warning),
            &cfg(LoadPolicyMode::Advisory, false),
        );
        assert_eq!(r.advisory, LoadAdvisory::Caution);
    }

    #[test]
    fn critical_admission_disabled_is_saturated() {
        let r = evaluate_load(
            &empty_snapshot(CapacityState::Critical),
            &cfg(LoadPolicyMode::Admission, false),
        );
        assert_eq!(r.advisory, LoadAdvisory::Saturated);
    }

    #[test]
    fn critical_admission_enabled_rejects() {
        let snap = empty_snapshot(CapacityState::Critical);
        let r = evaluate_load(&snap, &cfg(LoadPolicyMode::Admission, true));
        assert_eq!(r.advisory, LoadAdvisory::RejectAdmission);
        assert!(!allow_new_camera(
            &snap,
            &cfg(LoadPolicyMode::Admission, true)
        ));
    }

    #[test]
    fn critical_advisory_admission_enabled_blocks_but_stays_saturated() {
        let snap = empty_snapshot(CapacityState::Critical);
        let r = evaluate_load(&snap, &cfg(LoadPolicyMode::Advisory, true));
        assert_eq!(r.advisory, LoadAdvisory::Saturated);
        assert!(!allow_new_camera(
            &snap,
            &cfg(LoadPolicyMode::Advisory, true)
        ));
    }

    #[test]
    fn disabled_policy_is_normal() {
        let r = evaluate_load(
            &empty_snapshot(CapacityState::Critical),
            &cfg(LoadPolicyMode::Disabled, true),
        );
        assert_eq!(r.advisory, LoadAdvisory::Normal);
    }
}
