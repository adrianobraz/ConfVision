use std::sync::atomic::{AtomicBool, AtomicU64, Ordering};
use std::sync::Arc;

use tracing::info;

use crate::error::{AppError, AppResult};

use super::policy::{backend_label, mode_label, AccelerationPolicy, VideoAccelerationMode};
use super::probe::{DefaultSystemProbe, ProbeResult, SystemProbe};

const HW_BACKEND_NOT_IMPLEMENTED: &str =
    "VIDEO_ACCELERATION=gpu requer decode por hardware (Fase 4.1); backend HW ainda não implementado";

/// Runtime de aceleração — probe no boot + modo efetivo (Fase 4.0: sempre CPU no decode).
#[derive(Debug)]
pub struct AccelerationRuntime {
    policy: AccelerationPolicy,
    probe: ProbeResult,
    effective: VideoAccelerationMode,
    hw_decode_errors: AtomicU64,
    hw_fallback_to_cpu_count: AtomicU64,
    hardware_decode_active: AtomicBool,
}

impl AccelerationRuntime {
    pub fn bootstrap(policy: AccelerationPolicy) -> AppResult<Arc<Self>> {
        Self::bootstrap_with_probe(policy, &DefaultSystemProbe)
    }

    pub fn bootstrap_with_probe(
        policy: AccelerationPolicy,
        system: &dyn SystemProbe,
    ) -> AppResult<Arc<Self>> {
        let probe = super::probe::run_probe(system, policy.gpu_backend);

        let effective = match policy.mode {
            VideoAccelerationMode::Cpu => VideoAccelerationMode::Cpu,
            VideoAccelerationMode::Auto => VideoAccelerationMode::Cpu,
            VideoAccelerationMode::Gpu => {
                return Err(AppError::Config(HW_BACKEND_NOT_IMPLEMENTED.into()));
            }
        };

        Ok(Arc::new(Self {
            policy,
            probe,
            effective,
            hw_decode_errors: AtomicU64::new(0),
            hw_fallback_to_cpu_count: AtomicU64::new(0),
            hardware_decode_active: AtomicBool::new(false),
        }))
    }

    pub fn policy(&self) -> &AccelerationPolicy {
        &self.policy
    }

    pub fn probe(&self) -> &ProbeResult {
        &self.probe
    }

    pub fn effective_mode(&self) -> VideoAccelerationMode {
        self.effective
    }

    pub fn log_startup(&self) {
        info!(
            video_acceleration_requested = mode_label(self.policy.mode),
            video_gpu_backend_requested = backend_label(self.policy.gpu_backend),
            gpu_detected = self.probe.gpu_detected,
            ffmpeg_hw_decode_available = self.probe.ffmpeg_hw_decode_available,
            video_hw_backend = self.probe.inferred_hw_backend,
            video_acceleration_effective = mode_label(self.effective),
            hardware_decode_active = false,
            "video acceleration initialized (Fase 4.0 — decode CPU)"
        );
    }

    pub fn video_acceleration_requested(&self) -> &'static str {
        mode_label(self.policy.mode)
    }

    pub fn video_acceleration_effective(&self) -> &'static str {
        mode_label(self.effective)
    }

    pub fn video_hw_backend(&self) -> &'static str {
        self.probe.inferred_hw_backend
    }

    pub fn gpu_detected(&self) -> bool {
        self.probe.gpu_detected
    }

    pub fn ffmpeg_hw_decode_available(&self) -> bool {
        self.probe.ffmpeg_hw_decode_available
    }

    pub fn hardware_decode_active(&self) -> bool {
        self.hardware_decode_active.load(Ordering::Relaxed)
    }

    pub fn hw_decode_errors(&self) -> u64 {
        self.hw_decode_errors.load(Ordering::Relaxed)
    }

    pub fn hw_fallback_to_cpu_count(&self) -> u64 {
        self.hw_fallback_to_cpu_count.load(Ordering::Relaxed)
    }
}

#[cfg(test)]
mod tests {
    use super::super::policy::VideoGpuBackend;
    use super::super::probe::MockProbe;
    use super::*;

    #[test]
    fn auto_without_gpu_effective_cpu() {
        let policy =
            AccelerationPolicy::from_parts(VideoAccelerationMode::Auto, VideoGpuBackend::Auto);
        let probe = MockProbe {
            nvidia: false,
            vaapi: false,
            qsv: false,
            ffmpeg_hw: false,
        };
        let rt = AccelerationRuntime::bootstrap_with_probe(policy, &probe).unwrap();
        assert_eq!(rt.effective_mode(), VideoAccelerationMode::Cpu);
        assert!(!rt.gpu_detected());
    }

    #[test]
    fn auto_with_gpu_still_effective_cpu_in_phase_40() {
        let policy =
            AccelerationPolicy::from_parts(VideoAccelerationMode::Auto, VideoGpuBackend::Auto);
        let probe = MockProbe {
            nvidia: true,
            vaapi: false,
            qsv: false,
            ffmpeg_hw: true,
        };
        let rt = AccelerationRuntime::bootstrap_with_probe(policy, &probe).unwrap();
        assert_eq!(rt.effective_mode(), VideoAccelerationMode::Cpu);
        assert!(rt.gpu_detected());
        assert_eq!(rt.video_hw_backend(), "cuda");
    }

    #[test]
    fn cpu_forces_effective_cpu_even_with_gpu() {
        let policy =
            AccelerationPolicy::from_parts(VideoAccelerationMode::Cpu, VideoGpuBackend::Cuda);
        let probe = MockProbe {
            nvidia: true,
            vaapi: false,
            qsv: false,
            ffmpeg_hw: true,
        };
        let rt = AccelerationRuntime::bootstrap_with_probe(policy, &probe).unwrap();
        assert_eq!(rt.effective_mode(), VideoAccelerationMode::Cpu);
    }

    #[test]
    fn gpu_mode_errors_without_silent_fallback() {
        let policy =
            AccelerationPolicy::from_parts(VideoAccelerationMode::Gpu, VideoGpuBackend::Auto);
        let probe = MockProbe {
            nvidia: true,
            vaapi: false,
            qsv: false,
            ffmpeg_hw: true,
        };
        let err = AccelerationRuntime::bootstrap_with_probe(policy, &probe);
        assert!(err.is_err());
        let msg = err.unwrap_err().to_string();
        assert!(msg.contains("gpu") || msg.contains("GPU") || msg.contains("hardware"));
        assert!(msg.contains("4.1") || msg.contains("implementado"));
    }

    #[test]
    fn gpu_mode_errors_even_without_gpu() {
        let policy =
            AccelerationPolicy::from_parts(VideoAccelerationMode::Gpu, VideoGpuBackend::Auto);
        let probe = MockProbe {
            nvidia: false,
            vaapi: false,
            qsv: false,
            ffmpeg_hw: false,
        };
        assert!(AccelerationRuntime::bootstrap_with_probe(policy, &probe).is_err());
    }
}
