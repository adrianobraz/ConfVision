use std::collections::VecDeque;
use std::sync::atomic::{AtomicBool, AtomicU64, AtomicUsize, Ordering};
use std::sync::{Arc, Mutex};
use std::time::Instant;

use tokio::sync::Notify;

use crate::camera::SharedCameraState;
use crate::decode::{
    AccelerationRuntime, DecodeInput, DecodeOutcome, DecodePolicyCoordinator, H264Decoder,
    SessionDecodeContext,
};
use crate::metrics::ProcessorMetrics;
use crate::motion::{MotionDetector, MotionOutcome};
use crate::pipeline::PipelineFrame;

/// Fila bounded com política DROP-OLDEST (testável de forma síncrona).
#[derive(Debug)]
pub struct DropOldestQueue {
    capacity: usize,
    queue: VecDeque<PipelineFrame>,
    pub enqueued: u64,
    pub dropped: u64,
    pub full_events: u64,
}

impl DropOldestQueue {
    pub fn new(capacity: usize) -> Self {
        Self {
            capacity: capacity.max(1),
            queue: VecDeque::new(),
            enqueued: 0,
            dropped: 0,
            full_events: 0,
        }
    }

    pub fn capacity(&self) -> usize {
        self.capacity
    }

    pub fn len(&self) -> usize {
        self.queue.len()
    }

    pub fn is_empty(&self) -> bool {
        self.queue.is_empty()
    }

    /// Insere frame; se cheio, descarta o mais antigo. Nunca bloqueia.
    pub fn push(&mut self, frame: PipelineFrame) {
        if self.queue.len() >= self.capacity {
            self.queue.pop_front();
            self.dropped += 1;
            self.full_events += 1;
        }
        self.queue.push_back(frame);
        self.enqueued += 1;
    }

    pub fn pop(&mut self) -> Option<PipelineFrame> {
        self.queue.pop_front()
    }

    pub fn clear(&mut self) {
        self.queue.clear();
    }
}

/// Contadores do producer quando `try_write` no estado da câmera falha (consumer faz flush).
#[derive(Debug, Default)]
struct PendingProducerStats {
    enqueued: AtomicU64,
    dropped: AtomicU64,
    buffer_full_events: AtomicU64,
    buffer_size: AtomicUsize,
    buffer_size_pending: AtomicBool,
    decode_errors: AtomicU64,
    motion_errors: AtomicU64,
}

struct FramePipelineInner {
    queue: Mutex<DropOldestQueue>,
    closed: AtomicBool,
    notify: Notify,
    buffer_size: AtomicUsize,
    metrics: Arc<ProcessorMetrics>,
    state: SharedCameraState,
    pending_producer: PendingProducerStats,
}

/// Pipeline assíncrona por câmera (producer RTSP / consumer dedicado).
#[derive(Clone)]
pub struct FramePipeline {
    inner: Arc<FramePipelineInner>,
}

impl FramePipeline {
    pub fn new(capacity: usize, metrics: Arc<ProcessorMetrics>, state: SharedCameraState) -> Self {
        let cap = capacity.max(1);
        let pipeline = Self {
            inner: Arc::new(FramePipelineInner {
                queue: Mutex::new(DropOldestQueue::new(cap)),
                closed: AtomicBool::new(false),
                notify: Notify::new(),
                buffer_size: AtomicUsize::new(0),
                metrics,
                state,
                pending_producer: PendingProducerStats::default(),
            }),
        };
        pipeline.set_camera_buffer_capacity(cap);
        pipeline
    }

    fn apply_producer_enqueue_stats(&self, new_len: usize, buffer_was_full: bool) {
        if let Ok(mut s) = self.inner.state.try_write() {
            s.buffer_size = new_len as u64;
            s.frames_enqueued += 1;
            if buffer_was_full {
                s.frames_dropped += 1;
                s.buffer_full_events += 1;
            }
            return;
        }

        self.inner
            .pending_producer
            .enqueued
            .fetch_add(1, Ordering::Relaxed);
        if buffer_was_full {
            self.inner
                .pending_producer
                .dropped
                .fetch_add(1, Ordering::Relaxed);
            self.inner
                .pending_producer
                .buffer_full_events
                .fetch_add(1, Ordering::Relaxed);
        }
        self.inner
            .pending_producer
            .buffer_size
            .store(new_len, Ordering::Relaxed);
        self.inner
            .pending_producer
            .buffer_size_pending
            .store(true, Ordering::Relaxed);
    }

