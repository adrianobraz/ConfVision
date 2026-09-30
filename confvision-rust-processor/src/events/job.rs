use serde::{Deserialize, Serialize};
use serde_json::{Map, Value};

/// Payload compatível com `confvision/event_queue.py` (`EventJob.to_dict`).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EventJob {
    pub camera: Map<String, Value>,
    pub confianca: f64,
    #[serde(default = "default_detected_at")]
    pub detected_at: f64,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub snapshot_path: Option<String>,
}

fn default_detected_at() -> f64 {
    chrono::Utc::now().timestamp_millis() as f64 / 1000.0
}

impl EventJob {
    pub fn new(camera_id: i64, confianca: f64) -> Self {
        let mut camera = Map::new();
        camera.insert("id".into(), Value::Number(camera_id.into()));
        Self {
            camera,
            confianca,
            detected_at: default_detected_at(),
            snapshot_path: None,
        }
    }

    pub fn from_camera(camera: &Value, confianca: f64, snapshot_path: Option<String>) -> Self {
        let cam_map = if let Some(obj) = camera.as_object() {
            obj.clone()
        } else {
            let mut m = Map::new();
            m.insert("id".into(), camera.clone());
            m
        };
        Self {
            camera: cam_map,
            confianca,
            detected_at: default_detected_at(),
            snapshot_path,
        }
    }

    pub fn camera_id(&self) -> Option<i64> {
        self.camera.get("id").and_then(|v| v.as_i64())
    }
}

#[derive(Debug, Clone, Serialize)]
pub struct DlqEnvelope {
    pub reason: String,
    pub job: EventJob,
    pub at: f64,
}
