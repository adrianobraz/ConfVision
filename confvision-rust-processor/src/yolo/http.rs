use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::Arc;
use std::time::{Duration, SystemTime, UNIX_EPOCH};

use base64::{engine::general_purpose::STANDARD, Engine};
use reqwest::Client;
use reqwest::StatusCode;
use serde::Deserialize;
use tokio::sync::Semaphore;
use tracing::debug;

use crate::error::{AppError, AppResult};

use super::types::PersonDetection;

#[derive(Debug, Deserialize)]
struct SidecarHealth {
    busy: Option<bool>,
    active_requests: Option<u32>,
    max_queue: Option<u32>,
}

#[derive(Clone)]
struct SidecarGate {
    global: Arc<Semaphore>,
    busy_until_ms: Arc<AtomicU64>,
    backoff_ms: u64,
    health_probe: bool,
}

impl SidecarGate {
    fn mark_busy(&self) {
        let until = now_ms().saturating_add(self.backoff_ms);
        let _ = self
            .busy_until_ms
            .fetch_update(Ordering::Relaxed, Ordering::Relaxed, |prev| {
                Some(prev.max(until))
            });
    }

    fn in_backoff(&self) -> bool {
        now_ms() < self.busy_until_ms.load(Ordering::Relaxed)
    }
}

fn now_ms() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_millis() as u64)
        .unwrap_or(0)
}

#[derive(Clone)]
pub struct HttpYoloClient {
    http: Client,
    health_http: Client,
    url: String,
    gate: SidecarGate,
}

#[derive(Debug, Deserialize)]
struct DetectResponse {
    width: f64,
    height: f64,
    detections: Vec<DetItem>,
}

#[derive(Debug, Deserialize)]
struct DetItem {
    class_id: i32,
    confidence: f64,
    xyxy: [f64; 4],
}

impl HttpYoloClient {
    pub fn new(
        url: &str,
        timeout_sec: u64,
        global_max_inflight: usize,
        busy_backoff_ms: u64,
        health_probe: bool,
    ) -> AppResult<Self> {
        let timeout = Duration::from_secs(timeout_sec.max(5));
        let http = Client::builder()
            .timeout(timeout)
            .build()
            .map_err(|e| AppError::Config(e.to_string()))?;
        let health_http = Client::builder()
            .timeout(Duration::from_secs(3))
            .build()
            .map_err(|e| AppError::Config(e.to_string()))?;
        Ok(Self {
            http,
            health_http,
            url: url.trim_end_matches('/').to_string(),
            gate: SidecarGate {
                global: Arc::new(Semaphore::new(global_max_inflight.max(1))),
                busy_until_ms: Arc::new(AtomicU64::new(0)),
                backoff_ms: busy_backoff_ms.max(200),
                health_probe,
            },
        })
    }

    async fn sidecar_overloaded(&self) -> AppResult<bool> {
        if self.gate.in_backoff() {
            return Ok(true);
        }
        if !self.gate.health_probe {
            return Ok(false);
        }
        let resp = self
            .health_http
            .get(format!("{}/health", self.url))
            .send()
            .await
            .map_err(|e| AppError::Other(e.into()))?;
        if resp.status() == StatusCode::SERVICE_UNAVAILABLE {
            self.gate.mark_busy();
            return Ok(true);
        }
        if !resp.status().is_success() {
            return Ok(false);
        }
        let body: SidecarHealth = resp.json().await.map_err(|e| AppError::Other(e.into()))?;
        if body.busy == Some(true) {
            self.gate.mark_busy();
            return Ok(true);
        }
        if let (Some(act), Some(max)) = (body.active_requests, body.max_queue) {
            if max > 0 && act + 1 >= max {
                self.gate.mark_busy();
                return Ok(true);
            }
        }
        Ok(false)
    }

    pub async fn infer_jpeg(&self, jpeg: &[u8]) -> AppResult<Vec<PersonDetection>> {
        if self.gate.in_backoff() {
            return Err(AppError::YoloBusy("backoff local apos 503/fila cheia".into()));
        }
        if self.sidecar_overloaded().await? {
            return Err(AppError::YoloBusy(
                "sidecar health indica fila cheia ou busy".into(),
            ));
        }
        let permit = self
            .gate
            .global
            .clone()
            .try_acquire_owned()
            .map_err(|_| AppError::YoloBusy("limite global YOLO_GLOBAL_MAX_INFLIGHT".into()))?;

        let body = serde_json::json!({
            "jpeg_base64": STANDARD.encode(jpeg),
        });
        let resp = self
            .http
            .post(format!("{}/v1/detect", self.url))
            .json(&body)
            .send()
            .await
            .map_err(|e| AppError::Other(e.into()))?;

        if resp.status() == StatusCode::SERVICE_UNAVAILABLE {
            drop(permit);
            self.gate.mark_busy();
            let text = resp.text().await.unwrap_or_default();
            return Err(AppError::YoloBusy(format!("yolo http 503: {text}")));
        }
        if !resp.status().is_success() {
            drop(permit);
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            return Err(AppError::Config(format!("yolo http {status}: {text}")));
        }
        let parsed: DetectResponse = resp.json().await.map_err(|e| AppError::Other(e.into()))?;
        drop(permit);
        debug!(
            width = parsed.width,
            height = parsed.height,
            count = parsed.detections.len(),
            "yolo http infer"
        );
        Ok(parsed
            .detections
            .into_iter()
            .map(|d| PersonDetection {
                class_id: d.class_id,
                confidence: d.confidence,
                xyxy: d.xyxy,
            })
            .collect())
    }
}

pub type SharedHttpYolo = Arc<HttpYoloClient>;
