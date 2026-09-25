use std::sync::atomic::{AtomicBool, AtomicU32, Ordering};
use std::sync::Arc;

/// Estado compartilhado produtor RTSP ↔ consumer decode (como `yolo_ligado` no Python).
#[derive(Debug)]
pub struct MotionGatedSession {
    armed: AtomicBool,
    miss_streak: AtomicU32,
    miss_threshold: u32,
}

impl MotionGatedSession {
    pub fn new(miss_threshold: u32) -> Arc<Self> {
        Arc::new(Self {
            armed: AtomicBool::new(false),
            miss_streak: AtomicU32::new(0),
            miss_threshold: miss_threshold.max(1),
        })
    }

    pub fn is_armed(&self) -> bool {
        self.armed.load(Ordering::Relaxed)
    }

    /// Atualiza gate após motion no consumer. Referência inicial não altera estado.
    pub fn on_motion_analyzed(&self, detected: bool, reference_only: bool) {
        if reference_only {
            return;
        }
        if detected {
            self.armed.store(true, Ordering::Relaxed);
            self.miss_streak.store(0, Ordering::Relaxed);
            return;
        }
        if !self.is_armed() {
            return;
        }
        let streak = self.miss_streak.fetch_add(1, Ordering::Relaxed) + 1;
        if streak >= self.miss_threshold {
            self.armed.store(false, Ordering::Relaxed);
            self.miss_streak.store(0, Ordering::Relaxed);
        }
    }

    pub fn reset(&self) {
        self.armed.store(false, Ordering::Relaxed);
        self.miss_streak.store(0, Ordering::Relaxed);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn arms_on_motion() {
        let s = MotionGatedSession::new(3);
        assert!(!s.is_armed());
        s.on_motion_analyzed(true, false);
        assert!(s.is_armed());
    }

    #[test]
    fn disarms_after_miss_threshold() {
        let s = MotionGatedSession::new(3);
        s.on_motion_analyzed(true, false);
        s.on_motion_analyzed(false, false);
        s.on_motion_analyzed(false, false);
        assert!(s.is_armed());
        s.on_motion_analyzed(false, false);
        assert!(!s.is_armed());
    }

    #[test]
    fn idle_ignores_misses() {
        let s = MotionGatedSession::new(2);
        s.on_motion_analyzed(false, false);
        s.on_motion_analyzed(false, false);
        assert!(!s.is_armed());
    }
}
