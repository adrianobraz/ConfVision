//! Sessões de movimento (Started / Updated / Ended) alinhadas ao `MOTION_POST_ROLL_SEC` do Python.
//! Clips e upload vêm em `recording.rs` (Fase 2A).

use std::time::{Duration, Instant};

use chrono::Utc;
use serde::Serialize;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum MotionSessionPhase {
    Idle,
    Active,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum MotionSessionEventKind {
    MotionStarted,
    MotionUpdated,
    MotionEnded,
}

#[derive(Debug, Clone, Serialize)]
pub struct MotionSessionEvent {
    pub kind: MotionSessionEventKind,
    pub camera_id: i64,
    pub ts_unix_ms: i64,
    pub score_percent: u32,
    pub duration_ms: Option<u64>,
}

#[derive(Debug)]
pub struct MotionSessionTracker {
    camera_id: i64,
    post_roll: Duration,
    phase: MotionSessionPhase,
    session_start: Option<Instant>,
    last_motion: Option<Instant>,
}

impl MotionSessionTracker {
    pub fn new(camera_id: i64, post_roll: Duration) -> Self {
        Self {
            camera_id,
            post_roll,
            phase: MotionSessionPhase::Idle,
            session_start: None,
            last_motion: None,
        }
    }

    /// Chamado apenas em amostras `Analyzed` (não em reference/scene).
    pub fn on_motion_sample(&mut self, detected: bool, score_percent: u32) -> Vec<MotionSessionEvent> {
        let now = Instant::now();
        let mut out = Vec::new();

        if detected {
            self.last_motion = Some(now);
            match self.phase {
                MotionSessionPhase::Idle => {
                    self.phase = MotionSessionPhase::Active;
                    self.session_start = Some(now);
                    out.push(self.event(MotionSessionEventKind::MotionStarted, score_percent, None));
                }
                MotionSessionPhase::Active => {}
            }
            return out;
        }

        if self.phase != MotionSessionPhase::Active {
            return out;
        }

        let Some(last) = self.last_motion else {
            return out;
        };

        if now.duration_since(last) >= self.post_roll {
            let duration_ms = self
                .session_start
                .map(|s| now.duration_since(s).as_millis() as u64);
            out.push(self.event(
                MotionSessionEventKind::MotionEnded,
                score_percent,
                duration_ms,
            ));
            self.phase = MotionSessionPhase::Idle;
            self.session_start = None;
        }

        out
    }

    fn event(
        &self,
        kind: MotionSessionEventKind,
        score_percent: u32,
        duration_ms: Option<u64>,
    ) -> MotionSessionEvent {
        MotionSessionEvent {
            kind,
            camera_id: self.camera_id,
            ts_unix_ms: Utc::now().timestamp_millis(),
            score_percent,
            duration_ms,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn started_then_ended_after_post_roll() {
        let mut t = MotionSessionTracker::new(1, Duration::from_millis(50));
        let e1 = t.on_motion_sample(true, 10);
        assert_eq!(e1.len(), 1);
        assert!(matches!(e1[0].kind, MotionSessionEventKind::MotionStarted));

        std::thread::sleep(Duration::from_millis(60));
        let e2 = t.on_motion_sample(false, 0);
        assert_eq!(e2.len(), 1);
        assert!(matches!(e2[0].kind, MotionSessionEventKind::MotionEnded));
    }
}
