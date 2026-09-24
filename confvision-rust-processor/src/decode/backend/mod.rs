mod cpu;

#[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
mod nvdec;

pub use cpu::{is_eagain, map_ffmpeg_err, CpuFfmpegDecoder};

#[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
pub use nvdec::{NvdecCapability, NvdecH264Decoder};

use super::error::DecodeError;
use super::types::DecodedFrame;

/// Backend de decode de vídeo (CPU ou HW).
pub trait VideoDecodeBackend {
    fn needs_reinit(&self, extradata_generation: u64) -> bool;
    fn decode_access_unit(
        &mut self,
        data: &[u8],
        is_keyframe: bool,
    ) -> Result<DecodedFrame, DecodeError>;
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum DecodeBackendKind {
    Cpu,
    #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
    Nvdec,
}

impl DecodeBackendKind {
    pub fn metric_label(self) -> &'static str {
        match self {
            Self::Cpu => "cpu",
            #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
            Self::Nvdec => "nvdec",
        }
    }
}

#[cfg(test)]
mod tests;
