use std::sync::atomic::{AtomicBool, AtomicU64, Ordering};
use std::sync::Arc;

use tracing::info;

use crate::error::{AppError, AppResult};

use super::policy::{
    backend_label, mode_label, AccelerationPolicy, VideoAccelerationMode, VideoGpuBackend,
};
use super::probe::{DefaultSystemProbe, ProbeResult, SystemProbe};

/// Backend de decode planejado no boot (por câmera usa o mesmo plano).
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum PlannedDecodeBackend {
    Cpu,
    #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
    Nvdec,
}

impl PlannedDecodeBackend {
    pub fn log_label(self) -> &'static str {
        match self {
            Self::Cpu => "cpu",
            #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
            Self::Nvdec => "nvdec",
        }
    }
}

/// Runtime de aceleração — probe no boot + modo efetivo + métricas HW.
#[derive(Debug)]
pub struct AccelerationRuntime {
    policy: AccelerationPolicy,
    probe: ProbeResult,
    effective: VideoAccelerationMode,
    planned_backend: PlannedDecodeBackend,
    nvdec_cuda_init_ok: bool,
    hw_decode_errors: AtomicU64,
    hw_fallback_to_cpu_count: AtomicU64,
    hardware_decode_active: AtomicBool,
    frames_hw_decoded: AtomicU64,
    frames_cpu_decoded: AtomicU64,
    last_hw_decode_ms: AtomicU64,
    hw_transfer_to_cpu_ms: AtomicU64,
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
        let nvdec_cuda_init_ok = system.nvdec_cuda_init_ok();
        let cuvid = system.ffmpeg_h264_cuvid_decoder_available();
        let nvdec_stack_ready =
            nvdec_ready_for_policy(policy.gpu_backend, &probe, nvdec_cuda_init_ok, cuvid);

        let (effective, planned_backend) =
            resolve_effective_and_backend(policy.mode, policy.gpu_backend, nvdec_stack_ready)?;

