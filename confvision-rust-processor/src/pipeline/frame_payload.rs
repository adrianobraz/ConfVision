use std::sync::Arc;
use std::time::Instant;

/// Timestamp RTP copiado do stream (independente de `retina` no restante do pipeline).
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct RtpTimestamp {
    pub timestamp: i64,
    pub clock_rate_hz: u32,
    pub stream_start: u32,
}

/// Frame na fila bounded — bitstream H.264/H.26x codificado (Fase 3.0).
#[derive(Clone)]
pub struct PipelineFrame {
    pub seq: u64,
    pub captured_at: Instant,
    pub payload: Arc<[u8]>,
    pub is_keyframe: bool,
    pub rtp_timestamp: RtpTimestamp,
    /// Reinicializar libavcodec antes de decodificar (após AUs ignorados no RTSP).
    pub decoder_reset: bool,
}

impl PipelineFrame {
    pub fn new(
        seq: u64,
        payload: Arc<[u8]>,
        is_keyframe: bool,
        rtp_timestamp: RtpTimestamp,
    ) -> Self {
        Self::with_decoder_reset(seq, payload, is_keyframe, rtp_timestamp, false)
    }

    pub fn with_decoder_reset(
        seq: u64,
        payload: Arc<[u8]>,
        is_keyframe: bool,
        rtp_timestamp: RtpTimestamp,
        decoder_reset: bool,
    ) -> Self {
        Self {
            seq,
            captured_at: Instant::now(),
            payload,
            is_keyframe,
            rtp_timestamp,
            decoder_reset,
        }
    }
}

impl std::fmt::Debug for PipelineFrame {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("PipelineFrame")
            .field("seq", &self.seq)
            .field("payload_len", &self.payload.len())
            .field("is_keyframe", &self.is_keyframe)
            .field("rtp_timestamp", &self.rtp_timestamp)
            .finish_non_exhaustive()
    }
}

#[cfg(test)]
pub fn fake_h264_payload(seq: u64, tag: u8) -> Arc<[u8]> {
    Arc::from([tag, (seq >> 8) as u8, seq as u8, 0xFF].as_slice())
}
