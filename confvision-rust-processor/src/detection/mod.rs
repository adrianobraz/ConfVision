mod areas;
mod coordinator;
mod rules;

pub use areas::{areas_from_camera, AreaZone};
pub use coordinator::{DetectionContext, DetectionStats};
pub use rules::{evaluate_detections, MatchResult};
