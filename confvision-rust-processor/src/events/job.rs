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
}

#[derive(Debug, Clone, Serialize)]
pub struct DlqEnvelope {
    pub reason: String,
    pub job: EventJob,
    pub at: f64,
}
