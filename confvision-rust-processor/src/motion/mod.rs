mod detector;
mod gated_session;
mod hooks;
mod recording;
mod session;
mod shadow;
mod throttle;
mod types;

pub use detector::MotionDetector;
pub use gated_session::MotionGatedSession;
pub use hooks::{MotionHookConfig, MotionPipelineHooks};
pub use session::{MotionSessionEvent, MotionSessionEventKind, MotionSessionTracker};
pub use throttle::{MotionAnalysisThrottle, MotionEnqueueDecision, MotionEnqueueGate};
pub use types::{MotionOutcome, MotionSensitivity};
