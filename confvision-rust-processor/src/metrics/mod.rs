use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::Arc;
use std::time::Instant;

use serde::Serialize;

#[derive(Debug)]
pub struct ProcessorMetrics {
    pub processor_id: String,
    pub started_at: Instant,
    pub frames_received: AtomicU64,
    pub frames_dropped: AtomicU64,
    pub reconnects: AtomicU64,
    pub rtsp_errors: AtomicU64,
    pub errors: AtomicU64,
}

impl ProcessorMetrics {
    pub fn new(processor_id: impl Into<String>) -> Self {
        Self {
            processor_id: processor_id.into(),
            started_at: Instant::now(),
            frames_received: AtomicU64::new(0),
            frames_dropped: AtomicU64::new(0),
            reconnects: AtomicU64::new(0),
            rtsp_errors: AtomicU64::new(0),
            errors: AtomicU64::new(0),
        }
    }

    pub fn add_frames(&self, received: u64, dropped: u64) {
        self.frames_received.fetch_add(received, Ordering::Relaxed);
        self.frames_dropped.fetch_add(dropped, Ordering::Relaxed);
    }

    pub fn record_reconnect(&self) {
        self.reconnects.fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_rtsp_error(&self) {
        self.rtsp_errors.fetch_add(1, Ordering::Relaxed);
    }

    pub fn uptime_secs(&self) -> u64 {
        self.started_at.elapsed().as_secs()
    }

    pub fn snapshot(&self) -> MetricsSnapshot {
        MetricsSnapshot {
            processor_id: self.processor_id.clone(),
            uptime_secs: self.uptime_secs(),
            frames_received: self.frames_received.load(Ordering::Relaxed),
            frames_dropped: self.frames_dropped.load(Ordering::Relaxed),
            reconnects: self.reconnects.load(Ordering::Relaxed),
            rtsp_errors: self.rtsp_errors.load(Ordering::Relaxed),
            errors: self.errors.load(Ordering::Relaxed),
        }
    }
}

#[derive(Debug, Clone, Serialize)]
pub struct MetricsSnapshot {
    pub processor_id: String,
    pub uptime_secs: u64,
    pub frames_received: u64,
    pub frames_dropped: u64,
    pub reconnects: u64,
    pub rtsp_errors: u64,
    pub errors: u64,
}

pub type SharedMetrics = Arc<ProcessorMetrics>;
