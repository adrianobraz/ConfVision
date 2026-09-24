//! S3 / mídia — reservado para clips e snapshots (fase posterior).

use crate::config::Config;

pub fn log_media_status(cfg: &Config) {
    if cfg.s3_endpoint.is_some() && cfg.s3_bucket.is_some() {
        tracing::info!("s3 configurado (uso futuro)");
    }
}
