use std::time::Duration;

use super::types::{StreamFailureClass, StreamRetryConfig};

pub fn classify_rtsp_error(message: &str) -> StreamFailureClass {
    match classify_stream_error_code(message) {
        "rtsp_auth" => StreamFailureClass::Auth,
        "rtsp_404" => StreamFailureClass::PathAbsent,
        _ => StreamFailureClass::Transient,
    }
}

/// Código estável para ping/API (`stream_erro_classe`), independente da política de retry.
pub fn classify_stream_error_code(message: &str) -> &'static str {
    let m = message.to_ascii_lowercase();
    if m.contains("401") || m.contains("403") || m.contains("unauthorized") {
        return "rtsp_auth";
    }
    if m.contains("404")
        || m.contains("not found")
        || m.contains("describe failed")
        || m.contains("unexpected rtsp response status")
    {
        return "rtsp_404";
    }
    if m.contains("fu-a")
        || m.contains("fu_a")
        || (m.contains("fragmentation unit") && m.contains("h264"))
        || m.contains("start bit unset")
    {
        return "rtp_h264_fu_a";
    }
    if m.contains("timeout")
        || m.contains("timed out")
        || m.contains("connection refused")
        || m.contains("broken pipe")
        || m.contains("connection reset")
    {
        return "network_timeout";
    }
    "unknown_transient"
}

pub fn delay_after_failure(failures: u32, cfg: &StreamRetryConfig) -> Duration {
    let secs = if failures <= 5 {
        cfg.delay_fail_1_5_secs
    } else if failures <= 9 {
        cfg.delay_fail_6_9_secs
    } else if failures <= 19 {
        cfg.delay_fail_10_19_secs
    } else if failures <= 29 {
        cfg.delay_fail_20_29_secs
    } else if failures <= 59 {
        cfg.delay_fail_30_59_secs
    } else {
        cfg.delay_fail_60_plus_secs
    };
    Duration::from_secs(secs.max(1))
}

pub fn should_auto_pause(failures: u32, hourly: u32, cfg: &StreamRetryConfig) -> bool {
    failures >= cfg.hourly_failures_threshold && hourly >= cfg.hourly_max_attempts
}
