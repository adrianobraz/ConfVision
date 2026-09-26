use std::time::Duration;

use chrono::{DateTime, Utc};
use serde::Serialize;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum StreamFailureClass {
    PathAbsent,
    Transient,
    Auth,
}

#[derive(Debug, Clone, Serialize)]
pub struct StreamPolicySnapshot {
    pub failures_consecutive: u32,
    pub hourly_attempts: u32,
    pub next_probe_at: Option<DateTime<Utc>>,
    pub policy_generation: u64,
    pub ever_stream_ok: bool,
    pub local_paused: bool,
}

#[derive(Debug, Clone, Copy)]
pub struct StreamRetryConfig {
    pub enabled: bool,
    pub never_ok_grace_hours: u64,
    pub hourly_failures_threshold: u32,
    pub hourly_max_attempts: u32,
    pub delay_fail_1_5_secs: u64,
    pub delay_fail_6_9_secs: u64,
    pub delay_fail_10_19_secs: u64,
    pub delay_fail_20_29_secs: u64,
    pub delay_fail_30_59_secs: u64,
    pub delay_fail_60_plus_secs: u64,
}

impl StreamRetryConfig {
    pub fn defaults() -> Self {
        Self {
            enabled: true,
            never_ok_grace_hours: 72,
            hourly_failures_threshold: 60,
            hourly_max_attempts: 6,
            delay_fail_1_5_secs: 60,
            delay_fail_6_9_secs: 120,
            delay_fail_10_19_secs: 600,
            delay_fail_20_29_secs: 1200,
            delay_fail_30_59_secs: 1800,
            delay_fail_60_plus_secs: 3600,
        }
    }
}

#[derive(Debug, Clone)]
pub enum StreamHealthAction {
    RetryAfter(Duration),
    ReportStreamOk,
    PauseAnalytic { reason: String },
    ReportFailure {
        failures: u32,
        hourly: u32,
    },
}
