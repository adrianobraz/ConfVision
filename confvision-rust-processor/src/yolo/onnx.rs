//! YOLOv8 ONNX in-process (feature `yolo-onnx`, ONNX Runtime via `ort`).

use std::path::Path;
use std::sync::Mutex;

use image::ImageReader;
use ort::session::builder::GraphOptimizationLevel;
use ort::session::Session;
use ort::value::Tensor;
use tracing::debug;

use crate::decode::DecodedFrame;
use crate::error::{AppError, AppResult};

use super::onnx_postprocess::parse_yolov8_output;
use super::types::PersonDetection;

pub struct OnnxYolo {
    session: Mutex<Session>,
    input_name: String,
    input_size: u32,
    conf_min: f32,
    device: &'static str,
}

impl OnnxYolo {
    pub fn load(path: &str, input_size: u32, conf_min: f32) -> AppResult<Self> {
        if !Path::new(path).is_file() {
            return Err(AppError::Config(format!("YOLO_MODEL_PATH inexistente: {path}")));
        }

        let session = Session::builder()
            .map_err(|e| AppError::Config(format!("ort builder: {e}")))?
            .with_optimization_level(GraphOptimizationLevel::Level3)
            .map_err(|e| AppError::Config(format!("ort opt level: {e}")))?
            .with_intra_threads(num_cpus())
            .map_err(|e| AppError::Config(format!("ort threads: {e}")))?
            .commit_from_file(path)
            .map_err(|e| AppError::Config(format!("ort load {path}: {e}")))?;

        let inputs = session.inputs();
        let input_name = inputs
            .first()
            .map(|i| i.name().to_string())
            .unwrap_or_else(|| "images".to_string());

        Ok(Self {
            session: Mutex::new(session),
            input_name,
            input_size: input_size.max(320),
            conf_min: conf_min.clamp(0.01, 1.0) as f32,
            device: "onnx-cpu",
        })
    }

    pub fn device_label(&self) -> &'static str {
        self.device
    }

    pub fn infer_luma_sync(&self, decoded: &DecodedFrame) -> AppResult<Vec<PersonDetection>> {
        let (orig_w, orig_h) = (decoded.luma_width, decoded.luma_height);
        let input = preprocess_luma(&decoded.luma, orig_w, orig_h, self.input_size)?;
        self.run_tensor(&input, orig_w, orig_h)
    }

    pub fn infer_jpeg_sync(&self, jpeg: &[u8]) -> AppResult<Vec<PersonDetection>> {
        let img = ImageReader::new(std::io::Cursor::new(jpeg))
            .with_guessed_format()
            .map_err(|e| AppError::Other(e.into()))?
            .decode()
            .map_err(|e| AppError::Other(e.into()))?;
        let rgb = img.to_rgb8();
        let (orig_w, orig_h) = rgb.dimensions();
        let input = preprocess_rgb(&rgb, self.input_size);
        self.run_tensor(&input, orig_w, orig_h)
    }

    fn run_tensor(
        &self,
        input: &[f32],
        orig_w: u32,
        orig_h: u32,
    ) -> AppResult<Vec<PersonDetection>> {
        let size = self.input_size as usize;
        let shape = [1usize, 3, size, size];
        let tensor = Tensor::from_array((shape, input.to_vec()))
            .map_err(|e| AppError::Other(e.into()))?;

        let mut session = self
            .session
            .lock()
            .map_err(|_| AppError::Other(anyhow::anyhow!("ort session lock poisoned")))?;

        let outputs = session
            .run(ort::inputs![self.input_name.as_str() => tensor])
            .map_err(|e| AppError::Other(e.into()))?;

        let (_name, value) = outputs
            .into_iter()
            .next()
            .ok_or_else(|| AppError::Config("onnx sem outputs".into()))?;

        let (shape, data) = value
            .try_extract_tensor::<f32>()
            .map_err(|e| AppError::Other(e.into()))?;

        let detections = parse_output_tensor(&shape, &data, self.input_size, orig_w, orig_h, self.conf_min);
        debug!(count = detections.len(), "yolo onnx infer");
        Ok(detections)
    }
}

