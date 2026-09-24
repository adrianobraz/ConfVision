//! Redis reservado para cache/filas (fase distributed).
//! Config `REDIS_URL` aceita; conexão não é obrigatória na fase 1.

use crate::config::Config;

pub fn log_redis_status(cfg: &Config) {
    if cfg.redis_url.is_some() {
        tracing::info!("redis configurado (uso futuro)");
    }
}
