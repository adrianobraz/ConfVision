//! Redis — ping e status (Fase D2).

use crate::config::Config;
use crate::events::EventQueueHandle;

pub fn log_redis_status(cfg: &Config) {
    match cfg.queue_backend.as_str() {
        "redis" => {
            if cfg.redis_url.is_some() {
                tracing::info!(
                    queue_backend = "redis",
                    event_queue_key = %cfg.event_queue_key,
                    "redis configurado para fila de eventos (D2)"
                );
            } else {
                tracing::warn!("QUEUE_BACKEND=redis sem REDIS_URL");
            }
        }
        "memory" => tracing::info!(queue_backend = "memory", "fila de eventos em memória (dev)"),
        _ => {
            if cfg.redis_url.is_some() {
                tracing::info!("REDIS_URL definida; QUEUE_BACKEND=none (fila desligada)");
            }
        }
    }
}

pub async fn startup_redis_check(queue: &EventQueueHandle) {
    if queue.backend_label() != "redis" {
        return;
    }
    if queue.ping_redis().await {
        queue.refresh_depths().await;
        tracing::info!("redis PING ok (event queue)");
    } else {
        tracing::error!("redis PING falhou — verifique REDIS_URL e rede foxpro → Redis");
    }
}
