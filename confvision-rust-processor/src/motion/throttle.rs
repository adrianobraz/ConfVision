use std::sync::Arc;
use std::time::{Duration, Instant};

use super::MotionGatedSession;
use crate::config::Config;

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

    pub fn should_analyze(&mut self) -> bool {
        if self.is_unlimited() {
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

    pub fn mark_consumed(&mut self) {
        if !self.is_unlimited() {
            self.last_analysis = Some(Instant::now());
        }
    }
}

/// Decisão do produtor RTSP (Fase 6.4): evita enqueue/cópia de AUs descartados.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct MotionEnqueueDecision {
    pub enqueue: bool,
    pub decoder_reset: bool,
}

/// Gate no produtor: stride (legacy Python) + throttle FPS + sync H.264 após gap.
#[derive(Debug, Clone)]
pub struct MotionEnqueueGate {
    active_throttle: MotionAnalysisThrottle,
    probe_throttle: MotionAnalysisThrottle,
    motion_gated: bool,
    gated_session: Option<Arc<MotionGatedSession>>,
    motion_frame_stride: u64,
    decode_frame_stride: u64,
    video_au_index: u64,
    decoder_sync_pending: bool,
}

impl MotionEnqueueGate {
    pub fn from_max_fps(max_fps: f64) -> Self {
        Self::from_analysis_config(max_fps, 1, 1, false, None, 0.5)
    }

    pub fn from_config(cfg: &Config, gated_session: Option<Arc<MotionGatedSession>>) -> Self {
        let motion_gated = cfg.analysis_only_on_motion && gated_session.is_some();
        Self::from_analysis_config(
            cfg.motion_analysis_max_fps,
            cfg.motion_frame_stride,
            cfg.decode_frame_stride,
            motion_gated,
            gated_session,
            cfg.motion_gate_probe_max_fps,
        )
    }

    /// Alinhado a `MOTION_FRAME_SKIP` / `FRAME_SKIP` do worker Python (decode+motion no Rust).
    pub fn from_analysis_config(
        max_fps: f64,
        motion_frame_stride: usize,
        decode_frame_stride: usize,
        motion_gated: bool,
        gated_session: Option<Arc<MotionGatedSession>>,
        probe_max_fps: f64,
    ) -> Self {
        Self {
            active_throttle: MotionAnalysisThrottle::from_max_fps(max_fps),
            probe_throttle: MotionAnalysisThrottle::from_max_fps(probe_max_fps),
            motion_gated,
            gated_session,
            motion_frame_stride: motion_frame_stride.max(1) as u64,
            decode_frame_stride: decode_frame_stride.max(1) as u64,
            video_au_index: 0,
            decoder_sync_pending: false,
        }
    }

    fn in_probe_mode(&self) -> bool {
        self.motion_gated
            && self
                .gated_session
                .as_ref()
                .is_some_and(|s| !s.is_armed())
    }

    fn throttle_for_mode(&mut self) -> &mut MotionAnalysisThrottle {
        if self.in_probe_mode() {
            &mut self.probe_throttle
        } else {
            &mut self.active_throttle
        }
    }

    pub fn is_unlimited(&self) -> bool {
        if self.in_probe_mode() {
            return self.probe_throttle.is_unlimited();
        }
        self.active_throttle.is_unlimited()
            && self.motion_frame_stride <= 1
            && self.decode_frame_stride <= 1
    }

    fn strides_allow(&self, index: u64) -> bool {
        if self.in_probe_mode() {
            return true;
        }
        index % self.motion_frame_stride == 0 && index % self.decode_frame_stride == 0
    }

    pub fn decide(&mut self, is_keyframe: bool) -> MotionEnqueueDecision {
        self.video_au_index = self.video_au_index.saturating_add(1);

        if self.decoder_sync_pending {
            if !is_keyframe {
                return MotionEnqueueDecision {
                    enqueue: false,
                    decoder_reset: false,
                };
            }
            self.decoder_sync_pending = false;
            self.throttle_for_mode().mark_consumed();
            return MotionEnqueueDecision {
                enqueue: true,
                decoder_reset: true,
            };
        }

        if !self.strides_allow(self.video_au_index) {
            return MotionEnqueueDecision {
                enqueue: false,
                decoder_reset: false,
            };
        }

        if self.is_unlimited() {
            return MotionEnqueueDecision {
                enqueue: true,
                decoder_reset: false,
            };
        }

        if self.throttle_for_mode().should_analyze() {
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
        assert!(t.should_analyze());
        assert!(t.should_analyze());
    }

    #[test]
    fn caps_by_interval() {
        let mut t = MotionAnalysisThrottle::from_max_fps(10.0);
        assert!(t.should_analyze());
        assert!(!t.should_analyze());
        thread::sleep(Duration::from_millis(105));
        assert!(t.should_analyze());
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

    #[test]
    fn stride_reduces_enqueues() {
        let mut g = MotionEnqueueGate::from_analysis_config(1000.0, 3, 1, false, None, 0.5);
        assert!(!g.decide(false).enqueue);
        assert!(!g.decide(false).enqueue);
        assert!(g.decide(false).enqueue);
    }

    #[test]
    fn decode_stride_requires_both() {
        let mut g = MotionEnqueueGate::from_analysis_config(1000.0, 2, 5, false, None, 0.5);
        for _ in 0..9 {
            let _ = g.decide(false);
        }
        assert!(g.decide(false).enqueue);
    }

    #[test]
    fn keyframe_does_not_bypass_fps_cap() {
        let mut g = MotionEnqueueGate::from_max_fps(10.0);
        assert!(g.decide(false).enqueue);
        assert!(!g.decide(true).enqueue);
    }

    #[test]
    fn gated_idle_uses_probe_throttle() {
        let session = MotionGatedSession::new(5);
        let mut g = MotionEnqueueGate::from_analysis_config(1000.0, 1, 1, true, Some(session), 10.0);
        assert!(g.decide(false).enqueue);
        assert!(!g.decide(false).enqueue);
    }

    #[test]
    fn gated_armed_uses_active_stride() {
        let session = MotionGatedSession::new(5);
        session.on_motion_analyzed(true, false);
        let mut g = MotionEnqueueGate::from_analysis_config(1000.0, 3, 1, true, Some(session), 0.1);
        assert!(!g.decide(false).enqueue);
        assert!(!g.decide(false).enqueue);
        assert!(g.decide(false).enqueue);
    }
}
