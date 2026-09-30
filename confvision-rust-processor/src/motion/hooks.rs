//! Hooks pós-motion no consumer (sessão + shadow).

use std::time::Duration;

use chrono::Utc;

use super::session::{MotionSessionEvent, MotionSessionTracker};
use super::shadow::{MotionShadowConfig, MotionShadowWriter};

#[derive(Debug, Clone)]
pub struct MotionHookConfig {
    pub shadow: MotionShadowConfig,
    pub post_roll: Duration,
}

impl MotionHookConfig {
    pub fn from_config(cfg: &crate::config::Config) -> Self {
        Self {
            shadow: MotionShadowConfig::from_config(cfg),
            post_roll: Duration::from_secs(cfg.motion_post_roll_sec.max(1) as u64),
        }
    }
}

pub struct MotionPipelineHooks {
    session: MotionSessionTracker,
    shadow: Option<MotionShadowWriter>,
}

impl MotionPipelineHooks {
    pub fn new(camera_id: i64, cfg: &MotionHookConfig) -> Self {
        Self {
            session: MotionSessionTracker::new(camera_id, cfg.post_roll),
            shadow: MotionShadowWriter::new(&cfg.shadow, camera_id),
        }
    }

    pub fn on_analyzed(&mut self, detected: bool, score_percent: u32) -> Vec<MotionSessionEvent> {
        let ts = Utc::now().timestamp_millis();
        if let Some(w) = self.shadow.as_ref() {
            w.write_sample(detected, score_percent, ts);
        }
        let events = self.session.on_motion_sample(detected, score_percent);
        if let Some(w) = self.shadow.as_ref() {
            for ev in &events {
                w.write_session_event(ev);
            }
        }
        events
    }
}
