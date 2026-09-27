//! Fase D2/D3 — fila de eventos analíticos (Redis / memory), compatível com worker Python.

mod job;
mod queue;

pub use job::EventJob;
pub use queue::{EventQueueHandle, EventQueueStats, PublishOutcome};

use std::sync::Arc;
use std::time::Duration;

use tracing::info;

use crate::config::Config;
use crate::error::AppResult;

pub fn bootstrap_event_queue(cfg: &Config) -> AppResult<Arc<EventQueueHandle>> {
    let handle = Arc::new(EventQueueHandle::from_config(cfg)?);
    info!(
        queue_backend = handle.backend_label(),
        event_queue_key = %cfg.event_queue_key,
        dlq_key = %cfg.event_queue_dlq_key,
        max_size = cfg.event_queue_max_size,
        "event queue initialized"
    );
    Ok(handle)
}

pub async fn run_queue_metrics_loop(queue: Arc<EventQueueHandle>, interval: Duration) {
    if queue.backend_label() != "redis" {
        return;
    }
    let mut tick = tokio::time::interval(interval);
    tick.set_missed_tick_behavior(tokio::time::MissedTickBehavior::Skip);
    loop {
        tick.tick().await;
        let _ = queue.ping_redis().await;
        queue.refresh_depths().await;
    }
}
