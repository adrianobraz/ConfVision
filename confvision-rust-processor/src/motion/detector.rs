use std::sync::Arc;

use crate::decode::DecodedFrame;

use super::types::{
    MotionOutcome, MotionSensitivity, LUMA_HEIGHT, LUMA_PIXELS, LUMA_WIDTH,
    MOTION_PERCENT_THRESHOLD, PIXEL_DIFF_THRESHOLD,
};

/// Detector com referência de **cenário** (lenta) vs **movimento** (rápido).
#[derive(Debug)]
pub struct MotionDetector {
    scene: Option<Arc<[u8]>>,
    sensitivity: MotionSensitivity,
}

impl Default for MotionDetector {
    fn default() -> Self {
        Self {
            scene: None,
            sensitivity: MotionSensitivity::default(),
        }
    }
}

impl MotionDetector {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn with_sensitivity(sensitivity: MotionSensitivity) -> Self {
        Self {
            scene: None,
            sensitivity,
        }
    }

    fn diff_score_percent(luma: &[u8], reference: &[u8], pixel_threshold: u8) -> u32 {
        let mut changed = 0u32;
        for (a, b) in luma.iter().zip(reference.iter()) {
            if a.abs_diff(*b) >= pixel_threshold {
                changed += 1;
            }
        }
        (changed * 100) / LUMA_PIXELS as u32
    }

    /// Analisa o frame decodificado contra referência de cenário.
    pub fn analyze(&mut self, frame: &DecodedFrame) -> MotionOutcome {
        let luma = &frame.luma;
        if luma.len() != LUMA_PIXELS
            || frame.luma_width != LUMA_WIDTH
            || frame.luma_height != LUMA_HEIGHT
        {
            return MotionOutcome::Error;
        }

        let Some(scene) = self.scene.as_ref() else {
            self.scene = Some(Arc::clone(luma));
            return MotionOutcome::ReferenceSet;
        };

        if scene.len() != LUMA_PIXELS {
            self.scene = Some(Arc::clone(luma));
            return MotionOutcome::ReferenceSet;
        }

        let score_percent = Self::diff_score_percent(
            luma,
            scene,
            self.sensitivity.pixel_diff_threshold,
        );

        if score_percent >= self.sensitivity.scene_shift_percent {
            self.scene = Some(Arc::clone(luma));
            return MotionOutcome::SceneUpdated;
        }

        let detected = score_percent >= self.sensitivity.motion_percent_threshold;

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
        let mut det = MotionDetector::with_sensitivity(MotionSensitivity {
            pixel_diff_threshold: PIXEL_DIFF_THRESHOLD,
            motion_percent_threshold: 5,
            scene_shift_percent: 101,
        });
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
    fn full_scene_change_updates_without_motion_alarm() {
        let mut det = MotionDetector::with_sensitivity(MotionSensitivity {
            pixel_diff_threshold: PIXEL_DIFF_THRESHOLD,
            motion_percent_threshold: 5,
            scene_shift_percent: 50,
        });
        assert_eq!(det.analyze(&solid(10)), MotionOutcome::ReferenceSet);
        let out = det.analyze(&solid(200));
        assert_eq!(out, MotionOutcome::SceneUpdated);
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
