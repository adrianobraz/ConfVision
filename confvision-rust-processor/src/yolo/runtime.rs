use std::sync::Arc;

use tracing::info;

use crate::config::{Config, YoloBackend};
use crate::decode::DecodedFrame;
use crate::error::{AppError, AppResult};

use super::http::HttpYoloClient;
use super::types::PersonDetection;

#[derive(Clone)]
pub enum YoloRuntime {
    Off,
    Http(Arc<HttpYoloClient>),
    #[cfg(feature = "yolo-onnx")]
    Onnx(Arc<super::onnx::OnnxYolo>),
}

impl YoloRuntime {
    pub fn bootstrap(cfg: &Config) -> AppResult<Self> {
        if !cfg.yolo_enabled {
            info!(backend = "off", "YOLO desligado (YOLO_ENABLED=0)");
            return Ok(Self::Off);
        }
        let backend = if cfg.yolo_backend == YoloBackend::Off && cfg.yolo_http_url.is_some() {
            YoloBackend::Http
        } else {
            cfg.yolo_backend
        };
        match backend {
            YoloBackend::Off => {
                info!("YOLO_ENABLED=1 mas YOLO_BACKEND=off — inferência desligada");
                Ok(Self::Off)
            }
            YoloBackend::Http => {
                let url = cfg.yolo_http_url.as_ref().ok_or_else(|| {
                    AppError::Config("YOLO_BACKEND=http exige YOLO_HTTP_URL".into())
                })?;
                let client = HttpYoloClient::new(url, cfg.yolo_http_timeout_sec)?;
                info!(url = %url, "YOLO HTTP sidecar configurado");
                Ok(Self::Http(Arc::new(client)))
            }
            #[cfg(feature = "yolo-onnx")]
            YoloBackend::Onnx => {
                let path = cfg.yolo_model_path.as_ref().ok_or_else(|| {
                    AppError::Config("YOLO_BACKEND=onnx exige YOLO_MODEL_PATH".into())
                })?;
                let engine = super::onnx::OnnxYolo::load(
                    path,
                    cfg.yolo_onnx_input_size,
                    cfg.yolo_conf_default as f32,
                )?;
                info!(path = %path, input = cfg.yolo_onnx_input_size, "YOLO ONNX carregado");
                Ok(Self::Onnx(Arc::new(engine)))
            }
            #[cfg(not(feature = "yolo-onnx"))]
            YoloBackend::Onnx => Err(AppError::Config(
                "YOLO_BACKEND=onnx requer build --features yolo-onnx".into(),
            )),
        }
    }

    pub fn device_label(&self) -> &'static str {
        match self {
            Self::Off => "none",
            Self::Http(_) => "http",
            #[cfg(feature = "yolo-onnx")]
            Self::Onnx(e) => e.device_label(),
        }
    }

    pub fn is_active(&self) -> bool {
        !matches!(self, Self::Off)
    }

    pub fn uses_jpeg_for_infer(&self) -> bool {
        matches!(self, Self::Http(_))
    }

    /// Inferência preferencial: ONNX usa luma direto (sem JPEG/Base64); HTTP usa JPEG.
    pub async fn infer_frame(&self, decoded: &DecodedFrame, jpeg: &[u8]) -> AppResult<Vec<PersonDetection>> {
        match self {
            Self::Off => Ok(vec![]),
            Self::Http(c) => {
                if jpeg.is_empty() {
                    return Ok(vec![]);
                }
                c.infer_jpeg(jpeg).await
            }
            #[cfg(feature = "yolo-onnx")]
            Self::Onnx(e) => {
                let engine = Arc::clone(e);
                let frame = decoded.clone();
                tokio::task::spawn_blocking(move || engine.infer_luma_sync(&frame))
                    .await
                    .map_err(|e| AppError::Other(e.into()))?
            }
        }
    }

    pub async fn infer_jpeg(&self, jpeg: &[u8]) -> AppResult<Vec<PersonDetection>> {
        match self {
            Self::Off => Ok(vec![]),
            Self::Http(c) => c.infer_jpeg(jpeg).await,
            #[cfg(feature = "yolo-onnx")]
            Self::Onnx(e) => {
                let engine = Arc::clone(e);
                let data = jpeg.to_vec();
                tokio::task::spawn_blocking(move || engine.infer_jpeg_sync(&data))
                    .await
                    .map_err(|e| AppError::Other(e.into()))?
            }
        }
    }
}
