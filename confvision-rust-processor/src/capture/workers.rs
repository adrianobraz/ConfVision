use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::Arc;

use tracing::{debug, warn};

use crate::api::ConfVisionClient;
use crate::config::Config;
use crate::events::EventQueueHandle;

use super::locks::CaptureLocks;
use super::process::processar_deteccao;

#[derive(Default)]
pub struct CaptureStats {
    pub processed: AtomicU64,
    pub discarded: AtomicU64,
}

pub fn spawn_capture_workers(
    cfg: Arc<Config>,
    client: Arc<ConfVisionClient>,
    queue: Arc<EventQueueHandle>,
    locks: Arc<CaptureLocks>,
    stats: Arc<CaptureStats>,
) {
    if !cfg.capture_enabled {
        return;
    }
    let n = cfg.capture_workers.max(1);
    for worker_no in 1..=n {
        let cfg = cfg.clone();
        let client = client.clone();
        let queue = queue.clone();
        let locks = locks.clone();
        let stats = stats.clone();
        tokio::spawn(async move {
            capture_worker_loop(worker_no as u32, cfg, client, queue, locks, stats).await;
        });
    }
}

async fn capture_worker_loop(
    worker_no: u32,
    cfg: Arc<Config>,
    client: Arc<ConfVisionClient>,
    queue: Arc<EventQueueHandle>,
    locks: Arc<CaptureLocks>,
    stats: Arc<CaptureStats>,
) {
    loop {
        match queue.pop(1.0).await {
            Ok(Some(job)) => {
                let camera_id = job.camera.get("id").and_then(|v| v.as_i64());
                let Some(camera_id) = camera_id else {
                    stats.discarded.fetch_add(1, Ordering::Relaxed);
                    continue;
                };
                let Some(guard) = locks.try_acquire(camera_id).await else {
                    debug!(
                        worker_no,
                        camera_id, "captura em andamento — job descartado"
                    );
                    stats.discarded.fetch_add(1, Ordering::Relaxed);
                    if let Some(p) = &job.snapshot_path {
                        let _ = std::fs::remove_file(p);
                    }
                    continue;
                };
                debug!(worker_no, camera_id, conf = job.confianca, "capture job");
                processar_deteccao(&cfg, &client, &job, guard).await;
                stats.processed.fetch_add(1, Ordering::Relaxed);
            }
            Ok(None) => {}
            Err(e) => {
                warn!(worker_no, error = %e, "event queue pop");
                tokio::time::sleep(std::time::Duration::from_millis(500)).await;
            }
        }
    }
}
