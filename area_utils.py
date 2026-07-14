import json
from typing import Any, Optional

import cv2
import numpy as np


def parse_poligono(poligono_json: Any) -> Optional[list[tuple[float, float]]]:
    if not poligono_json:
        return None
    try:
        data = json.loads(poligono_json) if isinstance(poligono_json, str) else poligono_json
    except (json.JSONDecodeError, TypeError):
        return None
    if not isinstance(data, dict):
        return None
    raw_points = data.get("pontos") or data.get("points") or []
    if not isinstance(raw_points, list) or len(raw_points) < 3:
        return None
    points: list[tuple[float, float]] = []
    for item in raw_points:
        if not isinstance(item, dict):
            continue
        if "x" not in item or "y" not in item:
            continue
        points.append((float(item["x"]), float(item["y"])))
    return points if len(points) >= 3 else None


def bbox_foot_pct(box, frame_w: int, frame_h: int) -> tuple[float, float]:
    x1, y1, x2, y2 = box.xyxy[0].tolist()
    cx = ((x1 + x2) / 2.0 / frame_w) * 100.0
    cy = (y2 / frame_h) * 100.0
    return cx, cy


def point_in_polygon_pct(x_pct: float, y_pct: float, polygon: list[tuple[float, float]]) -> bool:
    pts = np.array(polygon, dtype=np.float32)
    result = cv2.pointPolygonTest(pts, (float(x_pct), float(y_pct)), False)
    return result >= 0


def find_area_for_point(
    x_pct: float, y_pct: float, areas: list[dict]
) -> Optional[dict]:
    for area in areas:
        if not area.get("ativo", True):
            continue
        polygon = parse_poligono(area.get("poligono_json"))
        if not polygon:
            continue
        if point_in_polygon_pct(x_pct, y_pct, polygon):
            return area
    return None


def areas_ativas(areas: Optional[list[dict]]) -> list[dict]:
    if not areas:
        return []
    result = []
    for area in areas:
        if not area.get("ativo", True):
            continue
        if parse_poligono(area.get("poligono_json")):
            result.append(area)
    return result
