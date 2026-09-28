use serde_json::Value;

#[derive(Debug, Clone)]
pub struct AreaZone {
    pub id: Option<i64>,
    pub nome: Option<String>,
    pub ativo: bool,
    pub polygon_pct: Vec<(f64, f64)>,
}

pub fn areas_from_camera(camera: &Value) -> Vec<AreaZone> {
    let Some(arr) = camera.get("areas").and_then(|v| v.as_array()) else {
        return vec![];
    };
    arr.iter().filter_map(parse_area).collect()
}

fn parse_area(v: &Value) -> Option<AreaZone> {
    let ativo = v.get("ativo").and_then(|x| x.as_bool()).unwrap_or(true);
    let poly = parse_poligono(v.get("poligono_json"))?;
    Some(AreaZone {
        id: v.get("id").and_then(|x| x.as_i64()),
        nome: v.get("nome").and_then(|x| x.as_str()).map(String::from),
        ativo,
        polygon_pct: poly,
    })
}

fn parse_poligono(raw: Option<&Value>) -> Option<Vec<(f64, f64)>> {
    let v = raw?;
    let data = if v.is_string() {
        serde_json::from_str(v.as_str()?).ok()?
    } else {
        v.clone()
    };
    let obj = data.as_object()?;
    let points = obj
        .get("pontos")
        .or_else(|| obj.get("points"))
        .and_then(|p| p.as_array())?;
    let mut out = Vec::new();
    for item in points {
        let x = item.get("x")?.as_f64()?;
        let y = item.get("y")?.as_f64()?;
        out.push((x, y));
    }
    if out.len() >= 3 {
        Some(out)
    } else {
        None
    }
}

pub fn normalize_modo_deteccao(raw: Option<&Value>) -> &'static str {
    let modo = raw
        .and_then(|v| v.as_str())
        .unwrap_or("dentro")
        .trim()
        .to_ascii_lowercase();
    match modo.as_str() {
        "fora" | "outside" => "fora",
        "ambos" | "both" => "ambos",
        "inside" | "dentro" => "dentro",
        _ => "dentro",
    }
}

/// Ponto dentro do polígono (coordenadas 0–100%, ray casting).
pub fn point_in_polygon_pct(x_pct: f64, y_pct: f64, polygon: &[(f64, f64)]) -> bool {
    let mut inside = false;
    let n = polygon.len();
    let mut j = n - 1;
    for i in 0..n {
        let (xi, yi) = polygon[i];
        let (xj, yj) = polygon[j];
        if ((yi > y_pct) != (yj > y_pct))
            && (x_pct < (xj - xi) * (y_pct - yi) / (yj - yi + f64::EPSILON) + xi)
        {
            inside = !inside;
        }
        j = i;
    }
    inside
}

pub fn find_area_for_point(x_pct: f64, y_pct: f64, areas: &[AreaZone]) -> Option<&AreaZone> {
    for area in areas {
        if !area.ativo {
            continue;
        }
        if point_in_polygon_pct(x_pct, y_pct, &area.polygon_pct) {
            return Some(area);
        }
    }
    None
}

fn bbox_probe_points_pct(xyxy: [f64; 4], frame_w: f64, frame_h: f64) -> Vec<(f64, f64)> {
    let (x1, y1, x2, y2) = (xyxy[0], xyxy[1], xyxy[2], xyxy[3]);
    let cx = ((x1 + x2) / 2.0 / frame_w) * 100.0;
    let cy_mid = ((y1 + y2) / 2.0 / frame_h) * 100.0;
    let cy_chest = (y1 + (y2 - y1) * 0.35) / frame_h * 100.0;
    let cy_foot = (y2 / frame_h) * 100.0;
    vec![(cx, cy_mid), (cx, cy_chest), (cx, cy_foot)]
}

pub fn find_area_for_box(
    xyxy: [f64; 4],
    frame_w: f64,
    frame_h: f64,
    areas: &[AreaZone],
) -> Option<&AreaZone> {
    for (x, y) in bbox_probe_points_pct(xyxy, frame_w, frame_h) {
        if let Some(a) = find_area_for_point(x, y, areas) {
            return Some(a);
        }
    }
    None
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    #[test]
    fn square_half_plane() {
        let poly = vec![(0.0, 0.0), (100.0, 0.0), (100.0, 100.0), (0.0, 100.0)];
        assert!(point_in_polygon_pct(50.0, 50.0, &poly));
        assert!(!point_in_polygon_pct(150.0, 50.0, &poly));
    }

    #[test]
    fn parse_areas_from_camera() {
        let cam = json!({
            "areas": [{
                "id": 1,
                "ativo": true,
                "poligono_json": {"pontos": [{"x": 0, "y": 0}, {"x": 100, "y": 0}, {"x": 100, "y": 100}]}
            }]
        });
        let a = areas_from_camera(&cam);
        assert_eq!(a.len(), 1);
    }
}
