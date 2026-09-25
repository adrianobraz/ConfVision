use std::time::{Duration, Instant};

/// Limita decode+motion por câmera (Fase 6.3). RTSP continua em taxa cheia.
#[derive(Debug, Clone)]
pub struct MotionAnalysisThrottle {
    min_interval: Duration,
    last_analysis: Option<Instant>,
}

impl MotionAnalysisThrottle {
    /// `max_fps <= 0` desativa o limite (comportamento anterior).
    pub fn from_max_fps(max_fps: f64) -> Self {
        let min_interval = if max_fps > 0.0 {
            Duration::from_secs_f64(1.0 / max_fps)
        } else {
            Duration::ZERO
        };
        Self {
            min_interval,
            last_analysis: None,
        }
    }

    pub fn is_unlimited(&self) -> bool {
        self.min_interval.is_zero()
    }

    /// Keyframes sempre passam (resync H.264 após frames ignorados).
    pub fn should_analyze(&mut self, is_keyframe: bool) -> bool {
        if self.is_unlimited() || is_keyframe {
            if !self.is_unlimited() {
                self.last_analysis = Some(Instant::now());
            }
            return true;
        }
        match self.last_analysis {
            None => {
                self.last_analysis = Some(Instant::now());
                true
            }
            Some(t) if t.elapsed() >= self.min_interval => {
                self.last_analysis = Some(Instant::now());
                true
            }
            Some(_) => false,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::thread;

    #[test]
    fn unlimited_always_true() {
        let mut t = MotionAnalysisThrottle::from_max_fps(0.0);
        assert!(t.should_analyze(false));
        assert!(t.should_analyze(false));
    }

    #[test]
    fn caps_non_keyframes() {
        let mut t = MotionAnalysisThrottle::from_max_fps(10.0);
        assert!(t.should_analyze(false));
        assert!(!t.should_analyze(false));
        assert!(t.should_analyze(true));
    }

    #[test]
    fn interval_allows_next_after_wait() {
        let mut t = MotionAnalysisThrottle::from_max_fps(100.0);
        assert!(t.should_analyze(false));
        assert!(!t.should_analyze(false));
        thread::sleep(Duration::from_millis(12));
        assert!(t.should_analyze(false));
    }
}
