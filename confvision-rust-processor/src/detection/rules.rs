use super::areas::{find_area_for_box, normalize_modo_deteccao, AreaZone};
use crate::yolo::PersonDetection;
use serde_json::Value;

#[derive(Debug, Clone, Copy)]
pub struct MatchResult {
    pub best_conf: f64,
    pub pessoas_total: u32,
    pub pessoas_match: u32,
}

pub fn evaluate_detections<'a>(
    detections: &[PersonDetection],
    conf_min: f64,
    areas: &'a [AreaZone],
    modo: &str,
    frame_w: f64,
    frame_h: f64,
) -> (MatchResult, Option<&'a AreaZone>) {
    let mut best_conf = 0.0f64;
    let mut best_area: Option<&AreaZone> = None;
    let mut pessoas = 0u32;
    let mut pessoas_match = 0u32;

    for det in detections {
        if det.class_id != 0 || det.confidence < conf_min {
            continue;
        }
        pessoas += 1;
        let area = if areas.is_empty() {
            None
        } else {
            find_area_for_box(det.xyxy, frame_w, frame_h, areas)
        };

        let bate = match modo {
            "ambos" => true,
            "fora" => area.is_none(),
            _ => area.is_some(),
        };
        if !bate {
            continue;
        }
        pessoas_match += 1;
        if det.confidence > best_conf {
            best_conf = det.confidence;
            best_area = area;
        }
    }

    (
        MatchResult {
            best_conf,
            pessoas_total: pessoas,
            pessoas_match,
        },
        best_area,
    )
}

pub fn camera_conf_min(camera: &Value, default: f64) -> f64 {
    camera
        .get("confianca_minima")
        .or_else(|| camera.get("confianca_min"))
        .or_else(|| camera.get("conf_min"))
        .and_then(|v| v.as_f64())
        .unwrap_or(default)
        .clamp(0.01, 1.0)
}

pub fn camera_cooldown_sec(camera: &Value) -> u64 {
    camera
        .get("cooldown_seg")
        .and_then(|v| v.as_u64().or_else(|| v.as_i64().map(|x| x as u64)))
        .unwrap_or(30)
        .max(1)
}

pub fn camera_modo(camera: &Value) -> &'static str {
    normalize_modo_deteccao(camera.get("modo_deteccao"))
}

pub fn camera_deteccao_humano(camera: &Value) -> bool {
    camera
        .get("deteccao_humano")
        .and_then(|v| v.as_bool())
        .unwrap_or(true)
}
