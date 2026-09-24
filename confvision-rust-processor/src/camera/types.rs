use std::time::Instant;

use chrono::{DateTime, Utc};
use serde::Serialize;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum CameraStatus {
    Starting,
    Online,
    Offline,
    Reconnecting,
    Stopped,
    Error,
}

#[derive(Debug, Clone, Serialize)]
pub struct CameraRuntimeState {
    pub camera_id: i64,
    pub stream_url_redacted: String,
    pub status: CameraStatus,
    pub last_frame_at: Option<DateTime<Utc>>,
    pub last_error: Option<String>,
    pub fps: f64,
    pub reconnect_count: u64,
    pub started_at: DateTime<Utc>,
    pub frames_received: u64,
    pub frames_dropped: u64,
}

impl CameraRuntimeState {
    pub fn new(camera_id: i64, stream_url_redacted: String) -> Self {
        Self {
            camera_id,
            stream_url_redacted,
            status: CameraStatus::Starting,
            last_frame_at: None,
            last_error: None,
            fps: 0.0,
            reconnect_count: 0,
            started_at: Utc::now(),
            frames_received: 0,
            frames_dropped: 0,
        }
    }
}

pub struct FpsEstimator {
    window_start: Instant,
    frames_in_window: u64,
}

impl FpsEstimator {
    pub fn new() -> Self {
        Self {
            window_start: Instant::now(),
            frames_in_window: 0,
        }
    }

    pub fn record_frame(&mut self) -> Option<f64> {
        self.frames_in_window += 1;
        let elapsed = self.window_start.elapsed().as_secs_f64();
        if elapsed >= 1.0 {
            let fps = self.frames_in_window as f64 / elapsed;
            self.frames_in_window = 0;
            self.window_start = Instant::now();
            return Some(fps);
        }
        None
    }
}

impl Default for FpsEstimator {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::thread;
    use std::time::Duration;

    #[test]
    fn fps_estimator_computes_rate() {
        let mut est = FpsEstimator::new();
        for _ in 0..10 {
            est.record_frame();
        }
        thread::sleep(Duration::from_millis(1100));
        let fps = est.record_frame().unwrap();
        assert!(fps > 0.0);
    }
}
