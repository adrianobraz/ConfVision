//! Facade de decode (Fase 6.2): stack GPU, backend efetivo e fallback runtime.

use std::sync::atomic::{AtomicBool, AtomicU32, Ordering};
use std::sync::Arc;

use serde::Serialize;

use super::acceleration::{AccelerationRuntime, PlannedDecodeBackend, VideoAccelerationMode};
use crate::capacity::{CapacitySnapshot, CapacityState, LimitingResource};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum HwStackGrade {
    None,
    Partial,
    Ready,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum GpuDecodeState {
    Unavailable,
    StackReady,
    Active,
    Saturated,
    CpuFallback,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum DecodeBackendEffective {
    Cpu,
    Nvdec,
}

impl DecodeBackendEffective {
    pub fn as_str(self) -> &'static str {
        match self {
            Self::Cpu => "cpu",
            Self::Nvdec => "nvdec",
        }
    }
}

#[derive(Debug, Clone)]
pub struct DecodeFallbackConfig {
    pub runtime_fallback_enabled: bool,
    pub hw_error_threshold: u32,
}

impl DecodeFallbackConfig {
    pub fn from_env_parts(runtime_fallback: bool, hw_error_threshold: u32) -> Self {
        Self {
            runtime_fallback_enabled: runtime_fallback,
            hw_error_threshold: hw_error_threshold.max(1),
        }
    }
}

/// Estado compartilhado do worker (fallback runtime NVDEC → CPU, uma via por processo).
#[derive(Debug)]
pub struct DecodePolicyCoordinator {
    fallback: DecodeFallbackConfig,
    runtime_cpu_fallback: AtomicBool,
}

impl DecodePolicyCoordinator {
    pub fn new(fallback: DecodeFallbackConfig) -> Arc<Self> {
        Arc::new(Self {
            fallback,
            runtime_cpu_fallback: AtomicBool::new(false),
        })
    }

    pub fn runtime_cpu_fallback(&self) -> bool {
        self.runtime_cpu_fallback.load(Ordering::Relaxed)
    }

    pub fn hw_error_threshold(&self) -> u32 {
        self.fallback.hw_error_threshold
    }

    pub fn runtime_fallback_allowed(&self, accel: &AccelerationRuntime) -> bool {
        self.fallback.runtime_fallback_enabled
            && accel.policy().mode == VideoAccelerationMode::Auto
            && !self.runtime_cpu_fallback()
    }

    /// Marca fallback global; retorna true se transição ocorreu.
    pub fn trigger_runtime_cpu_fallback(&self, accel: &AccelerationRuntime) -> bool {
        if !self.runtime_fallback_allowed(accel) {
            return false;
        }
        if self
            .runtime_cpu_fallback
            .compare_exchange(false, true, Ordering::SeqCst, Ordering::Relaxed)
            .is_ok()
        {
            accel.record_hw_fallback_to_cpu();
            true
        } else {
            false
        }
    }

    pub fn should_attempt_nvdec(&self, accel: &AccelerationRuntime) -> bool {
        if self.runtime_cpu_fallback() {
            return false;
        }
        nvdec_planned(accel)
    }
}

pub fn evaluate_hw_stack_grade(accel: &AccelerationRuntime) -> HwStackGrade {
    if nvdec_planned(accel) {
        return HwStackGrade::Ready;
    }
    if !accel.gpu_detected()
        && !accel.ffmpeg_hw_decode_available()
        && !accel.nvdec_cuda_init_ok()
    {
        return HwStackGrade::None;
    }
    if accel.gpu_detected()
        || accel.ffmpeg_hw_decode_available()
        || accel.nvdec_cuda_init_ok()
    {
        return HwStackGrade::Partial;
    }
    HwStackGrade::None
}

pub fn decode_backend_effective(
    accel: &AccelerationRuntime,
    coord: &DecodePolicyCoordinator,
) -> DecodeBackendEffective {
    if coord.runtime_cpu_fallback() {
        return DecodeBackendEffective::Cpu;
    }
    #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
    if accel.hardware_decode_active() || accel.frames_hw_decoded() > 0 {
        return DecodeBackendEffective::Nvdec;
    }
    if nvdec_planned(accel) && !coord.runtime_cpu_fallback() {
        return DecodeBackendEffective::Nvdec;
    }
    DecodeBackendEffective::Cpu
}

fn nvdec_planned(accel: &AccelerationRuntime) -> bool {
    #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
    {
        matches!(accel.planned_decode_backend(), PlannedDecodeBackend::Nvdec)
    }
    #[cfg(not(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec")))]
    {
        let _ = accel;
        false
    }
}

pub fn gpu_decode_state(
    accel: &AccelerationRuntime,
    coord: &DecodePolicyCoordinator,
    capacity: Option<&CapacitySnapshot>,
) -> GpuDecodeState {
    if coord.runtime_cpu_fallback() {
        return GpuDecodeState::CpuFallback;
    }

    let grade = evaluate_hw_stack_grade(accel);
    if grade == HwStackGrade::None {
        return GpuDecodeState::Unavailable;
    }

    if accel.hardware_decode_active() || accel.frames_hw_decoded() > 0 {
        if let Some(cap) = capacity {
            if cap.limiting_resource == LimitingResource::Gpu
                && matches!(cap.state, CapacityState::Critical | CapacityState::Warning)
            {
                return GpuDecodeState::Saturated;
            }
        }
        return GpuDecodeState::Active;
    }

    if grade == HwStackGrade::Ready {
        return GpuDecodeState::StackReady;
    }

    GpuDecodeState::Unavailable
}

#[derive(Debug)]
pub struct H264DecoderSessionState {
    consecutive_hw_errors: AtomicU32,
}

impl H264DecoderSessionState {
    pub fn new() -> Self {
        Self {
            consecutive_hw_errors: AtomicU32::new(0),
        }
    }

    pub fn record_hw_success(&self) {
        self.consecutive_hw_errors.store(0, Ordering::Relaxed);
    }

    pub fn record_hw_failure(&self) -> u32 {
        self.consecutive_hw_errors.fetch_add(1, Ordering::Relaxed) + 1
    }

    pub fn reset(&self) {
        self.consecutive_hw_errors.store(0, Ordering::Relaxed);
    }
}

impl Default for H264DecoderSessionState {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::decode::acceleration::MockProbe;
    use crate::decode::acceleration::{
        AccelerationPolicy, AccelerationRuntime, VideoAccelerationMode, VideoGpuBackend,
    };

    fn mock_probe(nvidia: bool, ffmpeg_hw: bool, cuvid: bool, cuda_init: bool) -> MockProbe {
        MockProbe {
            nvidia,
            vaapi: false,
            qsv: false,
            ffmpeg_hw: ffmpeg_hw,
            h264_cuvid: cuvid,
            nvdec_cuda_init_ok: cuda_init,
        }
    }

    fn coord() -> Arc<DecodePolicyCoordinator> {
        DecodePolicyCoordinator::new(DecodeFallbackConfig {
            runtime_fallback_enabled: true,
            hw_error_threshold: 3,
        })
    }

    #[test]
    fn hw_grade_none_without_gpu() {
        let rt = AccelerationRuntime::bootstrap_with_probe(
            AccelerationPolicy::from_parts(VideoAccelerationMode::Auto, VideoGpuBackend::Auto),
            &mock_probe(false, false, false, false),
        )
        .unwrap();
        assert_eq!(evaluate_hw_stack_grade(&rt), HwStackGrade::None);
    }

    #[test]
    fn hw_grade_partial_device_only() {
        let rt = AccelerationRuntime::bootstrap_with_probe(
            AccelerationPolicy::from_parts(VideoAccelerationMode::Auto, VideoGpuBackend::Auto),
            &mock_probe(true, false, false, false),
        )
        .unwrap();
        assert_eq!(evaluate_hw_stack_grade(&rt), HwStackGrade::Partial);
    }

    #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
    #[test]
    fn hw_grade_ready_full_stack() {
        let rt = AccelerationRuntime::bootstrap_with_probe(
            AccelerationPolicy::from_parts(VideoAccelerationMode::Auto, VideoGpuBackend::Auto),
            &mock_probe(true, true, true, true),
        )
        .unwrap();
        assert_eq!(evaluate_hw_stack_grade(&rt), HwStackGrade::Ready);
    }

    #[test]
    fn cpu_mode_effective_backend_cpu() {
        let rt = AccelerationRuntime::bootstrap_with_probe(
            AccelerationPolicy::from_parts(VideoAccelerationMode::Cpu, VideoGpuBackend::Auto),
            &mock_probe(true, true, true, true),
        )
        .unwrap();
        let c = coord();
        assert_eq!(
            decode_backend_effective(&rt, c.as_ref()),
            DecodeBackendEffective::Cpu
        );
    }

    #[test]
    fn cpu_fallback_state_after_trigger() {
        let rt = AccelerationRuntime::bootstrap_with_probe(
            AccelerationPolicy::from_parts(VideoAccelerationMode::Auto, VideoGpuBackend::Auto),
            &mock_probe(true, true, true, true),
        )
        .unwrap();
        let c = coord();
        assert!(c.trigger_runtime_cpu_fallback(&rt));
        assert_eq!(
            gpu_decode_state(&rt, c.as_ref(), None),
            GpuDecodeState::CpuFallback
        );
    }

    #[test]
    fn session_hw_error_counter_threshold() {
        let s = H264DecoderSessionState::new();
        assert_eq!(s.record_hw_failure(), 1);
        assert_eq!(s.record_hw_failure(), 2);
        s.record_hw_success();
        assert_eq!(s.record_hw_failure(), 1);
    }

    #[test]
    fn auto_without_stack_plans_cpu_unavailable_state() {
        let rt = AccelerationRuntime::bootstrap_with_probe(
            AccelerationPolicy::from_parts(VideoAccelerationMode::Auto, VideoGpuBackend::Auto),
            &mock_probe(false, false, false, false),
        )
        .unwrap();
        let c = coord();
        assert_eq!(
            gpu_decode_state(&rt, c.as_ref(), None),
            GpuDecodeState::Unavailable
        );
    }

    #[test]
    fn gpu_unavailable_at_bootstrap_is_error() {
        let err = AccelerationRuntime::bootstrap_with_probe(
            AccelerationPolicy::from_parts(VideoAccelerationMode::Gpu, VideoGpuBackend::Auto),
            &mock_probe(false, false, false, false),
        );
        assert!(err.is_err());
    }

    #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
    #[test]
    fn stack_ready_without_active_frames() {
        let rt = AccelerationRuntime::bootstrap_with_probe(
            AccelerationPolicy::from_parts(VideoAccelerationMode::Auto, VideoGpuBackend::Auto),
            &mock_probe(true, true, true, true),
        )
        .unwrap();
        let c = coord();
        assert_eq!(
            gpu_decode_state(&rt, c.as_ref(), None),
            GpuDecodeState::StackReady
        );
    }

    #[test]
    fn saturated_when_gpu_limiting_and_warning() {
        use crate::capacity::{
            CapacitySnapshot, CapacityState, LimitingResource, NetworkSnapshot, ResourceMetric,
            ResourceScope, StorageSnapshot,
        };
        use crate::config::CapacityMode;

        let rt = AccelerationRuntime::bootstrap_with_probe(
            AccelerationPolicy::from_parts(VideoAccelerationMode::Auto, VideoGpuBackend::Auto),
            &mock_probe(true, true, true, true),
        )
        .unwrap();
        rt.record_frame_decoded_nvdec(1, 1);
        let m = ResourceMetric {
            percent: None,
            target_percent: 80.0,
            headroom_percent: None,
            scope: ResourceScope::Unavailable,
        };
        let cap = CapacitySnapshot {
            mode: CapacityMode::Dynamic,
            state: CapacityState::Warning,
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
            current_cameras: 1,
            current_online_cameras: 1,
            current_fps: 5.0,
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
            limiting_resource: LimitingResource::Gpu,
            max_cameras_safety_limit: None,
            cameras: vec![],
            observation_ready: true,
            samples_in_window: 1,
        };
        let c = coord();
        assert_eq!(
            gpu_decode_state(&rt, c.as_ref(), Some(&cap)),
            GpuDecodeState::Saturated
        );
    }
}
