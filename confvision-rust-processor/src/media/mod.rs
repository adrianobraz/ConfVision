//! S3 / mídia — upload de snapshots e clips (D3).

mod upload;

pub use upload::{evento_clip_key, evento_snapshot_key, upload_file};

use crate::config::Config;

pub fn log_media_status(cfg: &Config) {
    if cfg.s3_ready() {
        tracing::info!("s3 configurado para upload de eventos");
    }
}
