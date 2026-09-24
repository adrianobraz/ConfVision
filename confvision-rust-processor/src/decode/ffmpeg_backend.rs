//! Compat — reexport do backend CPU (Fase 4.1).
pub use super::backend::cpu::{is_eagain, map_ffmpeg_err, CpuFfmpegDecoder as FfmpegH264Decoder};
