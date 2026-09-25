use crate::capacity::policy::CapacityPolicy;
use crate::capacity::types::{
    CapacityState, HistorySample, LimitingResource, ResourceMetric, ResourceScope,
};

#[derive(Debug, Clone, Default)]
pub struct AveragedSignals {
    pub cpu_percent: Option<f64>,
    pub memory_percent: Option<f64>,
    pub gpu_percent: Option<f64>,
    pub vram_percent: Option<f64>,
    pub fps: f64,
    pub online_cameras: f64,
    pub total_cameras: f64,
    pub queue_depth: f64,
    pub drop_rate_percent: Option<f64>,
    pub processing_latency_ms: f64,
    pub samples: usize,
}

pub fn average_history(history: &[HistorySample]) -> AveragedSignals {
    if history.is_empty() {
        return AveragedSignals::default();
    }
    let n = history.len() as f64;
    let mut cpu_sum = 0.0;
    let mut cpu_n = 0usize;
    let mut mem_sum = 0.0;
    let mut mem_n = 0usize;
    let mut gpu_sum = 0.0;
    let mut gpu_n = 0usize;
    let mut vram_sum = 0.0;
    let mut vram_n = 0usize;
    let mut fps = 0.0;
    let mut online = 0.0;
    let mut total = 0.0;
    let mut queue = 0.0;
    let mut latency = 0.0;
    let mut drop_num = 0u64;
    let mut recv_num = 0u64;

    for s in history {
        if let Some(c) = s.cpu_percent {
            cpu_sum += c;
            cpu_n += 1;
        }
        if let Some(m) = s.memory_percent {
            mem_sum += m;
            mem_n += 1;
        }
        if let Some(g) = s.gpu_percent {
            gpu_sum += g;
            gpu_n += 1;
        }
        if let Some(v) = s.vram_percent {
            vram_sum += v;
            vram_n += 1;
        }
        fps += s.fps;
        online += s.online_cameras as f64;
        total += s.total_cameras as f64;
        queue += s.queue_depth as f64;
        latency += s.processing_latency_ms as f64;
        drop_num += s.frames_dropped_delta;
        recv_num += s.frames_received_delta;
    }

    AveragedSignals {
        cpu_percent: if cpu_n > 0 {
            Some(cpu_sum / cpu_n as f64)
        } else {
            None
        },
        memory_percent: if mem_n > 0 {
            Some(mem_sum / mem_n as f64)
        } else {
            None
        },
        gpu_percent: if gpu_n > 0 {
            Some(gpu_sum / gpu_n as f64)
        } else {
            None
        },
        vram_percent: if vram_n > 0 {
            Some(vram_sum / vram_n as f64)
        } else {
            None
        },
        fps: fps / n,
        online_cameras: online / n,
        total_cameras: total / n,
        queue_depth: queue / n,
        drop_rate_percent: if recv_num > 0 {
            Some((drop_num as f64 / recv_num as f64) * 100.0)
        } else {
            None
        },
        processing_latency_ms: latency / n,
        samples: history.len(),
    }
}

pub struct EstimateInput<'a> {
    pub policy: &'a CapacityPolicy,
    pub signals: &'a AveragedSignals,
    pub observation_ready: bool,
    pub cpu_scope: ResourceScope,
    pub memory_scope: ResourceScope,
    pub gpu_scope: ResourceScope,
    pub vram_scope: ResourceScope,
    pub gpu_detected: bool,
}

pub struct EstimateOutput {
    pub state: CapacityState,
    pub avg_fps_per_camera: Option<f64>,
    pub estimated_capacity_cameras: Option<u32>,
    pub estimated_capacity_fps: Option<f64>,
    pub estimated_available_cameras: Option<i32>,
    pub estimated_available_fps: Option<f64>,
    pub capacity_used_percent: Option<f64>,
    pub limiting_resource: LimitingResource,
    pub cpu_metric: ResourceMetric,
    pub memory_metric: ResourceMetric,
    pub gpu_metric: ResourceMetric,
    pub vram_metric: ResourceMetric,
}

