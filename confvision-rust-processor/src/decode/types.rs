use std::sync::Arc;

use super::DecodeError;

/// Grade Y reduzida anexada a cada `DecodedFrame` (Fase 3.2 motion).
pub const DECODED_LUMA_WIDTH: u32 = 160;
pub const DECODED_LUMA_HEIGHT: u32 = 120;

/// Formato de pixel do frame decodificado (sem expor FFmpeg).
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum PixelFormat {
    Yuv420p,
    Nv12,
}

impl PixelFormat {
    pub fn as_str(self) -> &'static str {
        match self {
            Self::Yuv420p => "yuv420p",
            Self::Nv12 => "nv12",
        }
    }
}

/// Resultado mínimo de um decode bem-sucedido.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct DecodedFrame {
    pub width: u32,
    pub height: u32,
    pub format: PixelFormat,
    /// Plano Y reduzido (160×120) para motion / análise.
    pub luma: Arc<[u8]>,
    pub luma_width: u32,
    pub luma_height: u32,
}

/// Entrada para decode a partir de um `PipelineFrame`.
#[derive(Debug, Clone, Copy)]
pub struct DecodeInput<'a> {
    pub payload: &'a [u8],
    pub is_keyframe: bool,
    pub rtp_timestamp_ticks: Option<i64>,
}

/// Resultado da tentativa de decode de um access unit.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum DecodeOutcome {
    /// FFmpeg produziu um frame de vídeo.
    Decoded(DecodedFrame),
    /// Extradata ausente ou decoder precisa de mais dados (ex. EAGAIN).
    NotReady,
    /// Falha de decode (contabilizar `decode_errors`).
    Failed(DecodeError),
}
