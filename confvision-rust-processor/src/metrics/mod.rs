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
    pub frames_enqueued: AtomicU64,
    pub frames_processed: AtomicU64,
    pub buffer_full_events: AtomicU64,
    /// Última latência observada no consumer (ms).
    pub last_frame_latency_ms: AtomicU64,
    pub frames_decoded: AtomicU64,
    pub decode_errors: AtomicU64,
    pub last_decode_ms: AtomicU64,
    pub frames_motion_analyzed: AtomicU64,
    pub motion_detected: AtomicU64,
    pub motion_errors: AtomicU64,
    pub last_motion_score: AtomicU64,
    pub last_motion_ms: AtomicU64,
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
            frames_enqueued: AtomicU64::new(0),
            frames_processed: AtomicU64::new(0),
            buffer_full_events: AtomicU64::new(0),
            last_frame_latency_ms: AtomicU64::new(0),
            frames_decoded: AtomicU64::new(0),
            decode_errors: AtomicU64::new(0),
            last_decode_ms: AtomicU64::new(0),
            frames_motion_analyzed: AtomicU64::new(0),
            motion_detected: AtomicU64::new(0),
            motion_errors: AtomicU64::new(0),
            last_motion_score: AtomicU64::new(0),
            last_motion_ms: AtomicU64::new(0),
            reconnects: AtomicU64::new(0),
            rtsp_errors: AtomicU64::new(0),
            errors: AtomicU64::new(0),
        }
    }

    /// RTSP recebeu frame (não inclui drops da pipeline).
    pub fn add_frames_received(&self, received: u64) {
        self.frames_received.fetch_add(received, Ordering::Relaxed);
    }

    pub fn add_frames(&self, received: u64, dropped: u64) {
        self.frames_received.fetch_add(received, Ordering::Relaxed);
        self.frames_dropped.fetch_add(dropped, Ordering::Relaxed);
    }

    pub fn record_enqueued(&self, count: u64) {
        self.frames_enqueued.fetch_add(count, Ordering::Relaxed);
    }

    pub fn add_pipeline_dropped(&self, count: u64) {
        self.frames_dropped.fetch_add(count, Ordering::Relaxed);
    }

    pub fn record_processed(&self, latency_ms: u64) {
        self.frames_processed.fetch_add(1, Ordering::Relaxed);
        self.last_frame_latency_ms
            .store(latency_ms, Ordering::Relaxed);
    }

    pub fn record_buffer_full_event(&self) {
        self.buffer_full_events.fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_decode_success(&self, latency_ms: u64) {
        self.frames_decoded.fetch_add(1, Ordering::Relaxed);
        self.last_decode_ms.store(latency_ms, Ordering::Relaxed);
    }

    pub fn record_decode_error(&self) {
        self.decode_errors.fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_motion_analyzed(&self, score_percent: u32, detected: bool, latency_ms: u64) {
        self.frames_motion_analyzed.fetch_add(1, Ordering::Relaxed);
        self.last_motion_score
            .store(score_percent as u64, Ordering::Relaxed);
        self.last_motion_ms.store(latency_ms, Ordering::Relaxed);
        if detected {
            self.motion_detected.fetch_add(1, Ordering::Relaxed);
        }
    }

    pub fn record_motion_error(&self) {
        self.motion_errors.fetch_add(1, Ordering::Relaxed);
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
        self.snapshot_with_acceleration(None)
    }

    pub fn snapshot_with_acceleration(
        &self,
        acceleration: Option<&crate::decode::AccelerationRuntime>,
    ) -> MetricsSnapshot {
        let accel = acceleration.map(acceleration_snapshot_fields);
        MetricsSnapshot {
            processor_id: self.processor_id.clone(),
            uptime_secs: self.uptime_secs(),
            frames_received: self.frames_received.load(Ordering::Relaxed),
            frames_dropped: self.frames_dropped.load(Ordering::Relaxed),
            frames_enqueued: self.frames_enqueued.load(Ordering::Relaxed),
            frames_processed: self.frames_processed.load(Ordering::Relaxed),
            buffer_full_events: self.buffer_full_events.load(Ordering::Relaxed),
            frame_latency_ms: self.last_frame_latency_ms.load(Ordering::Relaxed),
            frames_decoded: self.frames_decoded.load(Ordering::Relaxed),
            decode_errors: self.decode_errors.load(Ordering::Relaxed),
            decode_ms: self.last_decode_ms.load(Ordering::Relaxed),
            last_decode_ms: self.last_decode_ms.load(Ordering::Relaxed),
            frames_motion_analyzed: self.frames_motion_analyzed.load(Ordering::Relaxed),
            motion_detected: self.motion_detected.load(Ordering::Relaxed),
            motion_errors: self.motion_errors.load(Ordering::Relaxed),
            last_motion_score: self.last_motion_score.load(Ordering::Relaxed),
            last_motion_ms: self.last_motion_ms.load(Ordering::Relaxed),
            reconnects: self.reconnects.load(Ordering::Relaxed),
            rtsp_errors: self.rtsp_errors.load(Ordering::Relaxed),
            errors: self.errors.load(Ordering::Relaxed),
            video_acceleration_requested: accel
                .as_ref()
                .map(|a| a.requested.clone())
                .unwrap_or_else(|| "unknown".into()),
            video_acceleration_effective: accel
                .as_ref()
                .map(|a| a.effective.clone())
                .unwrap_or_else(|| "cpu".into()),
            video_hw_backend: accel
                .as_ref()
                .map(|a| a.hw_backend.clone())
                .unwrap_or_else(|| "none".into()),
            gpu_detected: accel.as_ref().map(|a| a.gpu_detected).unwrap_or(false),
            ffmpeg_hw_decode_available: accel
                .as_ref()
                .map(|a| a.ffmpeg_hw_decode_available)
                .unwrap_or(false),
            hardware_decode_active: accel
                .as_ref()
                .map(|a| a.hardware_decode_active)
                .unwrap_or(false),
            hw_decode_errors: accel.as_ref().map(|a| a.hw_decode_errors).unwrap_or(0),
            hw_fallback_to_cpu_count: accel
                .as_ref()
                .map(|a| a.hw_fallback_to_cpu_count)
                .unwrap_or(0),
        }
    }
}

struct AccelerationSnapshotFields {
    requested: String,
    effective: String,
    hw_backend: String,
    gpu_detected: bool,
    ffmpeg_hw_decode_available: bool,
    hardware_decode_active: bool,
    hw_decode_errors: u64,
    hw_fallback_to_cpu_count: u64,
}

fn acceleration_snapshot_fields(
    acceleration: &crate::decode::AccelerationRuntime,
) -> AccelerationSnapshotFields {
    AccelerationSnapshotFields {
        requested: acceleration.video_acceleration_requested().into(),
        effective: acceleration.video_acceleration_effective().into(),
        hw_backend: acceleration.video_hw_backend().into(),
        gpu_detected: acceleration.gpu_detected(),
        ffmpeg_hw_decode_available: acceleration.ffmpeg_hw_decode_available(),
        hardware_decode_active: acceleration.hardware_decode_active(),
        hw_decode_errors: acceleration.hw_decode_errors(),
        hw_fallback_to_cpu_count: acceleration.hw_fallback_to_cpu_count(),
    }
}

#[derive(Debug, Clone, Serialize)]
pub struct MetricsSnapshot {
    pub processor_id: String,
    pub uptime_secs: u64,
    pub frames_received: u64,
    pub frames_dropped: u64,
    pub frames_enqueued: u64,
    pub frames_processed: u64,
    pub buffer_full_events: u64,
    pub frame_latency_ms: u64,
    pub frames_decoded: u64,
    pub decode_errors: u64,
    /// Último decode bem-sucedido (ms).
    pub decode_ms: u64,
    pub last_decode_ms: u64,
    pub frames_motion_analyzed: u64,
    pub motion_detected: u64,
    pub motion_errors: u64,
    pub last_motion_score: u64,
    pub last_motion_ms: u64,
    pub reconnects: u64,
    pub rtsp_errors: u64,
    pub errors: u64,
    pub video_acceleration_requested: String,
    pub video_acceleration_effective: String,
    pub video_hw_backend: String,
    pub gpu_detected: bool,
    pub ffmpeg_hw_decode_available: bool,
    pub hardware_decode_active: bool,
    pub hw_decode_errors: u64,
    pub hw_fallback_to_cpu_count: u64,
}

pub type SharedMetrics = Arc<ProcessorMetrics>;
