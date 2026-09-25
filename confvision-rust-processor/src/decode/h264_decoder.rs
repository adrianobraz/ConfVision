use std::sync::Arc;
#[cfg(feature = "ffmpeg-decode")]
use std::time::Instant;

#[cfg(feature = "ffmpeg-decode")]
use tracing::info;

use super::acceleration::AccelerationRuntime;
#[cfg(feature = "ffmpeg-decode")]
use super::acceleration::PlannedDecodeBackend;
use super::context::SessionDecodeContext;
use super::decode_policy::{DecodePolicyCoordinator, H264DecoderSessionState};
use super::error::DecodeError;
use super::types::{DecodeInput, DecodeOutcome, DecodedFrame};

/// Decoder H.264 por sessão RTSP (uma instância por consumer / reconexão).
pub struct H264Decoder {
    acceleration: Option<Arc<AccelerationRuntime>>,
    decode_policy: Option<Arc<DecodePolicyCoordinator>>,
    session_state: H264DecoderSessionState,
    #[cfg(feature = "ffmpeg-decode")]
    inner: Option<BackendInstance>,
    #[cfg(feature = "ffmpeg-decode")]
    active_kind: super::backend::DecodeBackendKind,
    #[cfg(feature = "ffmpeg-decode")]
    nvdec_startup_logged: bool,
}

#[cfg(feature = "ffmpeg-decode")]
enum BackendInstance {
    Cpu(super::backend::CpuFfmpegDecoder),
    #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
    Nvdec(super::backend::NvdecH264Decoder),
}

impl H264Decoder {
    pub fn new() -> Self {
        Self::from_parts(None, None)
    }

    pub fn with_acceleration(acceleration: Arc<AccelerationRuntime>) -> Self {
        Self::from_parts(Some(acceleration), None)
    }

    pub fn with_acceleration_and_policy(
        acceleration: Arc<AccelerationRuntime>,
        decode_policy: Arc<DecodePolicyCoordinator>,
    ) -> Self {
        Self::from_parts(Some(acceleration), Some(decode_policy))
    }

    fn from_parts(
        acceleration: Option<Arc<AccelerationRuntime>>,
        decode_policy: Option<Arc<DecodePolicyCoordinator>>,
    ) -> Self {
        Self {
            acceleration,
            decode_policy,
            session_state: H264DecoderSessionState::new(),
            #[cfg(feature = "ffmpeg-decode")]
            inner: None,
            #[cfg(feature = "ffmpeg-decode")]
            active_kind: super::backend::DecodeBackendKind::Cpu,
            #[cfg(feature = "ffmpeg-decode")]
            nvdec_startup_logged: false,
        }
    }

    /// Decodifica um access unit. Não propaga panic; erros viram `DecodeOutcome::Failed`.
    pub fn decode(
        &mut self,
        decode_ctx: &SessionDecodeContext,
        input: DecodeInput<'_>,
    ) -> DecodeOutcome {
        #[cfg(feature = "ffmpeg-decode")]
        {
            return self.decode_with_ffmpeg(decode_ctx, input);
        }
        #[cfg(not(feature = "ffmpeg-decode"))]
        {
            let _ = (self, decode_ctx, input);
            DecodeOutcome::NotReady
        }
    }
}

