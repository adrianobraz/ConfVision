pub use crate::decode::{DECODED_LUMA_HEIGHT as LUMA_HEIGHT, DECODED_LUMA_WIDTH as LUMA_WIDTH};

pub const LUMA_PIXELS: usize = (LUMA_WIDTH as usize) * (LUMA_HEIGHT as usize);

/// Diferença mínima |atual − referência| para contar pixel “alterado”.
pub const PIXEL_DIFF_THRESHOLD: u8 = 8;

/// Percentual mínimo de pixels alterados para declarar movimento (5%) — default testes.
pub const MOTION_PERCENT_THRESHOLD: u32 = 5;

/// Limiares configuráveis (env `MOTION_*_THRESHOLD`).
#[derive(Debug, Clone, Copy)]
pub struct MotionSensitivity {
    pub pixel_diff_threshold: u8,
    pub motion_percent_threshold: u32,
}

impl Default for MotionSensitivity {
    fn default() -> Self {
        Self {
            pixel_diff_threshold: PIXEL_DIFF_THRESHOLD,
            motion_percent_threshold: MOTION_PERCENT_THRESHOLD,
        }
    }
}

impl MotionSensitivity {
    pub fn from_config(cfg: &crate::config::Config) -> Self {
        Self {
            pixel_diff_threshold: cfg.motion_pixel_diff_threshold,
            motion_percent_threshold: cfg.motion_percent_threshold,
        }
    }
}

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
