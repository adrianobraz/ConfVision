mod acceleration;
mod context;
mod error;
mod h264_decoder;
mod luma;
mod types;

#[cfg(feature = "ffmpeg-decode")]
pub(crate) mod backend;

#[cfg(feature = "ffmpeg-decode")]
mod ffmpeg_backend;

#[cfg(test)]
mod fixtures;

pub use acceleration::{
    AccelerationPolicy, AccelerationRuntime, VideoAccelerationMode, VideoGpuBackend,
};
pub use context::SessionDecodeContext;
pub use error::DecodeError;
pub use h264_decoder::H264Decoder;
pub use types::{
    DecodeInput, DecodeOutcome, DecodedFrame, PixelFormat, DECODED_LUMA_HEIGHT, DECODED_LUMA_WIDTH,
};

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn decode_without_extradata_is_not_ready() {
        let mut decoder = H264Decoder::new();
        let ctx = SessionDecodeContext::new();
        let out = decoder.decode(
            &ctx,
            DecodeInput {
                payload: &[0x00, 0x00, 0x00, 0x01],
                is_keyframe: true,
                rtp_timestamp_ticks: None,
            },
        );
        assert!(matches!(out, DecodeOutcome::NotReady));
    }

    #[cfg(feature = "ffmpeg-decode")]
    #[test]
    fn empty_payload_is_decode_error_not_panic() {
        let mut decoder = H264Decoder::new();
        let ctx = SessionDecodeContext::new();
        ctx.update_extradata(crate::decode::fixtures::minimal::EXTRADATA);
        let out = decoder.decode(
            &ctx,
            DecodeInput {
                payload: &[],
                is_keyframe: true,
                rtp_timestamp_ticks: Some(9000),
            },
        );
        assert!(matches!(out, DecodeOutcome::Failed(_)));
    }

    #[cfg(feature = "ffmpeg-decode")]
    #[test]
    fn decoder_accepts_extradata_and_invalid_au() {
        let mut decoder = H264Decoder::new();
        let ctx = SessionDecodeContext::new();
        ctx.update_extradata(crate::decode::fixtures::minimal::EXTRADATA);
        let garbage = [0x00, 0x00, 0x00, 0x05, 0x65, 0x88, 0x84, 0x00, 0x10];
        let out = decoder.decode(
            &ctx,
            DecodeInput {
                payload: &garbage,
                is_keyframe: true,
                rtp_timestamp_ticks: None,
            },
        );
        assert!(matches!(
            out,
            DecodeOutcome::Failed(_) | DecodeOutcome::NotReady
        ));
    }
}