impl Default for H264Decoder {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(feature = "ffmpeg-decode")]
impl H264Decoder {
    fn decode_with_ffmpeg(
        &mut self,
        decode_ctx: &SessionDecodeContext,
        input: DecodeInput<'_>,
    ) -> DecodeOutcome {
        let Some(extradata) = decode_ctx.extradata() else {
            return DecodeOutcome::NotReady;
        };
        let generation = decode_ctx.generation();

        let needs_new = self
            .inner
            .as_ref()
            .map(|d| backend_needs_reinit(d, generation))
            .unwrap_or(true);

        if needs_new {
            match self.create_backend(&extradata, generation) {
                Ok(()) => {}
                Err(e) => return DecodeOutcome::Failed(e),
            }
        }

        let started = Instant::now();
        let decode_result = self.decode_access_unit(input.payload, input.is_keyframe);
        let decode_ms = started.elapsed().as_millis() as u64;

        match decode_result {
            Ok(frame) => {
                if self.active_kind != super::backend::DecodeBackendKind::Cpu {
                    self.session_state.record_hw_success();
                }
                if let Some(accel) = &self.acceleration {
                    match self.active_kind {
                        super::backend::DecodeBackendKind::Cpu => {
                            accel.record_frame_decoded_cpu();
                        }
                        #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
                        super::backend::DecodeBackendKind::Nvdec => {
                            let transfer_ms = self.nvdec_transfer_ms();
                            accel.record_frame_decoded_nvdec(decode_ms, transfer_ms);
                            if !self.nvdec_startup_logged {
                                info!(backend = "nvdec", "session hardware decode active");
                                self.nvdec_startup_logged = true;
                            }
                        }
                    }
                }
                DecodeOutcome::Decoded(frame)
            }
            Err(e) if super::backend::is_eagain(&e) => DecodeOutcome::NotReady,
            Err(e) => {
                if self.active_kind != super::backend::DecodeBackendKind::Cpu {
                    if let Some(accel) = &self.acceleration {
                        accel.record_hw_decode_error();
                    }
                    if self.try_runtime_fallback_to_cpu(decode_ctx, &extradata, generation) {
                        return DecodeOutcome::NotReady;
                    }
                }
                DecodeOutcome::Failed(e)
            }
        }
    }

    fn try_runtime_fallback_to_cpu(
        &mut self,
        decode_ctx: &SessionDecodeContext,
        extradata: &[u8],
        generation: u64,
    ) -> bool {
        let Some(policy) = self.decode_policy.as_ref() else {
            return false;
        };
        let Some(accel) = self.acceleration.as_ref() else {
            return false;
        };
        let failures = self.session_state.record_hw_failure();
        if failures < policy.hw_error_threshold() {
            return false;
        }
        if !policy.trigger_runtime_cpu_fallback(accel) {
            return false;
        }
        info!(
            consecutive_hw_errors = failures,
            threshold = policy.hw_error_threshold(),
            "runtime NVDEC→CPU fallback (same RTSP session)"
        );
        match super::backend::CpuFfmpegDecoder::try_new(extradata, generation) {
            Ok(cpu) => {
                self.inner = Some(BackendInstance::Cpu(cpu));
                self.active_kind = super::backend::DecodeBackendKind::Cpu;
                self.nvdec_startup_logged = false;
                self.session_state.reset();
                let _ = decode_ctx;
                true
            }
            Err(_) => false,
        }
    }

    fn create_backend(&mut self, extradata: &[u8], generation: u64) -> Result<(), DecodeError> {
        let planned = self
            .acceleration
            .as_ref()
            .map(|a| a.planned_decode_backend())
            .unwrap_or(PlannedDecodeBackend::Cpu);

        let attempt_nvdec = matches!(planned, PlannedDecodeBackend::Nvdec)
            && self
                .decode_policy
                .as_ref()
                .map(|p| {
                    self.acceleration
                        .as_ref()
                        .map(|a| p.should_attempt_nvdec(a))
                        .unwrap_or(true)
                })
                .unwrap_or(true);

        #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
        if attempt_nvdec {
            match super::backend::NvdecH264Decoder::try_new(extradata, generation) {
                Ok(d) => {
                    self.inner = Some(BackendInstance::Nvdec(d));
                    self.active_kind = super::backend::DecodeBackendKind::Nvdec;
                    return Ok(());
                }
                Err(e) => {
                    if let Some(accel) = &self.acceleration {
                        accel.record_hw_decode_error();
                        if accel.policy_forbids_cpu_fallback() {
                            return Err(e);
                        }
                        accel.record_hw_fallback_to_cpu();
                    }
                }
            }
        }

        let cpu = super::backend::CpuFfmpegDecoder::try_new(extradata, generation)?;
        self.inner = Some(BackendInstance::Cpu(cpu));
        self.active_kind = super::backend::DecodeBackendKind::Cpu;
        Ok(())
    }