    fn apply_producer_buffer_size(&self, new_len: usize) {
        if let Ok(mut s) = self.inner.state.try_write() {
            s.buffer_size = new_len as u64;
            return;
        }
        self.inner
            .pending_producer
            .buffer_size
            .store(new_len, Ordering::Relaxed);
        self.inner
            .pending_producer
            .buffer_size_pending
            .store(true, Ordering::Relaxed);
    }

    /// Mescla estatísticas enfileiradas pelo producer RTSP (chamar no consumer).
    pub async fn flush_producer_stats(&self) {
        let enqueued = self
            .inner
            .pending_producer
            .enqueued
            .swap(0, Ordering::Relaxed);
        let dropped = self
            .inner
            .pending_producer
            .dropped
            .swap(0, Ordering::Relaxed);
        let full_events = self
            .inner
            .pending_producer
            .buffer_full_events
            .swap(0, Ordering::Relaxed);
        let buffer_dirty = self
            .inner
            .pending_producer
            .buffer_size_pending
            .swap(false, Ordering::Relaxed);
        let decode_errors = self
            .inner
            .pending_producer
            .decode_errors
            .swap(0, Ordering::Relaxed);
        let motion_errors = self
            .inner
            .pending_producer
            .motion_errors
            .swap(0, Ordering::Relaxed);

        if enqueued == 0
            && dropped == 0
            && full_events == 0
            && !buffer_dirty
            && decode_errors == 0
            && motion_errors == 0
        {
            return;
        }

        let mut s = self.inner.state.write().await;
        if enqueued > 0 {
            s.frames_enqueued += enqueued;
        }
        if dropped > 0 {
            s.frames_dropped += dropped;
        }
        if full_events > 0 {
            s.buffer_full_events += full_events;
        }
        if buffer_dirty {
            s.buffer_size = self
                .inner
                .pending_producer
                .buffer_size
                .load(Ordering::Relaxed) as u64;
        }
        if decode_errors > 0 {
            s.decode_errors += decode_errors;
        }
        if motion_errors > 0 {
            s.motion_errors += motion_errors;
        }
    }

    pub fn capacity(&self) -> usize {
        self.inner.queue.lock().unwrap().capacity()
    }

    pub fn len(&self) -> usize {
        self.inner.buffer_size.load(Ordering::Relaxed)
    }

    pub fn is_closed(&self) -> bool {
        self.inner.closed.load(Ordering::Relaxed)
    }

    /// Producer: nunca bloqueia aguardando consumer.
    pub fn try_enqueue(&self, frame: PipelineFrame) -> bool {
        if self.is_closed() {
            return false;
        }

        let mut guard = self.inner.queue.lock().unwrap();
        let len_before = guard.len();
        let was_full = len_before >= guard.capacity();
        guard.push(frame);
        if was_full {
            self.inner.metrics.record_buffer_full_event();
            self.inner.metrics.add_pipeline_dropped(1);
        }
        self.inner.metrics.record_enqueued(1);
        let new_len = guard.len();
        drop(guard);

        self.inner.buffer_size.store(new_len, Ordering::Relaxed);
        self.apply_producer_enqueue_stats(new_len, was_full);
        self.inner.notify.notify_one();
        true
    }

    fn try_pop(&self) -> Option<PipelineFrame> {
        let mut guard = self.inner.queue.lock().unwrap();
        let frame = guard.pop();
        let new_len = guard.len();
        drop(guard);
        self.inner.buffer_size.store(new_len, Ordering::Relaxed);
        self.apply_producer_buffer_size(new_len);
        frame
    }

    /// Consumer: aguarda frame ou fechamento.
    pub async fn dequeue(&self) -> Option<PipelineFrame> {
        loop {
            if let Some(frame) = self.try_pop() {
                return Some(frame);
            }
            if self.is_closed() {
                return None;
            }
            self.inner.notify.notified().await;
        }
    }

    /// Encerra a pipeline e esvazia a fila (reconexão / shutdown de sessão).
    pub fn close(&self) {
        self.inner.closed.store(true, Ordering::Relaxed);
        {
            let mut guard = self.inner.queue.lock().unwrap();
            guard.clear();
        }
        self.inner.buffer_size.store(0, Ordering::Relaxed);
        self.apply_producer_buffer_size(0);
        self.inner.notify.notify_waiters();
    }

