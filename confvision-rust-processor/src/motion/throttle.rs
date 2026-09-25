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

/// Decisão do produtor RTSP (Fase 6.4): evita enqueue/cópia de AUs descartados.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct MotionEnqueueDecision {
    pub enqueue: bool,
    pub decoder_reset: bool,
}

/// Gate no produtor: throttle + sync H.264 (só enfileira após IDR se houve gap).
#[derive(Debug, Clone)]
pub struct MotionEnqueueGate {
    throttle: MotionAnalysisThrottle,
    decoder_sync_pending: bool,
}

impl MotionEnqueueGate {
    pub fn from_max_fps(max_fps: f64) -> Self {
        Self {
            throttle: MotionAnalysisThrottle::from_max_fps(max_fps),
            decoder_sync_pending: false,
        }
    }

    pub fn is_unlimited(&self) -> bool {
        self.throttle.is_unlimited()
    }

    pub fn decide(&mut self, is_keyframe: bool) -> MotionEnqueueDecision {
        if self.is_unlimited() {
            return MotionEnqueueDecision {
                enqueue: true,
                decoder_reset: false,
            };
        }
        if is_keyframe {
            let enqueue = self.throttle.should_analyze(true);
            let decoder_reset = self.decoder_sync_pending;
            self.decoder_sync_pending = false;
            return MotionEnqueueDecision {
                enqueue,
                decoder_reset,
            };
        }
        if self.decoder_sync_pending {
            return MotionEnqueueDecision {
                enqueue: false,
                decoder_reset: false,
            };
        }
        if self.throttle.should_analyze(false) {
            return MotionEnqueueDecision {
                enqueue: true,
                decoder_reset: false,
            };
        }
        self.decoder_sync_pending = true;
        MotionEnqueueDecision {
            enqueue: false,
            decoder_reset: false,
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

    #[test]
    fn enqueue_gate_skips_copy_until_keyframe_after_throttle() {
        let mut g = MotionEnqueueGate::from_max_fps(10.0);
        assert!(g.decide(false).enqueue);
        assert!(!g.decide(false).enqueue);
        let k = g.decide(true);
        assert!(k.enqueue);
        assert!(k.decoder_reset);
    }
}
