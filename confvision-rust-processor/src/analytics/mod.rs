use std::collections::HashMap;
use std::sync::Arc;

use serde_json::Value;
use tokio::sync::RwLock;

use crate::api::CameraRecord;
use crate::capture::{CaptureLocks, CaptureStats};
use crate::config::Config;
use crate::detection::DetectionStats;
use crate::events::EventQueueHandle;
use crate::yolo::YoloRuntime;

#[derive(Clone)]
pub struct AnalyticsRuntime {
    pub cfg: Arc<Config>,
    pub yolo: Arc<YoloRuntime>,
    pub event_queue: Arc<EventQueueHandle>,
    pub detection_stats: Arc<DetectionStats>,
    pub capture_stats: Arc<CaptureStats>,
    pub capture_locks: Arc<CaptureLocks>,
    pub camera_configs: Arc<RwLock<HashMap<i64, Value>>>,
}

impl AnalyticsRuntime {
    pub fn bootstrap(
        cfg: Arc<Config>,
        yolo: Arc<YoloRuntime>,
        event_queue: Arc<EventQueueHandle>,
    ) -> Arc<Self> {
        Arc::new(Self {
            cfg,
            yolo,
            event_queue,
            detection_stats: Arc::new(DetectionStats::default()),
            capture_stats: Arc::new(CaptureStats::default()),
            capture_locks: CaptureLocks::new(),
            camera_configs: Arc::new(RwLock::new(HashMap::new())),
        })
    }

    pub async fn upsert_cameras(&self, cameras: &[CameraRecord]) {
        let mut map = self.camera_configs.write().await;
        for cam in cameras {
            if let Ok(v) = serde_json::to_value(cam) {
                map.insert(cam.id, v);
            }
        }
    }

    pub async fn remove_camera(&self, camera_id: i64) {
        self.camera_configs.write().await.remove(&camera_id);
    }

    pub async fn camera_json(&self, camera_id: i64) -> Option<Value> {
        self.camera_configs.read().await.get(&camera_id).cloned()
    }
}