pub fn compute_estimate(input: EstimateInput<'_>) -> EstimateOutput {
    let p = input.policy;
    let s = input.signals;

    let cpu_metric = metric(s.cpu_percent, p.cpu_target_percent, input.cpu_scope);
    let memory_metric = metric(
        s.memory_percent,
        p.memory_target_percent,
        input.memory_scope,
    );
    let gpu_metric = metric(
        if input.gpu_detected {
            s.gpu_percent
        } else {
            None
        },
        p.gpu_target_percent,
        if input.gpu_detected {
            input.gpu_scope
        } else {
            ResourceScope::Unavailable
        },
    );
    let vram_metric = metric(
        if input.gpu_detected {
            s.vram_percent
        } else {
            None
        },
        p.vram_target_percent,
        if input.gpu_detected {
            input.vram_scope
        } else {
            ResourceScope::Unavailable
        },
    );

    if !input.observation_ready || s.samples == 0 {
        return EstimateOutput {
            state: CapacityState::Unknown,
            avg_fps_per_camera: None,
            estimated_capacity_cameras: None,
            estimated_capacity_fps: None,
            estimated_available_cameras: None,
            estimated_available_fps: None,
            capacity_used_percent: None,
            limiting_resource: LimitingResource::None,
            cpu_metric,
            memory_metric,
            gpu_metric,
            vram_metric,
        };
    }

    let online = s.online_cameras.round() as usize;
    let current_fps = s.fps;

    let avg_fps_per_camera = if online > 0 && current_fps > 0.0 {
        Some(current_fps / s.online_cameras)
    } else {
        None
    };

    if online == 0 || current_fps <= 0.0 {
        return EstimateOutput {
            state: classify_usage(s.cpu_percent, s.memory_percent, p, CapacityState::Healthy),
            avg_fps_per_camera,
            estimated_capacity_cameras: None,
            estimated_capacity_fps: None,
            estimated_available_cameras: None,
            estimated_available_fps: None,
            capacity_used_percent: None,
            limiting_resource: LimitingResource::None,
            cpu_metric,
            memory_metric,
            gpu_metric,
            vram_metric,
        };
    }

    let avg_fps = avg_fps_per_camera.unwrap_or(current_fps / online as f64);
    if avg_fps <= 0.0 {
        return unknown_partial(cpu_metric, memory_metric, gpu_metric, vram_metric);
    }

    let mut candidates: Vec<(LimitingResource, f64, f64)> = Vec::new();

    if let Some(cpu) = s.cpu_percent {
        if cpu >= 0.5 {
            let cap_fps = current_fps * p.effective_cpu_target() / cpu;
            let cap_cam = cap_fps / avg_fps;
            candidates.push((LimitingResource::Cpu, cap_fps, cap_cam));
        }
    }

    if let Some(mem) = s.memory_percent {
        if mem >= 0.5 {
            let cap_fps = current_fps * p.effective_memory_target() / mem;
            let cap_cam = cap_fps / avg_fps;
            candidates.push((LimitingResource::Memory, cap_fps, cap_cam));
        }
    }

    if input.gpu_detected {
        if let Some(gpu) = s.gpu_percent {
            if gpu >= 0.5 {
                let cap_fps = current_fps * (p.gpu_target_percent * p.safety_factor) / gpu;
                let cap_cam = cap_fps / avg_fps;
                candidates.push((LimitingResource::Gpu, cap_fps, cap_cam));
            }
        }
        if let Some(vram) = s.vram_percent {
            if vram >= 0.5 {
                let cap_fps = current_fps * (p.vram_target_percent * p.safety_factor) / vram;
                let cap_cam = cap_fps / avg_fps;
                candidates.push((LimitingResource::Vram, cap_fps, cap_cam));
            }
        }
    }

    if candidates.is_empty() {
        return EstimateOutput {
            state: CapacityState::Unknown,
            avg_fps_per_camera: Some(avg_fps),
            estimated_capacity_cameras: None,
            estimated_capacity_fps: None,
            estimated_available_cameras: None,
            estimated_available_fps: None,
            capacity_used_percent: None,
            limiting_resource: LimitingResource::None,
            cpu_metric,
            memory_metric,
            gpu_metric,
            vram_metric,
        };
    }

    let (limiting, mut cap_fps, mut cap_cam) = candidates
        .into_iter()
        .min_by(|a, b| a.2.partial_cmp(&b.2).unwrap_or(std::cmp::Ordering::Equal))
        .unwrap();

    if let Some(hard) = p.max_cameras_safety {
        if cap_cam > hard as f64 {
            cap_cam = hard as f64;
            cap_fps = cap_cam * avg_fps;
            if limiting != LimitingResource::SafetyLimit {
                // safety is stricter than resource estimate
            }
        }
    }

    let cap_cam_u = cap_cam.floor().max(0.0) as u32;
    let cap_fps_r = (cap_fps * 10.0).floor() / 10.0;

    let available_cam = cap_cam_u as i32 - online as i32;
    let available_fps = (cap_fps - current_fps).max(0.0);

    let used_pct = if cap_cam_u > 0 {
        Some((online as f64 / cap_cam_u as f64) * 100.0)
    } else {
        None
    };

    let mut state = classify_usage(s.cpu_percent, s.memory_percent, p, CapacityState::Healthy);
    if let Some(u) = used_pct {
        if u >= 100.0 {
            state = CapacityState::Critical;
        } else if u >= 90.0 {
            state = state.max_severity(CapacityState::Warning);
        }
    }

    let final_limiting = if p.max_cameras_safety == Some(online)
        && cap_cam_u == p.max_cameras_safety.unwrap() as u32
    {
        LimitingResource::SafetyLimit
    } else {
        limiting
    };

    EstimateOutput {
        state,
        avg_fps_per_camera: Some(avg_fps),
        estimated_capacity_cameras: Some(cap_cam_u),
        estimated_capacity_fps: Some(cap_fps_r),
        estimated_available_cameras: Some(available_cam.max(0)),
        estimated_available_fps: Some((available_fps * 10.0).floor() / 10.0),
        capacity_used_percent: used_pct,
        limiting_resource: final_limiting,
        cpu_metric,
        memory_metric,
        gpu_metric,
        vram_metric,
    }
}