    fn set_camera_buffer_capacity(&self, cap: usize) {
        if let Ok(mut s) = self.inner.state.try_write() {
            s.buffer_capacity = cap as u64;
        }
    }

    pub async fn record_processed(&self, frame: &PipelineFrame) {
        let latency_ms = frame.captured_at.elapsed().as_millis() as u64;
        self.inner.metrics.record_processed(latency_ms);

        let mut s = self.inner.state.write().await;
        s.frames_processed += 1;
        s.last_frame_latency_ms = latency_ms;
    }

    pub async fn record_decode_success(&self, latency_ms: u64) {
        self.inner.metrics.record_decode_success(latency_ms);
        let mut s = self.inner.state.write().await;
        s.frames_decoded += 1;
        s.last_decode_ms = latency_ms;
    }

    pub fn record_decode_error(&self) {
        self.inner.metrics.record_decode_error();
        if let Ok(mut s) = self.inner.state.try_write() {
            s.decode_errors += 1;
        } else {
            self.inner
                .pending_producer
                .decode_errors
                .fetch_add(1, Ordering::Relaxed);
        }
    }

    pub async fn record_motion_analyzed(
        &self,
        score_percent: u32,
        detected: bool,
        latency_ms: u64,
    ) {
        self.inner
            .metrics
            .record_motion_analyzed(score_percent, detected, latency_ms);
        let mut s = self.inner.state.write().await;
        s.frames_motion_analyzed += 1;
        s.last_motion_score = score_percent as u64;
        s.last_motion_ms = latency_ms;
        if detected {
            s.motion_detected += 1;
        }
    }

    pub fn record_motion_error(&self) {
        self.inner.metrics.record_motion_error();
        if let Ok(mut s) = self.inner.state.try_write() {
            s.motion_errors += 1;
        } else {
            self.inner
                .pending_producer
                .motion_errors
                .fetch_add(1, Ordering::Relaxed);
        }
    }

    /// Atualiza métricas globais e estado da câmera com um único lock por frame processado.
    pub async fn record_consumer_frame(
        &self,
        frame: &PipelineFrame,
        outcome: ConsumerFrameOutcome,
    ) {
        let latency_ms = frame.captured_at.elapsed().as_millis() as u64;
        self.inner.metrics.record_processed(latency_ms);

        match outcome {
            ConsumerFrameOutcome::Decoded { decode_ms, motion } => {
                self.inner.metrics.record_decode_success(decode_ms);
                if let Some(m) = motion {
                    self.inner.metrics.record_motion_analyzed(
                        m.score_percent,
                        m.detected,
                        m.latency_ms,
                    );
                }
            }
            ConsumerFrameOutcome::DecodeFailed => {
                self.inner.metrics.record_decode_error();
            }
            ConsumerFrameOutcome::MotionFailed { decode_ms } => {
                self.inner.metrics.record_decode_success(decode_ms);
                self.inner.metrics.record_motion_error();
            }
            ConsumerFrameOutcome::ProcessedOnly | ConsumerFrameOutcome::NotReady => {}
        }

        let mut s = self.inner.state.write().await;
        s.frames_processed += 1;
        s.last_frame_latency_ms = latency_ms;

        match outcome {
            ConsumerFrameOutcome::Decoded { decode_ms, motion } => {
                s.frames_decoded += 1;
                s.last_decode_ms = decode_ms;
                if let Some(m) = motion {
                    s.frames_motion_analyzed += 1;
                    s.last_motion_score = m.score_percent as u64;
                    s.last_motion_ms = m.latency_ms;
                    if m.detected {
                        s.motion_detected += 1;
                    }
                }
            }
            ConsumerFrameOutcome::DecodeFailed => {
                s.decode_errors += 1;
            }
            ConsumerFrameOutcome::MotionFailed { decode_ms } => {
                s.frames_decoded += 1;
                s.last_decode_ms = decode_ms;
                s.motion_errors += 1;
            }
            ConsumerFrameOutcome::NotReady | ConsumerFrameOutcome::ProcessedOnly => {}
        }
    }
}

#[derive(Debug, Clone, Copy)]
pub struct ConsumerMotionStats {
    pub score_percent: u32,
    pub detected: bool,
    pub latency_ms: u64,
}

