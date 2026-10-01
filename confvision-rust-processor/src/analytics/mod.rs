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

    pub async fn upsert_cameras(&self, cameras: &[CameraRecord], areas: &[Value]) {
        let mut by_camera: HashMap<i64, Vec<Value>> = HashMap::new();
        for area in areas {
            let Some(cid) = area
                .get("vis_camera_id")
                .and_then(|v| v.as_i64().or_else(|| v.as_u64().map(|u| u as i64)))
            else {
                continue;
            };
            by_camera.entry(cid).or_default().push(area.clone());
        }

        let mut map = self.camera_configs.write().await;
        for cam in cameras {
            if let Ok(mut v) = serde_json::to_value(cam) {
                if let Some(zones) = by_camera.get(&cam.id) {
                    v["areas"] = Value::Array(zones.clone());
                }
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
