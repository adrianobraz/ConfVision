use thiserror::Error;

#[derive(Debug, Clone, Error, PartialEq, Eq)]
pub enum DecodeError {
    #[error("decode FFmpeg desabilitado (feature ffmpeg-decode)")]
    Disabled,
    #[error("extradata H.264 indisponível")]
    NoExtradata,
    #[error("FFmpeg: {0}")]
    Ffmpeg(String),
    #[error("nenhum frame decodificado para este access unit")]
    NoFrame,
    #[error("formato de pixel inesperado: {0}")]
    UnsupportedPixelFormat(String),
}
