use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::Arc;

use serde_json::Value;
use tokio::sync::Mutex;
use tracing::{debug, info, warn};

use crate::capture::write_detection_snapshot;
use crate::config::Config;
use crate::decode::DecodedFrame;
use crate::events::{EventJob, EventQueueHandle, PublishOutcome};
use crate::motion::MotionGatedSession;
use crate::yolo::YoloRuntime;

use super::areas::areas_from_camera;
use super::rules::{
    camera_conf_min, camera_cooldown_sec, camera_deteccao_humano, camera_modo, evaluate_detections,
};

#[derive(Default)]
pub struct DetectionStats {
    pub yolo_inferences: AtomicU64,
    pub matches: AtomicU64,
    pub events_published: AtomicU64,
    pub events_queue_full: AtomicU64,
}

pub struct DetectionContext {
    camera_id: i64,
    camera: Value,
    cfg: Arc<Config>,
    yolo: Arc<YoloRuntime>,
    queue: Arc<EventQueueHandle>,
    stats: Arc<DetectionStats>,
    motion_gate: Option<Arc<MotionGatedSession>>,
    frame_index: AtomicU64,
    last_emit_epoch: Mutex<f64>,
    yolo_miss_streak: Mutex<u32>,
}

impl DetectionContext {
    pub fn new(
        camera_id: i64,
        camera: Value,
        cfg: Arc<Config>,
        yolo: Arc<YoloRuntime>,
        queue: Arc<EventQueueHandle>,
        stats: Arc<DetectionStats>,
        motion_gate: Option<Arc<MotionGatedSession>>,
    ) -> Arc<Self> {
        Arc::new(Self {
            camera_id,
            camera,
            cfg,
            yolo,
            queue,
            stats,
            motion_gate,
            frame_index: AtomicU64::new(0),
            last_emit_epoch: Mutex::new(0.0),
            yolo_miss_streak: Mutex::new(0),
        })
    }

    pub async fn on_decoded_frame(self: &Arc<Self>, decoded: &DecodedFrame, motion_detected: bool) {
        if !self.cfg.yolo_enabled || !self.yolo.is_active() {
            return;
        }
        if !camera_deteccao_humano(&self.camera) {
            return;
        }

        let idx = self.frame_index.fetch_add(1, Ordering::Relaxed);
        if self.cfg.yolo_frame_stride > 1 && idx % self.cfg.yolo_frame_stride as u64 != 0 {
            return;
        }

        if self.cfg.analysis_only_on_motion {
            let armed = self
                .motion_gate
                .as_ref()
                .map(|g| g.is_armed())
                .unwrap_or(true);
            if !armed && !motion_detected {
                return;
            }
        }

        let jpeg = match crate::capture::snapshot::luma_to_jpeg_bytes(
            &decoded.luma,
            decoded.luma_width,
            decoded.luma_height,
            self.cfg.snapshot_jpeg_quality,
        ) {
            Ok(b) => b,
            Err(e) => {
                warn!(camera_id = self.camera_id, error = %e, "jpeg encode");
                return;
            }
        };

        self.stats.yolo_inferences.fetch_add(1, Ordering::Relaxed);
        let detections = match self.yolo.infer_jpeg(&jpeg).await {
            Ok(d) => d,
            Err(e) => {
                warn!(camera_id = self.camera_id, error = %e, "yolo infer");
                return;
            }
        };

        let areas = areas_from_camera(&self.camera);
        let modo = camera_modo(&self.camera);
        let conf_min = camera_conf_min(&self.camera, self.cfg.yolo_conf_default);
        let (result, _area) = evaluate_detections(
            &detections,
            conf_min,
            &areas,
            modo,
            decoded.luma_width as f64,
            decoded.luma_height as f64,
        );

        if result.pessoas_match == 0 || result.best_conf <= 0.0 {
            let mut miss = self.yolo_miss_streak.lock().await;
            *miss += 1;
            return;
        }
        {
            let mut miss = self.yolo_miss_streak.lock().await;
            *miss = 0;
        }

        self.stats.matches.fetch_add(1, Ordering::Relaxed);

        let now = chrono::Utc::now().timestamp_millis() as f64 / 1000.0;
        let cooldown = camera_cooldown_sec(&self.camera);
        {
            let mut last = self.last_emit_epoch.lock().await;
            if now - *last < cooldown as f64 {
                debug!(camera_id = self.camera_id, "cooldown ativo");
                return;
            }
            *last = now;
        }

        let snapshot_path = write_detection_snapshot(
            self.camera_id,
            &decoded.luma,
            &self.cfg.capture_dir,
            self.cfg.snapshot_jpeg_quality,
        )
        .ok()
        .map(|p| p.to_string_lossy().into_owned());

        let job = EventJob::from_camera(&self.camera, result.best_conf, snapshot_path.clone());
        match self.queue.publish(job).await {
            PublishOutcome::Ok => {
                self.stats.events_published.fetch_add(1, Ordering::Relaxed);
                info!(
                    camera_id = self.camera_id,
                    conf = result.best_conf,
                    "evento publicado na fila"
                );
            }
            PublishOutcome::QueueFull => {
                self.stats.events_queue_full.fetch_add(1, Ordering::Relaxed);
                warn!(camera_id = self.camera_id, "fila cheia");
                if let Some(p) = snapshot_path {
                    let _ = std::fs::remove_file(p);
                }
            }
            other => {
                warn!(camera_id = self.camera_id, ?other, "fila publish falhou");
            }
        }
    }
}
