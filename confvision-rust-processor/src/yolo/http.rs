use std::sync::Arc;
use std::time::Duration;

use base64::{engine::general_purpose::STANDARD, Engine};
use reqwest::Client;
use serde::Deserialize;
use tracing::debug;

use crate::error::{AppError, AppResult};

use super::types::PersonDetection;

#[derive(Clone)]
pub struct HttpYoloClient {
    http: Client,
    url: String,
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
    pub fn new(url: &str) -> AppResult<Self> {
        let http = Client::builder()
            .timeout(Duration::from_secs(15))
            .build()
            .map_err(|e| AppError::Config(e.to_string()))?;
        Ok(Self {
            http,
            url: url.trim_end_matches('/').to_string(),
        })
    }

    pub async fn infer_jpeg(&self, jpeg: &[u8]) -> AppResult<Vec<PersonDetection>> {
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
        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            return Err(AppError::Config(format!("yolo http {status}: {text}")));
        }
        let parsed: DetectResponse = resp
            .json()
            .await
            .map_err(|e| AppError::Other(e.into()))?;
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