        Ok(Arc::new(Self {
            policy,
            probe,
            effective,
            planned_backend,
            nvdec_cuda_init_ok,
            hw_decode_errors: AtomicU64::new(0),
            hw_fallback_to_cpu_count: AtomicU64::new(0),
            hardware_decode_active: AtomicBool::new(false),
            frames_hw_decoded: AtomicU64::new(0),
            frames_cpu_decoded: AtomicU64::new(0),
            last_hw_decode_ms: AtomicU64::new(0),
            hw_transfer_to_cpu_ms: AtomicU64::new(0),
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

    pub fn planned_decode_backend(&self) -> PlannedDecodeBackend {
        self.planned_backend
    }

    pub fn nvdec_cuda_init_ok(&self) -> bool {
        self.nvdec_cuda_init_ok
    }

    pub fn log_startup(&self) {
        info!(
            video_acceleration_requested = mode_label(self.policy.mode),
            video_gpu_backend_requested = backend_label(self.policy.gpu_backend),
            gpu_detected = self.probe.gpu_detected,
            ffmpeg_hw_decode_available = self.probe.ffmpeg_hw_decode_available,
            nvdec_cuda_init_ok = self.nvdec_cuda_init_ok,
            video_hw_backend = self.probe.inferred_hw_backend,
            video_acceleration_effective = mode_label(self.effective),
            decode_backend = self.planned_backend.log_label(),
            hardware_decode_active = self.hardware_decode_active(),
            "video acceleration initialized (Fase 4.1)"
        );
    }

    pub fn video_acceleration_requested(&self) -> &'static str {
        mode_label(self.policy.mode)
    }

    pub fn video_acceleration_effective(&self) -> &'static str {
        mode_label(self.effective)
    }

    pub fn video_hw_backend(&self) -> &'static str {
        if self.hardware_decode_active.load(Ordering::Relaxed) {
            "nvdec"
        } else {
            self.probe.inferred_hw_backend
        }
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

    pub fn frames_hw_decoded(&self) -> u64 {
        self.frames_hw_decoded.load(Ordering::Relaxed)
    }

    pub fn frames_cpu_decoded(&self) -> u64 {
        self.frames_cpu_decoded.load(Ordering::Relaxed)
    }

    pub fn last_hw_decode_ms(&self) -> u64 {
        self.last_hw_decode_ms.load(Ordering::Relaxed)
    }

    pub fn hw_transfer_to_cpu_ms(&self) -> u64 {
        self.hw_transfer_to_cpu_ms.load(Ordering::Relaxed)
    }

    pub fn record_hw_decode_error(&self) {
        self.hw_decode_errors.fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_hw_fallback_to_cpu(&self) {
        self.hw_fallback_to_cpu_count
            .fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_frame_decoded_cpu(&self) {
        self.frames_cpu_decoded.fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_frame_decoded_nvdec(&self, decode_ms: u64, transfer_ms: u64) {
        self.frames_hw_decoded.fetch_add(1, Ordering::Relaxed);
        self.hardware_decode_active.store(true, Ordering::Relaxed);
        self.last_hw_decode_ms.store(decode_ms, Ordering::Relaxed);
        self.hw_transfer_to_cpu_ms
            .store(transfer_ms, Ordering::Relaxed);
    }

    pub fn policy_forbids_cpu_fallback(&self) -> bool {
        self.policy.mode == VideoAccelerationMode::Gpu
    }
}

fn nvdec_ready_for_policy(
    gpu_backend: VideoGpuBackend,
    probe: &ProbeResult,
    nvdec_cuda_init_ok: bool,
    cuvid: bool,
) -> bool {
    #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
    {
        let nvidia_path = matches!(gpu_backend, VideoGpuBackend::Auto | VideoGpuBackend::Cuda);
        if !nvidia_path {
            return false;
        }
        probe.gpu_detected && probe.ffmpeg_hw_decode_available && cuvid && nvdec_cuda_init_ok
    }
    #[cfg(not(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec")))]
    {
        let _ = (gpu_backend, probe, nvdec_cuda_init_ok, cuvid);
        false
    }
}

fn resolve_effective_and_backend(
    mode: VideoAccelerationMode,
    _gpu_backend: VideoGpuBackend,
    nvdec_stack_ready: bool,
) -> AppResult<(VideoAccelerationMode, PlannedDecodeBackend)> {
    match mode {
        VideoAccelerationMode::Cpu => Ok((VideoAccelerationMode::Cpu, PlannedDecodeBackend::Cpu)),
        VideoAccelerationMode::Auto => {
            #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
            if nvdec_stack_ready {
                return Ok((VideoAccelerationMode::Gpu, PlannedDecodeBackend::Nvdec));
            }
            Ok((VideoAccelerationMode::Cpu, PlannedDecodeBackend::Cpu))
        }
        VideoAccelerationMode::Gpu => {
            #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
            if nvdec_stack_ready {
                return Ok((VideoAccelerationMode::Gpu, PlannedDecodeBackend::Nvdec));
            }
            let _ = nvdec_stack_ready;
            Err(gpu_unavailable_error())
        }
    }
}

fn gpu_unavailable_error() -> AppError {
    #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
    {
        AppError::Config(
            "VIDEO_ACCELERATION=gpu exige NVIDIA NVDEC (GPU detectada, FFmpeg h264_cuvid e dispositivo CUDA inicializável); indisponível neste ambiente"
                .into(),
        )
    }
    #[cfg(not(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec")))]
    {
        AppError::Config(
            "VIDEO_ACCELERATION=gpu exige build com features ffmpeg-decode,ffmpeg-nvdec e runtime NVIDIA NVDEC"
                .into(),
        )
    }
}

#[cfg(test)]
mod tests {
    use super::super::policy::VideoGpuBackend;
    use super::super::probe::MockProbe;
    use super::*;

    fn mock_probe(nvidia: bool, ffmpeg_hw: bool, cuvid: bool, cuda_init: bool) -> MockProbe {
        MockProbe {
            nvidia,
            vaapi: false,
            qsv: false,
            ffmpeg_hw,
            h264_cuvid: cuvid,
            nvdec_cuda_init_ok: cuda_init,
        }
    }

    #[test]
    fn auto_without_gpu_effective_cpu() {
        let policy =
            AccelerationPolicy::from_parts(VideoAccelerationMode::Auto, VideoGpuBackend::Auto);
        let probe = mock_probe(false, false, false, false);
        let rt = AccelerationRuntime::bootstrap_with_probe(policy, &probe).unwrap();
        assert_eq!(rt.effective_mode(), VideoAccelerationMode::Cpu);
        assert_eq!(rt.planned_decode_backend(), PlannedDecodeBackend::Cpu);
        assert!(!rt.gpu_detected());
    }

    #[test]
    fn cpu_forces_effective_cpu_even_with_gpu() {
        let policy =
            AccelerationPolicy::from_parts(VideoAccelerationMode::Cpu, VideoGpuBackend::Cuda);
        let probe = mock_probe(true, true, true, true);
        let rt = AccelerationRuntime::bootstrap_with_probe(policy, &probe).unwrap();
        assert_eq!(rt.effective_mode(), VideoAccelerationMode::Cpu);
        assert_eq!(rt.planned_decode_backend(), PlannedDecodeBackend::Cpu);
    }

    #[test]
    fn gpu_mode_errors_without_silent_fallback() {
        let policy =
            AccelerationPolicy::from_parts(VideoAccelerationMode::Gpu, VideoGpuBackend::Auto);
        let probe = mock_probe(false, false, false, false);
        let err = AccelerationRuntime::bootstrap_with_probe(policy, &probe);
        assert!(err.is_err());
        let msg = err.unwrap_err().to_string();
        assert!(msg.contains("gpu") || msg.contains("GPU") || msg.contains("NVDEC"));
    }

    #[test]
    fn gpu_mode_errors_even_with_gpu_but_no_cuda_init() {
        let policy =
            AccelerationPolicy::from_parts(VideoAccelerationMode::Gpu, VideoGpuBackend::Auto);
        let probe = mock_probe(true, true, true, false);
        assert!(AccelerationRuntime::bootstrap_with_probe(policy, &probe).is_err());
    }

    #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
    #[test]
    fn auto_with_full_nvdec_stack_plans_nvdec() {
        let policy =
            AccelerationPolicy::from_parts(VideoAccelerationMode::Auto, VideoGpuBackend::Auto);
        let probe = mock_probe(true, true, true, true);
        let rt = AccelerationRuntime::bootstrap_with_probe(policy, &probe).unwrap();
        assert_eq!(rt.effective_mode(), VideoAccelerationMode::Gpu);
        assert_eq!(rt.planned_decode_backend(), PlannedDecodeBackend::Nvdec);
    }

    #[cfg(not(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec")))]
    #[test]
    fn auto_with_gpu_still_cpu_without_nvdec_feature() {
        let policy =
            AccelerationPolicy::from_parts(VideoAccelerationMode::Auto, VideoGpuBackend::Auto);
        let probe = mock_probe(true, true, true, true);
        let rt = AccelerationRuntime::bootstrap_with_probe(policy, &probe).unwrap();
        assert_eq!(rt.effective_mode(), VideoAccelerationMode::Cpu);
        assert_eq!(rt.planned_decode_backend(), PlannedDecodeBackend::Cpu);
    }
}