#[derive(Debug, Clone, Copy)]
pub enum ConsumerFrameOutcome {
    ProcessedOnly,
    NotReady,
    DecodeFailed,
    Decoded {
        decode_ms: u64,
        motion: Option<ConsumerMotionStats>,
    },
    MotionFailed {
        decode_ms: u64,
    },
}

/// Consumer: decode H.264 (Fase 3.1) + motion (Fase 3.2) + métricas de pipeline.
pub async fn run_frame_consumer(
    pipeline: FramePipeline,
    mut session_shutdown: tokio::sync::watch::Receiver<bool>,
    mut global_shutdown: tokio::sync::watch::Receiver<bool>,
    decode_ctx: Arc<SessionDecodeContext>,
    acceleration: Arc<AccelerationRuntime>,
    decode_policy: Arc<DecodePolicyCoordinator>,
    decode_enabled: bool,
) {
    let mut h264_decoder = H264Decoder::with_acceleration_and_policy(acceleration, decode_policy);
    let mut motion_detector = MotionDetector::new();
    let mut flush_every: u32 = 0;

    loop {
        if *global_shutdown.borrow() {
            break;
        }
        if *session_shutdown.borrow() {
            break;
        }

        tokio::select! {
            biased;
            _ = global_shutdown.changed() => {
                if *global_shutdown.borrow() { break; }
            }
            _ = session_shutdown.changed() => {
                if *session_shutdown.borrow() { break; }
            }
            frame = pipeline.dequeue() => {
                match frame {
                    Some(f) => {
                        flush_every = flush_every.wrapping_add(1);
                        if flush_every % 8 == 0 {
                            pipeline.flush_producer_stats().await;
                        }

                        let outcome = if decode_enabled {
                            if f.decoder_reset {
                                h264_decoder.discard_session_state();
                            }
                            let started = Instant::now();
                            let decode_result = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                                h264_decoder.decode(
                                    &decode_ctx,
                                    DecodeInput {
                                        payload: &f.payload,
                                        is_keyframe: f.is_keyframe,
                                        rtp_timestamp_ticks: Some(f.rtp_timestamp.timestamp),
                                    },
                                )
                            }));
                            match decode_result {
                                Ok(DecodeOutcome::Decoded(decoded)) => {
                                    let decode_ms = started.elapsed().as_millis() as u64;
                                    let motion_started = Instant::now();
                                    let motion_result = std::panic::catch_unwind(
                                        std::panic::AssertUnwindSafe(|| motion_detector.analyze(&decoded)),
                                    );
                                    match motion_result {
                                        Ok(MotionOutcome::ReferenceSet) => {
                                            tracing::debug!(seq = f.seq, "motion reference set");
                                            ConsumerFrameOutcome::Decoded {
                                                decode_ms,
                                                motion: Some(ConsumerMotionStats {
                                                    score_percent: 0,
                                                    detected: false,
                                                    latency_ms: motion_started
                                                        .elapsed()
                                                        .as_millis() as u64,
                                                }),
                                            }
                                        }
                                        Ok(MotionOutcome::Analyzed {
                                            detected,
                                            score_percent,
                                        }) => {
                                            if detected {
                                                tracing::debug!(
                                                    seq = f.seq,
                                                    score_percent,
                                                    "motion detected"
                                                );
                                            }
                                            ConsumerFrameOutcome::Decoded {
                                                decode_ms,
                                                motion: Some(ConsumerMotionStats {
                                                    score_percent,
                                                    detected,
                                                    latency_ms: motion_started
                                                        .elapsed()
                                                        .as_millis() as u64,
                                                }),
                                            }
                                        }
                                        Ok(MotionOutcome::Error) => {
                                            tracing::debug!(seq = f.seq, "motion luma error");
                                            ConsumerFrameOutcome::MotionFailed { decode_ms }
                                        }
                                        Err(_) => {
                                            tracing::warn!(seq = f.seq, "motion analyze panicked");
                                            ConsumerFrameOutcome::MotionFailed { decode_ms }
                                        }
                                    }
                                }
                                Ok(DecodeOutcome::NotReady) => ConsumerFrameOutcome::NotReady,
                                Ok(DecodeOutcome::Failed(e)) => {
                                    tracing::debug!(error = %e, seq = f.seq, "h264 decode");
                                    ConsumerFrameOutcome::DecodeFailed
                                }
                                Err(_) => {
                                    tracing::warn!(seq = f.seq, "h264 decode panicked");
                                    ConsumerFrameOutcome::DecodeFailed
                                }
                            }
                        } else {
                            ConsumerFrameOutcome::ProcessedOnly
                        };

                        pipeline.record_consumer_frame(&f, outcome).await;
                    }
                    None if pipeline.is_closed() => break,
                    None => continue,
                }
            }
        }
    }

    pipeline.flush_producer_stats().await;
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::sync::atomic::AtomicU64;
    use std::sync::Arc;
    use tokio::sync::RwLock;

    use crate::camera::{CameraRuntimeState, SharedCameraState};
    use crate::decode::{
        AccelerationPolicy, AccelerationRuntime, VideoAccelerationMode, VideoGpuBackend,
    };

    fn test_state(id: i64) -> SharedCameraState {
        Arc::new(RwLock::new(CameraRuntimeState::new(
            id,
            "rtsp://test".into(),
        )))
    }

    fn test_acceleration_runtime() -> Arc<AccelerationRuntime> {
        AccelerationRuntime::bootstrap(AccelerationPolicy::from_parts(
            VideoAccelerationMode::Cpu,
            VideoGpuBackend::Auto,
        ))
        .unwrap()
    }

    use crate::pipeline::{fake_h264_payload, RtpTimestamp};

    fn test_rtp() -> RtpTimestamp {
        RtpTimestamp {
            timestamp: 90_000,
            clock_rate_hz: 90_000,
            stream_start: 0,
        }
    }

    fn test_frame(seq: u64, tag: u8, keyframe: bool) -> PipelineFrame {
        PipelineFrame::new(seq, fake_h264_payload(seq, tag), keyframe, test_rtp())
    }

    fn test_metrics() -> Arc<ProcessorMetrics> {
        Arc::new(ProcessorMetrics::new("test-pipeline"))
    }

    #[test]
    fn empty_buffer() {
        let q = DropOldestQueue::new(3);
        assert!(q.is_empty());
        assert_eq!(q.len(), 0);
        assert_eq!(q.capacity(), 3);
    }

    #[test]
    fn enqueue_dequeue_normal() {
        let mut q = DropOldestQueue::new(3);
        q.push(test_frame(1, 0xAA, false));
        q.push(test_frame(2, 0xBB, true));
        assert_eq!(q.len(), 2);
        let f1 = q.pop().unwrap();
        assert_eq!(f1.seq, 1);
        assert_eq!(f1.payload[0], 0xAA);
        let f2 = q.pop().unwrap();
        assert_eq!(f2.seq, 2);
        assert!(f2.is_keyframe);
        assert!(q.is_empty());
    }

    #[test]
    fn drop_oldest_when_full() {
        let mut q = DropOldestQueue::new(2);
        for seq in 1..=3 {
            q.push(test_frame(seq, seq as u8, seq == 3));
        }
        assert_eq!(q.dropped, 1);
        assert_eq!(q.full_events, 1);
        assert_eq!(q.enqueued, 3);
        assert_eq!(q.len(), 2);
        let f2 = q.pop().unwrap();
        assert_eq!(f2.seq, 2);
        assert_eq!(f2.payload[0], 2);
        let f3 = q.pop().unwrap();
        assert_eq!(f3.seq, 3);
        assert!(f3.is_keyframe);
    }

    #[test]
    fn dropped_counter_increments() {
        let mut q = DropOldestQueue::new(1);
        q.push(test_frame(1, 1, false));
        q.push(test_frame(2, 2, false));
        assert_eq!(q.dropped, 1);
    }

    #[test]
    fn payload_not_corrupted_in_queue() {
        let mut q = DropOldestQueue::new(2);
        q.push(test_frame(10, 0xDE, false));
        q.push(test_frame(11, 0xAD, true));
        let a = q.pop().unwrap();
        let b = q.pop().unwrap();
        assert_eq!(&*a.payload, &*fake_h264_payload(10, 0xDE));
        assert_eq!(&*b.payload, &*fake_h264_payload(11, 0xAD));
        assert!(!a.is_keyframe);
        assert!(b.is_keyframe);
    }

    #[test]
    fn rtp_timestamp_preserved() {
        let rtp = RtpTimestamp {
            timestamp: 123_456,
            clock_rate_hz: 90_000,
            stream_start: 42,
        };
        let frame = PipelineFrame::new(1, fake_h264_payload(1, 0x01), false, rtp);
        let mut q = DropOldestQueue::new(1);
        q.push(frame);
        let out = q.pop().unwrap();
        assert_eq!(out.rtp_timestamp, rtp);
    }

    #[tokio::test]
    async fn processed_increments_via_pipeline() {
        let metrics = test_metrics();
        let state = test_state(1);
        let pipeline = FramePipeline::new(2, metrics.clone(), state.clone());
        pipeline.try_enqueue(test_frame(1, 0x01, false));
        let frame = pipeline.dequeue().await.unwrap();
        pipeline.record_processed(&frame).await;
        assert_eq!(metrics.frames_processed.load(Ordering::Relaxed), 1);
        assert_eq!(state.read().await.frames_processed, 1);
    }

    #[tokio::test]
    async fn buffer_size_tracks_len() {
        let metrics = test_metrics();
        let state = test_state(2);
        let pipeline = FramePipeline::new(3, metrics, state.clone());
        pipeline.try_enqueue(test_frame(1, 1, false));
        pipeline.try_enqueue(test_frame(2, 2, false));
        assert_eq!(pipeline.len(), 2);
        assert_eq!(state.try_read().unwrap().buffer_size, 2);
    }

    #[tokio::test]
    async fn close_clears_queue() {
        let metrics = test_metrics();
        let pipeline = FramePipeline::new(2, metrics, test_state(3));
        pipeline.try_enqueue(test_frame(1, 1, false));
        pipeline.try_enqueue(test_frame(2, 2, false));
        pipeline.close();
        assert_eq!(pipeline.len(), 0);
        assert!(pipeline.is_closed());
    }

    #[tokio::test]
    async fn cameras_isolated() {
        let metrics = test_metrics();
        let p1 = FramePipeline::new(2, metrics.clone(), test_state(1));
        let p2 = FramePipeline::new(2, metrics.clone(), test_state(2));
        p1.try_enqueue(test_frame(10, 0x10, false));
        p2.try_enqueue(test_frame(20, 0x20, false));
        p2.try_enqueue(test_frame(21, 0x21, true));
        assert_eq!(p1.len(), 1);
        assert_eq!(p2.len(), 2);
    }

    #[tokio::test]
    async fn fast_producer_slow_consumer_drops() {
        let metrics = test_metrics();
        let pipeline = FramePipeline::new(2, metrics.clone(), test_state(1));
        for seq in 1..=10 {
            pipeline.try_enqueue(test_frame(seq, seq as u8, false));
        }
        assert_eq!(pipeline.len(), 2);
        assert!(metrics.frames_dropped.load(Ordering::Relaxed) >= 8);
        assert!(metrics.buffer_full_events.load(Ordering::Relaxed) >= 8);
    }

    #[tokio::test]
    async fn reconnect_new_pipeline_empty() {
        let metrics = test_metrics();
        let old = FramePipeline::new(2, metrics.clone(), test_state(1));
        old.try_enqueue(test_frame(1, 1, false));
        old.close();
        let new_pipe = FramePipeline::new(2, metrics, test_state(1));
        assert_eq!(new_pipe.len(), 0);
        assert!(!new_pipe.is_closed());
    }

    #[tokio::test]
    async fn consumer_drains_without_deadlock() {
        let metrics = test_metrics();
        let pipeline = FramePipeline::new(4, metrics.clone(), test_state(1));
        let (tx, rx) = tokio::sync::watch::channel(false);
        let (gtx, grx) = tokio::sync::watch::channel(false);
        let p = pipeline.clone();
        let decode_ctx = SessionDecodeContext::new();
        let acceleration = test_acceleration_runtime();
        let consumer = tokio::spawn(async move {
            let decode_policy =
                crate::decode::DecodePolicyCoordinator::new(crate::decode::DecodeFallbackConfig {
                    runtime_fallback_enabled: true,
                    hw_error_threshold: 10,
                });
            run_frame_consumer(p, rx, grx, decode_ctx, acceleration, decode_policy, false).await;
        });
        for seq in 1..=5 {
            pipeline.try_enqueue(test_frame(seq, seq as u8, false));
        }
        tokio::time::sleep(std::time::Duration::from_millis(50)).await;
        pipeline.close();
        tx.send(true).unwrap();
        tokio::time::timeout(std::time::Duration::from_secs(2), consumer)
            .await
            .expect("deadlock")
            .unwrap();
        assert!(metrics.frames_processed.load(Ordering::Relaxed) >= 1);
        let _ = gtx;
    }
}