impl CapacityState {
    fn max_severity(self, other: CapacityState) -> CapacityState {
        use CapacityState::*;
        let rank = |s: CapacityState| match s {
            Unknown => 0,
            Healthy => 1,
            Warning => 2,
            Critical => 3,
        };
        if rank(other) > rank(self) {
            other
        } else {
            self
        }
    }
}

fn classify_usage(
    cpu: Option<f64>,
    memory: Option<f64>,
    policy: &CapacityPolicy,
    base: CapacityState,
) -> CapacityState {
    let mut state = base;
    for (value, target) in [
        (cpu, policy.cpu_target_percent),
        (memory, policy.memory_target_percent),
    ] {
        if let Some(v) = value {
            if v >= target * policy.critical_ratio {
                state = CapacityState::Critical;
            } else if v >= target * policy.warning_ratio {
                state = state.max_severity(CapacityState::Warning);
            }
        }
    }
    state
}

fn metric(percent: Option<f64>, target: f64, scope: ResourceScope) -> ResourceMetric {
    ResourceMetric {
        headroom_percent: percent.map(|p| (target - p).max(0.0)),
        percent,
        target_percent: target,
        scope,
    }
}

fn unknown_partial(
    cpu_metric: ResourceMetric,
    memory_metric: ResourceMetric,
    gpu_metric: ResourceMetric,
    vram_metric: ResourceMetric,
) -> EstimateOutput {
    EstimateOutput {
        state: CapacityState::Unknown,
        avg_fps_per_camera: None,
        estimated_capacity_cameras: None,
        estimated_capacity_fps: None,
        estimated_available_cameras: None,
        estimated_available_fps: None,
        capacity_used_percent: None,
        limiting_resource: LimitingResource::None,
        cpu_metric,
        memory_metric,
        gpu_metric,
        vram_metric,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::capacity::policy::CapacityPolicy;
    use crate::config::{CapacityMode, Config, ShardMode};
    use std::time::Duration;

    fn policy() -> CapacityPolicy {
        CapacityPolicy::from_config(&Config {
            confvision_api_url: "http://x".into(),
            vis_worker_api_key: String::new(),
            mediamtx_rtsp_base: "rtsp://x".into(),
            rtmp_publish_secret: None,
            processor_id: "p".into(),
            processor_hostname: "h".into(),
            processor_version: "0.1.0".into(),
            worker_id: "w".into(),
            worker_tipo: "rust".into(),
            shard_mode: ShardMode::Auto,
            worker_shard_index: -1,
            worker_shard_total: 0,
            mediamtx_node_id: 0,
            redis_url: None,
            s3_endpoint: None,
            s3_bucket: None,
            http_host: "0.0.0.0".into(),
            http_port: 8090,
            log_level: "info".into(),
            max_cameras: 100,
            sync_interval: Duration::from_secs(60),
            ping_interval: Duration::from_secs(30),
            rtsp_connect_timeout: Duration::from_secs(5),
            rtsp_reconnect_base: Duration::from_secs(10),
            rtsp_frame_timeout: Duration::from_secs(30),
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
            load_policy_mode: crate::load::LoadPolicyMode::Advisory,
            load_admission_enabled: false,
            motion_analysis_max_fps: 0.0,
        })
    }

    fn signals(cpu: f64, mem: f64, fps: f64, online: usize) -> AveragedSignals {
        AveragedSignals {
            cpu_percent: Some(cpu),
            memory_percent: Some(mem),
            gpu_percent: None,
            vram_percent: None,
            fps,
            online_cameras: online as f64,
            total_cameras: online as f64,
            queue_depth: 0.0,
            drop_rate_percent: Some(0.1),
            processing_latency_ms: 1.0,
            samples: 10,
        }
    }

    fn run(cpu: f64, mem: f64, fps: f64, online: usize) -> EstimateOutput {
        let p = policy();
        compute_estimate(EstimateInput {
            policy: &p,
            signals: &signals(cpu, mem, fps, online),
            observation_ready: true,
            cpu_scope: ResourceScope::Host,
            memory_scope: ResourceScope::Host,
            gpu_scope: ResourceScope::Unavailable,
            vram_scope: ResourceScope::Unavailable,
            gpu_detected: false,
        })
    }

    #[test]
    fn headroom_cpu() {
        let out = run(70.0, 40.0, 110.0, 10);
        assert!(out.cpu_metric.headroom_percent.unwrap() > 0.0);
    }

    #[test]
    fn capacity_scales_with_cpu() {
        let out = run(80.0, 30.0, 120.0, 10);
        assert_eq!(out.limiting_resource, LimitingResource::Cpu);
        let cap = out.estimated_capacity_cameras.unwrap();
        assert!(cap >= 8, "expected headroom from CPU scaling, got {cap}");
    }

    #[test]
    fn memory_can_limit() {
        let out = run(20.0, 79.0, 100.0, 5);
        assert_eq!(out.limiting_resource, LimitingResource::Memory);
    }

    #[test]
    fn unknown_without_observation() {
        let p = policy();
        let s = signals(50.0, 50.0, 100.0, 5);
        let out = compute_estimate(EstimateInput {
            policy: &p,
            signals: &s,
            observation_ready: false,
            cpu_scope: ResourceScope::Host,
            memory_scope: ResourceScope::Host,
            gpu_scope: ResourceScope::Unavailable,
            vram_scope: ResourceScope::Unavailable,
            gpu_detected: false,
        });
        assert_eq!(out.state, CapacityState::Unknown);
        assert!(out.estimated_capacity_cameras.is_none());
    }

    #[test]
    fn zero_cameras_no_capacity_numbers() {
        let out = run(40.0, 40.0, 0.0, 0);
        assert!(out.estimated_capacity_cameras.is_none());
    }

    #[test]
    fn zero_fps_no_capacity_numbers() {
        let out = run(40.0, 40.0, 0.0, 3);
        assert!(out.estimated_capacity_cameras.is_none());
    }

    #[test]
    fn zero_cpu_no_crash() {
        let p = policy();
        let mut s = signals(0.0, 40.0, 100.0, 5);
        s.memory_percent = None;
        let out = compute_estimate(EstimateInput {
            policy: &p,
            signals: &s,
            observation_ready: true,
            cpu_scope: ResourceScope::Host,
            memory_scope: ResourceScope::Unavailable,
            gpu_scope: ResourceScope::Unavailable,
            vram_scope: ResourceScope::Unavailable,
            gpu_detected: false,
        });
        assert_eq!(out.state, CapacityState::Unknown);
        assert!(out.estimated_capacity_cameras.is_none());
    }

    #[test]
    fn gpu_unavailable_still_estimates() {
        let out = run(50.0, 50.0, 100.0, 5);
        assert_eq!(out.gpu_metric.scope, ResourceScope::Unavailable);
        assert!(out.estimated_capacity_cameras.is_some());
    }

    #[test]
    fn floors_camera_capacity() {
        let out = run(79.9, 10.0, 122.1, 11);
        let cap = out.estimated_capacity_cameras.unwrap();
        let raw: f64 = 122.1 * (80.0 * 0.8) / 79.9 / (122.1 / 11.0);
        assert_eq!(cap, raw.floor() as u32);
    }

    #[test]
    fn safety_factor_reduces_capacity_vs_raw_target() {
        let p = policy();
        let s = signals(50.0, 20.0, 100.0, 5);
        let out = compute_estimate(EstimateInput {
            policy: &p,
            signals: &s,
            observation_ready: true,
            cpu_scope: ResourceScope::Host,
            memory_scope: ResourceScope::Host,
            gpu_scope: ResourceScope::Unavailable,
            vram_scope: ResourceScope::Unavailable,
            gpu_detected: false,
        });
        let cap = out.estimated_capacity_cameras.unwrap() as f64;
        let naive: f64 = 100.0 * 80.0 / 50.0 / (100.0 / 5.0);
        assert!(cap <= naive.floor());
    }

    #[test]
    fn warning_when_near_cpu_target() {
        let out = run(75.0, 20.0, 100.0, 8);
        assert!(matches!(
            out.state,
            CapacityState::Warning | CapacityState::Critical
        ));
    }

    #[test]
    fn critical_above_target() {
        let out = run(85.0, 20.0, 100.0, 8);
        assert_eq!(out.state, CapacityState::Critical);
    }

    #[test]
    fn healthy_below_thresholds() {
        let out = run(40.0, 40.0, 80.0, 4);
        assert_eq!(out.state, CapacityState::Healthy);
    }

    #[test]
    fn drop_rate_in_signals() {
        let mut s = signals(40.0, 40.0, 80.0, 4);
        s.drop_rate_percent = Some(2.5);
        assert!((s.drop_rate_percent.unwrap() - 2.5).abs() < 0.01);
    }

    #[test]
    fn avg_fps_per_camera() {
        let out = run(50.0, 40.0, 110.0, 10);
        assert!((out.avg_fps_per_camera.unwrap() - 11.0).abs() < 0.01);
    }
}
