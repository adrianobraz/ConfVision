pub use crate::decode::{DECODED_LUMA_HEIGHT as LUMA_HEIGHT, DECODED_LUMA_WIDTH as LUMA_WIDTH};

pub const LUMA_PIXELS: usize = (LUMA_WIDTH as usize) * (LUMA_HEIGHT as usize);

/// Diferença mínima |atual − referência| para contar pixel “alterado”.
pub const PIXEL_DIFF_THRESHOLD: u8 = 8;

/// Percentual mínimo de pixels alterados para declarar movimento (5%).
pub const MOTION_PERCENT_THRESHOLD: u32 = 5;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum MotionOutcome {
    /// Primeiro frame válido — referência armazenada, sem movimento.
    ReferenceSet,
    /// Comparação frame-a-frame concluída.
    Analyzed {
        detected: bool,
        /// Percentual de pixels com diff ≥ limiar (0–100).
        score_percent: u32,
    },
    /// Luma inválida ou dimensão inesperada.
    Error,
}
