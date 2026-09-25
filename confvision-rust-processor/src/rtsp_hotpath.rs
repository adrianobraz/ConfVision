use std::sync::atomic::{AtomicU64, Ordering};

use serde::Serialize;

/// Chamadas `session.next()` com duração abaixo disto contam como retorno imediato (sem espera I/O).
const SESSION_NEXT_FAST_MAX_NANOS: u64 = 500_000;

/// Acumuladores atômicos do hot path RTSP (Fase 6.7 — diagnóstico, sem locks).
#[derive(Debug, Default)]
pub struct RtspHotpathStats {
    session_next_calls: AtomicU64,
    session_next_nanos: AtomicU64,
    session_next_fast_calls: AtomicU64,
    session_next_non_video_items: AtomicU64,
    video_au_count: AtomicU64,
    post_au_nanos: AtomicU64,
    throttle_nanos: AtomicU64,
    metrics_nanos: AtomicU64,
    loop_iter_count: AtomicU64,
    loop_iter_nanos: AtomicU64,
}

impl RtspHotpathStats {
    pub fn record_session_next(&self, nanos: u64) {
        self.session_next_calls.fetch_add(1, Ordering::Relaxed);
        self.session_next_nanos.fetch_add(nanos, Ordering::Relaxed);
        if nanos < SESSION_NEXT_FAST_MAX_NANOS {
            self.session_next_fast_calls.fetch_add(1, Ordering::Relaxed);
        }
    }

    pub fn record_session_next_non_video(&self) {
        self.session_next_non_video_items
            .fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_video_au(&self, post_au_nanos: u64, throttle_nanos: u64, metrics_nanos: u64) {
        self.video_au_count.fetch_add(1, Ordering::Relaxed);
        self.post_au_nanos
            .fetch_add(post_au_nanos, Ordering::Relaxed);
        self.throttle_nanos
            .fetch_add(throttle_nanos, Ordering::Relaxed);
        self.metrics_nanos
            .fetch_add(metrics_nanos, Ordering::Relaxed);
    }

    pub fn record_loop_iter(&self, nanos: u64) {
        self.loop_iter_count.fetch_add(1, Ordering::Relaxed);
        self.loop_iter_nanos.fetch_add(nanos, Ordering::Relaxed);
    }

    pub fn snapshot(&self) -> RtspHotpathMetrics {
        let session_next_calls = self.session_next_calls.load(Ordering::Relaxed);
        let session_next_nanos = self.session_next_nanos.load(Ordering::Relaxed);
        let session_next_fast_calls = self.session_next_fast_calls.load(Ordering::Relaxed);
        let session_next_non_video_items =
            self.session_next_non_video_items.load(Ordering::Relaxed);
        let video_au_count = self.video_au_count.load(Ordering::Relaxed);
        let post_au_nanos = self.post_au_nanos.load(Ordering::Relaxed);
        let throttle_nanos = self.throttle_nanos.load(Ordering::Relaxed);
        let metrics_nanos = self.metrics_nanos.load(Ordering::Relaxed);
        let loop_iter_count = self.loop_iter_count.load(Ordering::Relaxed);
        let loop_iter_nanos = self.loop_iter_nanos.load(Ordering::Relaxed);

        let loop_overhead_nanos =
            loop_iter_nanos.saturating_sub(session_next_nanos + post_au_nanos);

        let ms = |n: u64| n as f64 / 1_000_000.0;
        let us_avg = |total_nanos: u64, count: u64| {
            if count == 0 {
                0.0
            } else {
                (total_nanos as f64 / count as f64) / 1_000.0
            }
        };

        let denom = (session_next_nanos + post_au_nanos + loop_overhead_nanos).max(1) as f64;
        let pct = |part: u64| (part as f64 / denom) * 100.0;

        let session_next_slow_calls = session_next_calls.saturating_sub(session_next_fast_calls);

        RtspHotpathMetrics {
            session_next_calls,
            session_next_elapsed_ms: ms(session_next_nanos),
            session_next_avg_us: us_avg(session_next_nanos, session_next_calls),
            session_next_fast_calls,
            session_next_slow_calls,
            session_next_non_video_items,
            video_au_count,
            post_au_elapsed_ms: ms(post_au_nanos),
            post_au_avg_us: us_avg(post_au_nanos, video_au_count),
            throttle_elapsed_ms: ms(throttle_nanos),
            metrics_elapsed_ms: ms(metrics_nanos),
            loop_iter_count,
            loop_elapsed_ms: ms(loop_iter_nanos),
            loop_overhead_elapsed_ms: ms(loop_overhead_nanos),
            session_next_share_percent: pct(session_next_nanos),
            post_au_share_percent: pct(post_au_nanos),
            throttle_share_percent: pct(throttle_nanos),
            metrics_share_percent: pct(metrics_nanos),
            loop_overhead_share_percent: pct(loop_overhead_nanos),
        }
    }
}

#[derive(Debug, Clone, Serialize)]
pub struct RtspHotpathMetrics {
    pub session_next_calls: u64,
    pub session_next_elapsed_ms: f64,
    pub session_next_avg_us: f64,
    /// Retornos rápidos (menos de 500µs): demux/RTCP sem bloquear em I/O.
    pub session_next_fast_calls: u64,
    /// Retornos lentos: tipicamente inclui espera por dados TCP.
    pub session_next_slow_calls: u64,
    pub session_next_non_video_items: u64,
    pub video_au_count: u64,
    pub post_au_elapsed_ms: f64,
    pub post_au_avg_us: f64,
    pub throttle_elapsed_ms: f64,
    pub metrics_elapsed_ms: f64,
    pub loop_iter_count: u64,
    pub loop_elapsed_ms: f64,
    pub loop_overhead_elapsed_ms: f64,
    pub session_next_share_percent: f64,
    pub post_au_share_percent: f64,
    pub throttle_share_percent: f64,
    pub metrics_share_percent: f64,
    pub loop_overhead_share_percent: f64,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn snapshot_computes_shares() {
        let s = RtspHotpathStats::default();
        s.record_session_next(600);
        s.record_video_au(300, 100, 200);
        s.record_loop_iter(1000);
        let snap = s.snapshot();
        assert_eq!(snap.session_next_calls, 1);
        assert_eq!(snap.video_au_count, 1);
        assert!(snap.session_next_share_percent > 50.0);
    }
}
