mod policy;
mod types;

pub use policy::{
    classify_rtsp_error, classify_stream_error_code, delay_after_failure, should_auto_pause,
};
pub use types::{
    StreamFailureClass, StreamHealthAction, StreamPolicySnapshot, StreamRetryConfig,
};

use chrono::{DateTime, Utc};
use std::time::Duration;

/// Estado mutável por câmera (processor + espelho parcial do Postgres).
#[derive(Debug, Clone)]
pub struct StreamPolicyState {
    pub failures_consecutive: u32,
    pub hourly_attempts: u32,
    pub next_probe_at: Option<DateTime<Utc>>,
    pub policy_generation: u64,
    pub ever_stream_ok: bool,
    pub local_paused: bool,
}

impl Default for StreamPolicyState {
    fn default() -> Self {
        Self {
            failures_consecutive: 0,
            hourly_attempts: 0,
            next_probe_at: None,
            policy_generation: 0,
            ever_stream_ok: false,
            local_paused: false,
        }
    }
}

impl StreamPolicyState {
    pub fn snapshot(&self) -> StreamPolicySnapshot {
        StreamPolicySnapshot {
            failures_consecutive: self.failures_consecutive,
            hourly_attempts: self.hourly_attempts,
            next_probe_at: self.next_probe_at,
            policy_generation: self.policy_generation,
            ever_stream_ok: self.ever_stream_ok,
            local_paused: self.local_paused,
        }
    }

    pub fn apply_camera_metadata(
        &mut self,
        policy_generation: u64,
        ultimo_stream_ok_em: Option<&str>,
    ) {
        if policy_generation != self.policy_generation {
            self.failures_consecutive = 0;
            self.hourly_attempts = 0;
            self.next_probe_at = None;
            self.local_paused = false;
            self.policy_generation = policy_generation;
        }
        if let Some(raw) = ultimo_stream_ok_em {
            if !raw.trim().is_empty() {
                self.ever_stream_ok = true;
            }
        }
    }

    pub fn wait_before_probe(&self) -> Option<Duration> {
        let at = self.next_probe_at?;
        let now = Utc::now();
        if at <= now {
            return None;
        }
        let secs = (at - now).num_seconds().max(0) as u64;
        Some(Duration::from_secs(secs))
    }

    pub fn on_success(&mut self, _cfg: &StreamRetryConfig) -> StreamHealthAction {
        self.failures_consecutive = 0;
        self.hourly_attempts = 0;
        self.next_probe_at = None;
        self.local_paused = false;
        self.ever_stream_ok = true;
        StreamHealthAction::ReportStreamOk
    }

    pub fn on_failure(
        &mut self,
        cfg: &StreamRetryConfig,
        class: StreamFailureClass,
        camera_created_at: Option<DateTime<Utc>>,
    ) -> StreamHealthAction {
        if !cfg.enabled {
            return StreamHealthAction::RetryAfter(Duration::from_secs(30));
        }

        if class == StreamFailureClass::Auth {
            let delay = Duration::from_secs(30 * 60);
            self.next_probe_at =
                Some(Utc::now() + chrono::Duration::from_std(delay).unwrap_or(chrono::Duration::minutes(30)));
            return StreamHealthAction::RetryAfter(delay);
        }

        if class == StreamFailureClass::Transient {
            self.failures_consecutive = self.failures_consecutive.saturating_add(1);
            let delay = delay_after_failure(self.failures_consecutive, cfg);
            self.next_probe_at = Some(
                Utc::now()
                    + chrono::Duration::from_std(delay).unwrap_or(chrono::Duration::minutes(1)),
            );
            return StreamHealthAction::RetryAfter(delay);
        }

        self.failures_consecutive = self.failures_consecutive.saturating_add(1);

        if !self.ever_stream_ok && grace_elapsed(cfg, camera_created_at) {
            self.local_paused = true;
            return StreamHealthAction::PauseAnalytic {
                reason: "sistema_stream_sem_historico".into(),
            };
        }

        if self.failures_consecutive >= cfg.hourly_failures_threshold {
            self.hourly_attempts = self.hourly_attempts.saturating_add(1);
            if self.hourly_attempts >= cfg.hourly_max_attempts {
                self.local_paused = true;
                return StreamHealthAction::PauseAnalytic {
                    reason: "sistema_stream_6h_horarias".into(),
                };
            }
        }

        let delay = delay_after_failure(self.failures_consecutive, cfg);
        self.next_probe_at = Some(
            Utc::now() + chrono::Duration::from_std(delay).unwrap_or(chrono::Duration::minutes(1)),
        );
        StreamHealthAction::RetryAfter(delay)
    }
}

fn grace_elapsed(cfg: &StreamRetryConfig, created_at: Option<DateTime<Utc>>) -> bool {
    let Some(created) = created_at else {
        return true;
    };
    let grace = chrono::Duration::hours(cfg.never_ok_grace_hours as i64);
    Utc::now() - created > grace
}

#[cfg(test)]
mod tests {
    use super::*;

    fn cfg() -> StreamRetryConfig {
        StreamRetryConfig::defaults()
    }

    #[test]
    fn resets_on_generation_change() {
        let mut s = StreamPolicyState {
            failures_consecutive: 10,
            policy_generation: 1,
            ..Default::default()
        };
        s.apply_camera_metadata(2, None);
        assert_eq!(s.failures_consecutive, 0);
        assert_eq!(s.policy_generation, 2);
    }

    #[test]
    fn tier_delays() {
        assert_eq!(delay_after_failure(3, &cfg()), Duration::from_secs(60));
        assert_eq!(delay_after_failure(7, &cfg()), Duration::from_secs(120));
        assert_eq!(delay_after_failure(15, &cfg()), Duration::from_secs(600));
    }

    #[test]
    fn classifies_fu_a_and_404() {
        use super::policy::{classify_rtsp_error, classify_stream_error_code};
        use super::types::StreamFailureClass;

        let fu = "FU-A has start bit unset on NAL type 1";
        assert_eq!(classify_stream_error_code(fu), "rtp_h264_fu_a");
        assert_eq!(classify_rtsp_error(fu), StreamFailureClass::Transient);

        let nf = "unexpected rtsp response status: 404";
        assert_eq!(classify_stream_error_code(nf), "rtsp_404");
        assert_eq!(classify_rtsp_error(nf), StreamFailureClass::PathAbsent);
    }
}