    fn decode_access_unit(
        &mut self,
        payload: &[u8],
        is_keyframe: bool,
    ) -> Result<DecodedFrame, DecodeError> {
        use super::backend::VideoDecodeBackend;
        match self.inner.as_mut() {
            Some(BackendInstance::Cpu(d)) => d.decode_access_unit(payload, is_keyframe),
            #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
            Some(BackendInstance::Nvdec(d)) => d.decode_access_unit(payload, is_keyframe),
            None => Err(DecodeError::Ffmpeg("decoder não inicializado".into())),
        }
    }

    #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
    fn nvdec_transfer_ms(&self) -> u64 {
        match self.inner.as_ref() {
            Some(BackendInstance::Nvdec(d)) => d.last_transfer_to_cpu_ms,
            _ => 0,
        }
    }
}

#[cfg(feature = "ffmpeg-decode")]
fn backend_needs_reinit(instance: &BackendInstance, generation: u64) -> bool {
    use super::backend::VideoDecodeBackend;
    match instance {
        BackendInstance::Cpu(d) => d.needs_reinit(generation),
        #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
        BackendInstance::Nvdec(d) => d.needs_reinit(generation),
    }
}

#[cfg(all(test, feature = "ffmpeg-decode"))]
mod fallback_tests {
    use super::*;
    use crate::decode::acceleration::MockProbe;
    use crate::decode::acceleration::{AccelerationPolicy, VideoAccelerationMode, VideoGpuBackend};
    use crate::decode::DecodeFallbackConfig;

    fn mock_probe() -> MockProbe {
        MockProbe {
            nvidia: true,
            vaapi: false,
            qsv: false,
            ffmpeg_hw: true,
            h264_cuvid: true,
            nvdec_cuda_init_ok: true,
        }
    }

    #[test]
    fn gpu_mode_does_not_allow_runtime_fallback_trigger() {
        let accel = AccelerationRuntime::bootstrap_with_probe(
            AccelerationPolicy::from_parts(VideoAccelerationMode::Gpu, VideoGpuBackend::Auto),
            &mock_probe(),
        )
        .unwrap();
        let policy = DecodePolicyCoordinator::new(DecodeFallbackConfig {
            runtime_fallback_enabled: true,
            hw_error_threshold: 1,
        });
        assert!(!policy.runtime_fallback_allowed(accel.as_ref()));
    }

    #[test]
    fn fallback_disabled_blocks_trigger() {
        let accel = AccelerationRuntime::bootstrap_with_probe(
            AccelerationPolicy::from_parts(VideoAccelerationMode::Auto, VideoGpuBackend::Auto),
            &mock_probe(),
        )
        .unwrap();
        let policy = DecodePolicyCoordinator::new(DecodeFallbackConfig {
            runtime_fallback_enabled: false,
            hw_error_threshold: 1,
        });
        assert!(!policy.runtime_fallback_allowed(accel.as_ref()));
    }

    #[test]
    fn auto_with_fallback_enabled_allows_once() {
        let accel = AccelerationRuntime::bootstrap_with_probe(
            AccelerationPolicy::from_parts(VideoAccelerationMode::Auto, VideoGpuBackend::Auto),
            &mock_probe(),
        )
        .unwrap();
        let policy = DecodePolicyCoordinator::new(DecodeFallbackConfig {
            runtime_fallback_enabled: true,
            hw_error_threshold: 10,
        });
        assert!(policy.runtime_fallback_allowed(accel.as_ref()));
        assert!(policy.trigger_runtime_cpu_fallback(accel.as_ref()));
        assert!(!policy.runtime_fallback_allowed(accel.as_ref()));
        assert!(!policy.should_attempt_nvdec(accel.as_ref()));
    }
}
