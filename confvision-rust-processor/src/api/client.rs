use std::time::Duration;

use reqwest::header::{HeaderMap, HeaderValue, AUTHORIZATION};
use serde::{Deserialize, Serialize};
use serde_json::Value;
use tracing::{debug, warn};

use crate::config::Config;
use crate::error::{AppError, AppResult};

#[derive(Clone)]
pub struct ConfVisionClient {
    http: reqwest::Client,
    base: String,
    headers: HeaderMap,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct CameraRecord {
    pub id: i64,
    #[serde(default)]
    pub ativo: bool,
    #[serde(default)]
    pub nome: Option<String>,
    #[serde(default)]
    pub rtsp_url_sec: Option<String>,
    #[serde(default)]
    pub mediamtx_rtsp_base: Option<String>,
    #[serde(default)]
    pub analitico_pausado: Option<bool>,
    #[serde(default)]
    pub deteccao_humano: Option<bool>,
    #[serde(default)]
    pub worker_id: Option<String>,
    #[serde(flatten)]
    pub extra: Value,
}

#[derive(Debug, Deserialize)]
pub struct SyncCamerasResponse {
    #[serde(default)]
    pub unchanged: bool,
    #[serde(default)]
    pub cameras: Vec<CameraRecord>,
    #[serde(default)]
    pub config_version: Option<String>,
}

#[derive(Debug, Serialize)]
pub struct WorkerPingRequest {
    pub worker_id: String,
    pub worker_tipo: String,
    pub hostname: String,
    pub versao: String,
    pub cameras_ativas: i32,
    pub ultimo_ping_em: String,
    pub ativo: bool,
    pub yolo_device: String,
    pub queue_backend: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub vis_mediamtx_node_id: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_cameras: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub shard_index: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub shard_total: Option<i32>,
}

impl ConfVisionClient {
    pub fn new(cfg: &Config) -> AppResult<Self> {
        let http = reqwest::Client::builder()
            .timeout(Duration::from_secs(45))
            .build()
            .map_err(|e| AppError::Api(e.to_string()))?;

        let mut headers = HeaderMap::new();
        if !cfg.vis_worker_api_key.is_empty() {
            headers.insert(
                "X-Vis-Worker-Key",
                HeaderValue::from_str(&cfg.vis_worker_api_key)
                    .map_err(|e| AppError::Api(e.to_string()))?,
            );
        }

        Ok(Self {
            http,
            base: cfg.confvision_api_url.clone(),
            headers,
        })
    }

    pub async fn sync_cameras_ativas(
        &self,
        worker_id: Option<&str>,
        node_id: u32,
        since_version: Option<&str>,
    ) -> AppResult<SyncCamerasResponse> {
        let mut url = format!("{}/vis_camera_sync_ativas", self.base);
        let mut qs = vec![];
        if let Some(w) = worker_id {
            qs.push(format!("worker_id={}", urlencoding(w)));
        }
        if node_id > 0 {
            qs.push(format!("vis_mediamtx_node_id={node_id}"));
        }
        if let Some(v) = since_version {
            qs.push(format!("since_version={}", urlencoding(v)));
        }
        if !qs.is_empty() {
            url.push('?');
            url.push_str(&qs.join("&"));
        }

        debug!(url = %url, "sync cameras");
        let resp = self
            .http
            .get(&url)
            .headers(self.headers.clone())
            .send()
            .await
            .map_err(|e| AppError::Api(e.to_string()))?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            return Err(AppError::Api(format!("sync HTTP {status}: {body}")));
        }

        let raw: Value = resp
            .json()
            .await
            .map_err(|e| AppError::Api(e.to_string()))?;

        if raw.get("unchanged").and_then(|v| v.as_bool()) == Some(true) {
            return Ok(SyncCamerasResponse {
                unchanged: true,
                cameras: vec![],
                config_version: raw
                    .get("config_version")
                    .and_then(|v| v.as_str())
                    .map(String::from),
            });
        }

        let cameras = parse_cameras(&raw);
        Ok(SyncCamerasResponse {
            unchanged: false,
            cameras,
            config_version: raw
                .get("config_version")
                .and_then(|v| v.as_str())
                .map(String::from),
        })
    }

