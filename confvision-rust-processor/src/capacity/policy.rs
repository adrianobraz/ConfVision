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
        Config::test_stub()
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
