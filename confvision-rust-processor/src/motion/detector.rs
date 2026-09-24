use std::sync::Arc;

use crate::decode::DecodedFrame;

use super::types::{
    MotionOutcome, LUMA_HEIGHT, LUMA_PIXELS, LUMA_WIDTH, MOTION_PERCENT_THRESHOLD,
    PIXEL_DIFF_THRESHOLD,
};

/// Detector frame-a-frame sobre grade Y reduzida (stateful por sessão consumer).
#[derive(Debug, Default)]
pub struct MotionDetector {
    reference: Option<Arc<[u8]>>,
}

impl MotionDetector {
    pub fn new() -> Self {
        Self::default()
    }

    /// Analisa o frame decodificado. Primeiro frame válido só define referência.
    pub fn analyze(&mut self, frame: &DecodedFrame) -> MotionOutcome {
        let luma = &frame.luma;
        if luma.len() != LUMA_PIXELS
            || frame.luma_width != LUMA_WIDTH
            || frame.luma_height != LUMA_HEIGHT
        {
            return MotionOutcome::Error;
        }

        let current = Arc::clone(luma);

        let Some(reference) = self.reference.as_ref() else {
            self.reference = Some(current);
            return MotionOutcome::ReferenceSet;
        };

        if reference.len() != LUMA_PIXELS {
            self.reference = Some(current);
            return MotionOutcome::ReferenceSet;
        }

        let mut changed = 0u32;
        for (a, b) in current.iter().zip(reference.iter()) {
            let diff = a.abs_diff(*b);
            if diff >= PIXEL_DIFF_THRESHOLD {
                changed += 1;
            }
        }

        let score_percent = (changed * 100) / LUMA_PIXELS as u32;
        let detected = score_percent >= MOTION_PERCENT_THRESHOLD;

        self.reference = Some(current);

        MotionOutcome::Analyzed {
            detected,
            score_percent,
        }
    }
}

#[cfg(test)]
mod tests {
    use std::sync::Arc;

    use crate::decode::{DecodedFrame, PixelFormat};

    use super::*;
    use crate::motion::types::LUMA_PIXELS;

    fn frame_from_bytes(bytes: Vec<u8>) -> DecodedFrame {
        DecodedFrame {
            width: 640,
            height: 480,
            format: PixelFormat::Yuv420p,
            luma: Arc::from(bytes.into_boxed_slice()),
            luma_width: LUMA_WIDTH,
            luma_height: LUMA_HEIGHT,
        }
    }

    fn solid(value: u8) -> DecodedFrame {
        frame_from_bytes(vec![value; LUMA_PIXELS])
    }

    #[test]
    fn first_frame_sets_reference_without_motion() {
        let mut det = MotionDetector::new();
        let out = det.analyze(&solid(100));
        assert_eq!(out, MotionOutcome::ReferenceSet);
    }

    #[test]
    fn identical_second_frame_no_motion() {
        let mut det = MotionDetector::new();
        let a = solid(50);
        assert_eq!(det.analyze(&a), MotionOutcome::ReferenceSet);
        let out = det.analyze(&solid(50));
        assert!(matches!(
            out,
            MotionOutcome::Analyzed {
                detected: false,
                score_percent: 0
            }
        ));
    }

    #[test]
    fn small_diff_below_threshold_no_motion() {
        let mut det = MotionDetector::new();
        assert_eq!(det.analyze(&solid(100)), MotionOutcome::ReferenceSet);
        let out = det.analyze(&solid(107));
        assert!(matches!(
            out,
            MotionOutcome::Analyzed {
                detected: false,
                ..
            }
        ));
        if let MotionOutcome::Analyzed { score_percent, .. } = out {
            assert!(score_percent < MOTION_PERCENT_THRESHOLD);
        }
    }

    #[test]
    fn large_shift_triggers_motion() {
        let mut det = MotionDetector::new();
        assert_eq!(det.analyze(&solid(0)), MotionOutcome::ReferenceSet);
        let out = det.analyze(&solid(255));
        assert!(matches!(
            out,
            MotionOutcome::Analyzed {
                detected: true,
                score_percent: 100
            }
        ));
    }

    #[test]
    fn invalid_luma_length_is_error() {
        let mut det = MotionDetector::new();
        let bad = DecodedFrame {
            width: 640,
            height: 480,
            format: PixelFormat::Yuv420p,
            luma: Arc::from(vec![0u8; 10]),
            luma_width: LUMA_WIDTH,
            luma_height: LUMA_HEIGHT,
        };
        assert_eq!(det.analyze(&bad), MotionOutcome::Error);
    }
}
