mod frame_payload;
mod frame_pipeline;

pub use frame_payload::{PipelineFrame, RtpTimestamp};
pub use frame_pipeline::{run_frame_consumer, FramePipeline};

#[cfg(test)]
pub use frame_payload::fake_h264_payload;
