use std::time::Duration;

use crate::config::{CapacityMode, Config};

#[derive(Debug, Clone)]
pub struct CapacityPolicy {
    pub mode: CapacityMode,
    pub cpu_target_percent: f64,
    pub memory_target_percent: f64,
    pub gpu_target_percent: f64,
    pub vram_target_percent: f64,
    pub min_sample: Duration,
    pub safety_factor: f64,
    pub history_size: usize,
    pub sample_interval: Duration,
    /// Limite duro local (não é capacidade estimada). `None` = sem teto.
    pub max_cameras_safety: Option<usize>,
    pub warning_ratio: f64,
    pub critical_ratio: f64,
}

impl CapacityPolicy {
    pub fn from_config(cfg: &Config) -> Self {
        Self {
            mode: cfg.capacity_mode,
            cpu_target_percent: cfg.capacity_cpu_target_percent,
            memory_target_percent: cfg.capacity_memory_target_percent,
            gpu_target_percent: cfg.capacity_gpu_target_percent,
            vram_target_percent: cfg.capacity_vram_target_percent,
            min_sample: Duration::from_secs(cfg.capacity_min_sample_sec),
            safety_factor: cfg.capacity_safety_factor,
            history_size: cfg.capacity_history_size.max(1),
            sample_interval: Duration::from_secs(cfg.capacity_sample_interval_sec.max(1)),
            max_cameras_safety: cfg.max_cameras_hard_limit(),
            warning_ratio: 0.9,
            critical_ratio: 1.0,
        }
    }

    pub fn effective_cpu_target(&self) -> f64 {
        self.cpu_target_percent * self.safety_factor
    }

    pub fn effective_memory_target(&self) -> f64 {
        self.memory_target_percent * self.safety_factor
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::config::ShardMode;
    use std::time::Duration;

    fn base_config() -> Config {
        Config {
            confvision_api_url: "http://localhost".into(),
            vis_worker_api_key: String::new(),
            mediamtx_rtsp_base: "rtsp://localhost".into(),
            rtmp_publish_secret: None,
            processor_id: "p".into(),
            processor_hostname: "h".into(),
            processor_version: "0.1.0".into(),
            worker_id: "w".into(),
            worker_tipo: "rust_processor".into(),
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
            max_cameras: 10,
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
        }
    }

    #[test]
    fn safety_factor_applied_to_targets() {
        let policy = CapacityPolicy::from_config(&base_config());
        assert!((policy.effective_cpu_target() - 64.0).abs() < 0.01);
    }

    #[test]
    fn max_cameras_zero_means_no_hard_limit() {
        let mut cfg = base_config();
        cfg.max_cameras = 0;
        let policy = CapacityPolicy::from_config(&cfg);
        assert!(policy.max_cameras_safety.is_none());
    }
}
