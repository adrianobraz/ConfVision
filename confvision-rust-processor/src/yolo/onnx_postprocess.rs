//! Pós-processamento YOLOv8 ONNX (layout Ultralytics `[1, 84, N]`).

use super::types::PersonDetection;

const PERSON_CLASS_ID: i32 = 0;

/// Converte caixas do espaço letterbox (quadrado `input_size`) para pixels da imagem origem.
pub fn map_box_to_orig(
    cx: f32,
    cy: f32,
    w: f32,
    h: f32,
    input_size: u32,
    orig_w: f32,
    orig_h: f32,
) -> [f64; 4] {
    let input_size = input_size as f32;
    let scale = (input_size / orig_w).min(input_size / orig_h);
    let pad_x = (input_size - orig_w * scale) * 0.5;
    let pad_y = (input_size - orig_h * scale) * 0.5;

    let x1 = (cx - w * 0.5 - pad_x) / scale;
    let y1 = (cy - h * 0.5 - pad_y) / scale;
    let x2 = (cx + w * 0.5 - pad_x) / scale;
    let y2 = (cy + h * 0.5 - pad_y) / scale;

    [
        x1.max(0.0) as f64,
        y1.max(0.0) as f64,
        x2.min(orig_w) as f64,
        y2.min(orig_h) as f64,
    ]
}

/// `data` em layout `[1, channels=84, anchors=n]` (canal major).
pub fn parse_yolov8_output(
    data: &[f32],
    channels: usize,
    anchors: usize,
    conf_min: f32,
    input_size: u32,
    orig_w: f32,
    orig_h: f32,
) -> Vec<PersonDetection> {
    if channels < 5 || anchors == 0 || data.len() < channels * anchors {
        return vec![];
    }

    let num_classes = channels - 4;
    let mut out = Vec::new();

    for j in 0..anchors {
        let cx = data[0 * anchors + j];
        let cy = data[1 * anchors + j];
        let w = data[2 * anchors + j];
        let h = data[3 * anchors + j];

        let mut best_class = 0usize;
        let mut best_score = 0f32;
        for c in 0..num_classes {
            let score = data[(4 + c) * anchors + j];
            if score > best_score {
                best_score = score;
                best_class = c;
            }
        }

        if best_class != PERSON_CLASS_ID as usize || best_score < conf_min {
            continue;
        }

        let xyxy = map_box_to_orig(cx, cy, w, h, input_size, orig_w, orig_h);
        out.push(PersonDetection {
            class_id: PERSON_CLASS_ID,
            confidence: best_score as f64,
            xyxy,
        });
    }

    out.sort_by(|a, b| {
        b.confidence
            .partial_cmp(&a.confidence)
            .unwrap_or(std::cmp::Ordering::Equal)
    });
    out.truncate(32);
    out
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn empty_when_too_few_channels() {
        assert!(parse_yolov8_output(&[0.0; 4], 4, 1, 0.5, 640, 160.0, 120.0).is_empty());
    }

    #[test]
    fn detects_person_channel() {
        let anchors = 2;
        let channels = 84;
        let mut data = vec![0f32; channels * anchors];
        // anchor 0: person with high score at center
        data[0 * anchors + 0] = 320.0;
        data[1 * anchors + 0] = 320.0;
        data[2 * anchors + 0] = 80.0;
        data[3 * anchors + 0] = 120.0;
        data[4 * anchors + 0] = 0.92;

        let dets = parse_yolov8_output(&data, channels, anchors, 0.5, 640, 160.0, 120.0);
        assert_eq!(dets.len(), 1);
        assert!(dets[0].confidence > 0.9);
    }
}
