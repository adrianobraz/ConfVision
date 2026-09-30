//! Fachada do fluxo Detection → EventJob → fila (Fase 3).
//!
//! Hoje a orquestração principal permanece em `detection::DetectionContext`.
//! Este módulo concentra publicação na fila e campos de correlação em logs.

use tracing::info;

use super::catalog::ConfVisionEventKind;
use super::job::EventJob;
use super::queue::{EventQueueHandle, PublishOutcome};

pub struct EventPublishResult {
    pub outcome: PublishOutcome,
    pub kind: ConfVisionEventKind,
}

pub async fn publish_analytic_detection(
    queue: &EventQueueHandle,
    camera_id: i64,
    job: EventJob,
    best_conf: f64,
) -> EventPublishResult {
    let detected_at = job.detected_at;
    let outcome = queue.publish(job).await;
    if outcome == PublishOutcome::Ok {
        info!(
            camera_id,
            conf = best_conf,
            detected_at,
            event_kind = ConfVisionEventKind::AnalyticDetection.as_str(),
            component = "events",
            operation = "publish",
            "event job enqueued"
        );
    }
    EventPublishResult {
        outcome,
        kind: ConfVisionEventKind::AnalyticDetection,
    }
}
