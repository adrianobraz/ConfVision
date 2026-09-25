mod detector;
mod gated_session;
mod throttle;
mod types;

pub use detector::MotionDetector;
pub use gated_session::MotionGatedSession;
pub use throttle::{MotionAnalysisThrottle, MotionEnqueueDecision, MotionEnqueueGate};
pub use types::MotionOutcome;