    pub async fn worker_ping(&self, body: &WorkerPingRequest) -> AppResult<()> {
        let url = format!("{}/vis_worker_ping", self.base);
        let resp = self
            .http
            .post(&url)
            .headers(self.headers.clone())
            .json(body)
            .send()
            .await
            .map_err(|e| AppError::Api(e.to_string()))?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            warn!(status = %status, body = %text, "worker ping failed");
            return Err(AppError::Api(format!("ping HTTP {status}")));
        }
        Ok(())
    }
}

fn parse_cameras(raw: &Value) -> Vec<CameraRecord> {
    let arr = raw
        .get("cameras")
        .or_else(|| raw.get("dados"))
        .and_then(|v| v.as_array());

    let Some(arr) = arr else {
        return vec![];
    };

    arr.iter()
        .filter_map(|item| {
            serde_json::from_value::<CameraRecord>(item.clone())
                .map_err(|e| {
                    warn!(error = %e, "camera parse skip");
                    e
                })
                .ok()
        })
        .collect()
}

fn urlencoding(s: &str) -> String {
    url::form_urlencoded::byte_serialize(s.as_bytes()).collect()
}

#[cfg(test)]
mod tests {
    use super::*;
    use wiremock::matchers::{method, path};
    use wiremock::{Mock, MockServer, ResponseTemplate};

    #[tokio::test]
    async fn sync_parses_cameras() {
        let server = MockServer::start().await;
        Mock::given(method("GET"))
            .and(path("/vis_camera_sync_ativas"))
            .respond_with(ResponseTemplate::new(200).set_body_json(serde_json::json!({
                "cameras": [{"id": 42, "ativo": true, "deteccao_humano": true}],
                "config_version": "1c-0a-42"
            })))
            .mount(&server)
            .await;

        let cfg = test_config(&server.uri());
        let client = ConfVisionClient::new(&cfg).unwrap();
        let resp = client.sync_cameras_ativas(None, 0, None).await.unwrap();
        assert_eq!(resp.cameras.len(), 1);
        assert_eq!(resp.cameras[0].id, 42);
    }

    fn test_config(base: &str) -> Config {
        Config {
            confvision_api_url: base.to_string(),
            vis_worker_api_key: "test-key".into(),
            mediamtx_rtsp_base: "rtsp://localhost:8554".into(),
            rtmp_publish_secret: None,
            processor_id: "test".into(),
            processor_hostname: "host".into(),
            processor_version: "0.1.0".into(),
            worker_tipo: "rust_processor".into(),
            worker_id: "worker-01".into(),
            shard_mode: crate::config::ShardMode::Auto,
            worker_shard_index: -1,
            worker_shard_total: 0,
            mediamtx_node_id: 0,
            redis_url: None,
            s3_endpoint: None,
            s3_bucket: None,
            http_host: "127.0.0.1".into(),
            http_port: 8090,
            log_level: "info".into(),
            max_cameras: 5,
            sync_interval: Duration::from_secs(60),
            ping_interval: Duration::from_secs(30),
            rtsp_connect_timeout: Duration::from_secs(5),
            rtsp_reconnect_base: Duration::from_secs(10),
            rtsp_frame_timeout: Duration::from_secs(30),
            frame_buffer_max: 2,
            queue_backend: "none".into(),
            capacity_mode: crate::config::CapacityMode::Dynamic,
            capacity_cpu_target_percent: 80.0,
            capacity_memory_target_percent: 80.0,
            capacity_gpu_target_percent: 80.0,
            capacity_vram_target_percent: 80.0,
            capacity_min_sample_sec: 30,
            capacity_safety_factor: 0.80,
            capacity_history_size: 120,
            capacity_sample_interval_sec: 5,
            decode_runtime_fallback: true,
            decode_hw_error_threshold: 10,
            load_policy_mode: crate::load::LoadPolicyMode::Advisory,
            load_admission_enabled: false,
            motion_analysis_max_fps: 0.0,
            motion_frame_stride: 1,
            decode_frame_stride: 1,
            analysis_only_on_motion: false,
            motion_gate_probe_max_fps: 0.5,
            motion_gate_miss_frames: 10,
        }
    }
}