fn parse_output_tensor(
    shape: &[i64],
    data: &[f32],
    input_size: u32,
    orig_w: u32,
    orig_h: u32,
    conf_min: f32,
) -> Vec<PersonDetection> {
    if shape.len() == 3 {
        let channels = shape[1] as usize;
        let anchors = shape[2] as usize;
        return parse_yolov8_output(
            data,
            channels,
            anchors,
            conf_min,
            input_size,
            orig_w as f32,
            orig_h as f32,
        );
    }
    if shape.len() == 2 {
        let anchors = shape[0] as usize;
        let channels = shape[1] as usize;
        if channels >= 5 && data.len() >= anchors * channels {
            let mut transposed = vec![0f32; channels * anchors];
            for a in 0..anchors {
                for c in 0..channels {
                    transposed[c * anchors + a] = data[a * channels + c];
                }
            }
            return parse_yolov8_output(
                &transposed,
                channels,
                anchors,
                conf_min,
                input_size,
                orig_w as f32,
                orig_h as f32,
            );
        }
    }
    vec![]
}

fn preprocess_luma(luma: &[u8], orig_w: u32, orig_h: u32, input_size: u32) -> AppResult<Vec<f32>> {
    let expected = (orig_w as usize) * (orig_h as usize);
    if luma.len() != expected {
        return Err(AppError::Config(format!(
            "luma size mismatch: {} vs {expected}",
            luma.len()
        )));
    }
    let mut rgb = vec![0u8; expected * 3];
    for (i, &y) in luma.iter().enumerate() {
        let o = i * 3;
        rgb[o] = y;
        rgb[o + 1] = y;
        rgb[o + 2] = y;
    }
    Ok(letterbox_nchw(&rgb, orig_w, orig_h, input_size))
}

fn preprocess_rgb(rgb: &image::RgbImage, input_size: u32) -> Vec<f32> {
    let (w, h) = rgb.dimensions();
    letterbox_nchw(rgb.as_raw(), w, h, input_size)
}

fn letterbox_nchw(rgb: &[u8], orig_w: u32, orig_h: u32, input_size: u32) -> Vec<f32> {
    let input_size = input_size as f32;
    let orig_w = orig_w as f32;
    let orig_h = orig_h as f32;
    let scale = (input_size / orig_w).min(input_size / orig_h);
    let new_w = (orig_w * scale).round() as u32;
    let new_h = (orig_h * scale).round() as u32;
    let pad_x = ((input_size as u32).saturating_sub(new_w)) / 2;
    let pad_y = ((input_size as u32).saturating_sub(new_h)) / 2;

    let size = input_size as usize;
    let mut out = vec![0f32; 3 * size * size];
    let inv255 = 1.0 / 255.0;

    for dst_y in 0..new_h {
        for dst_x in 0..new_w {
            let src_x = ((dst_x as f32 / scale).floor() as u32).min(orig_w as u32 - 1);
            let src_y = ((dst_y as f32 / scale).floor() as u32).min(orig_h as u32 - 1);
            let src_idx = ((src_y * orig_w as u32 + src_x) * 3) as usize;
            if src_idx + 2 >= rgb.len() {
                continue;
            }
            let px = pad_x + dst_x;
            let py = pad_y + dst_y;
            if px >= size as u32 || py >= size as u32 {
                continue;
            }
            let plane_idx = py as usize * size + px as usize;
            out[plane_idx] = rgb[src_idx] as f32 * inv255;
            out[size * size + plane_idx] = rgb[src_idx + 1] as f32 * inv255;
            out[2 * size * size + plane_idx] = rgb[src_idx + 2] as f32 * inv255;
        }
    }
    out
}

fn num_cpus() -> usize {
    std::thread::available_parallelism()
        .map(|n| n.get().max(1))
        .unwrap_or(1)
}
