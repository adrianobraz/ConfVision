use std::sync::Arc;

use super::acceleration::AccelerationRuntime;
use super::context::SessionDecodeContext;
use super::error::DecodeError;
use super::types::{DecodeInput, DecodeOutcome, DecodedFrame};

/// Decoder H.264 por sessão RTSP (uma instância por consumer / reconexão).
pub struct H264Decoder {
    #[allow(dead_code)]
    acceleration: Option<Arc<AccelerationRuntime>>,
    #[cfg(feature = "ffmpeg-decode")]
    inner: Option<super::ffmpeg_backend::FfmpegH264Decoder>,
}

impl H264Decoder {
    pub fn new() -> Self {
        Self::from_acceleration(None)
    }

    /// Fase 4.0+: policy/runtime para futura seleção CPU/HW (decode continua CPU).
    pub fn with_acceleration(acceleration: Arc<AccelerationRuntime>) -> Self {
        Self::from_acceleration(Some(acceleration))
    }

    fn from_acceleration(acceleration: Option<Arc<AccelerationRuntime>>) -> Self {
        Self {
            acceleration,
            #[cfg(feature = "ffmpeg-decode")]
            inner: None,
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
            .map(|d| d.needs_reinit(generation))
            .unwrap_or(true);

        if needs_new {
            match super::ffmpeg_backend::FfmpegH264Decoder::try_new(&extradata, generation) {
                Ok(d) => self.inner = Some(d),
                Err(e) => return DecodeOutcome::Failed(e),
            }
        }

        match self
            .inner
            .as_mut()
            .unwrap()
            .decode_access_unit(input.payload, input.is_keyframe)
        {
            Ok(frame) => DecodeOutcome::Decoded(frame),
            Err(e) if super::ffmpeg_backend::is_eagain(&e) => DecodeOutcome::NotReady,
            Err(e) => DecodeOutcome::Failed(e),
        }
    }
}
