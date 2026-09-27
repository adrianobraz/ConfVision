//! YOLOv8 ONNX (feature `yolo-onnx`).

use crate::error::{AppError, AppResult};

use super::types::PersonDetection;

pub struct OnnxYolo {
    _path: String,
}

impl OnnxYolo {
    pub fn load(path: &str) -> AppResult<Self> {
        let _ = std::path::Path::new(path);
        // Extensão futura: ort Session + preprocess/postprocess YOLOv8.
        Err(AppError::Config(format!(
            "YOLO ONNX em {path}: implementação ort pendente — use YOLO_BACKEND=http ou aguarde build onnx"
        )))
    }

    pub async fn infer_jpeg(&self, _jpeg: &[u8]) -> AppResult<Vec<PersonDetection>> {
        Ok(vec![])
    }
}
