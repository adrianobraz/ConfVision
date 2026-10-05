"""Desenho de bounding boxes nas fotos de evento (snapshot YOLO)."""

from __future__ import annotations

import json
from typing import Any, Iterable

import cv2
import numpy as np

from area_utils import find_area_for_box
from config import DETECTION_DRAW_ALL, DETECTION_OVERLAY_LABELS

# Estilo Intelbras-like: cantos amarelos, laterais ciano (BGR)
BBOX_CORNER_BGR = (0, 255, 255)
BBOX_EDGE_BGR = (255, 220, 0)
BBOX_THICKNESS = 2
CORNER_FRAC = 0.22
CORNER_MIN_PX = 10
CORNER_MAX_PX = 48
LABEL_FONT = cv2.FONT_HERSHEY_SIMPLEX
LABEL_SCALE = 0.55
LABEL_THICKNESS = 1


def _person_box_matches(
    box,
    conf_min: float,
    frame_w: int,
    frame_h: int,
    zonas: list[dict],
    modo: str,
) -> tuple[bool, float]:
    if int(box.cls[0]) != 0:
        return False, 0.0
    conf = float(box.conf[0])
    if conf < conf_min:
        return False, conf
    area = find_area_for_box(box, frame_w, frame_h, zonas) if zonas else None
    if modo == "ambos":
        bate = True
    elif modo == "fora":
        bate = area is None
    else:
        bate = area is not None
    return bate, conf


def _box_xyxy_int(box) -> tuple[int, int, int, int, float]:
    x1, y1, x2, y2 = box.xyxy[0].tolist()
    return int(x1), int(y1), int(x2), int(y2), float(box.conf[0])


def all_person_boxes(
    results,
    conf_min: float,
) -> list[tuple[int, int, int, int, float]]:
    """Todas as pessoas (classe 0) acima do limiar — util para debug/calibracao."""
    out: list[tuple[int, int, int, int, float]] = []
    if results.boxes is None:
        return out
    for box in results.boxes:
        if int(box.cls[0]) != 0:
            continue
        conf = float(box.conf[0])
        if conf < conf_min:
            continue
        out.append(_box_xyxy_int(box))
    return out


def matching_person_boxes(
    results,
    conf_min: float,
    zonas: list[dict],
    modo: str,
) -> list[tuple[int, int, int, int, float]]:
    """Caixas (x1,y1,x2,y2,conf) das pessoas que gerariam evento."""
    h, w = results.orig_shape
    out: list[tuple[int, int, int, int, float]] = []
    if results.boxes is None:
        return out
    for box in results.boxes:
        bate, conf = _person_box_matches(box, conf_min, w, h, zonas, modo)
        if not bate:
            continue
        x1, y1, x2, y2, _ = _box_xyxy_int(box)
        out.append((x1, y1, x2, y2, conf))
    return out


def boxes_to_bbox_json(
    boxes: Iterable[tuple[int, int, int, int, float]],
    frame_w: int,
    frame_h: int,
) -> str:
    """JSON normalizado (0-100 pct) para vis_evento.bbox_json."""
    items = []
    for x1, y1, x2, y2, conf in boxes:
        if frame_w <= 0 or frame_h <= 0:
            continue
        items.append(
            {
                "cls": "pessoa",
                "conf": round(conf, 4),
                "x1_pct": round(100.0 * x1 / frame_w, 2),
                "y1_pct": round(100.0 * y1 / frame_h, 2),
                "x2_pct": round(100.0 * x2 / frame_w, 2),
                "y2_pct": round(100.0 * y2 / frame_h, 2),
            }
        )
    return json.dumps(items, ensure_ascii=False)


def _corner_len(x1: int, y1: int, x2: int, y2: int) -> int:
    bw = max(1, x2 - x1)
    bh = max(1, y2 - y1)
    cl = int(min(bw, bh) * CORNER_FRAC)
    return max(CORNER_MIN_PX, min(CORNER_MAX_PX, cl))


def draw_corner_bracket_box(
    frame: np.ndarray,
    x1: int,
    y1: int,
    x2: int,
    y2: int,
    thickness: int = BBOX_THICKNESS,
) -> None:
    """Cantos amarelos + trechos centrais ciano (estilo DVR)."""
    x1, y1 = max(0, x1), max(0, y1)
    x2, y2 = max(x1 + 1, x2), max(y1 + 1, y2)
    cl = _corner_len(x1, y1, x2, y2)
    t = thickness

    def seg(p1, p2, color):
        cv2.line(frame, p1, p2, color, t, cv2.LINE_AA)

    # Cantos (amarelo)
    seg((x1, y1), (x1 + cl, y1), BBOX_CORNER_BGR)
    seg((x1, y1), (x1, y1 + cl), BBOX_CORNER_BGR)
    seg((x2 - cl, y1), (x2, y1), BBOX_CORNER_BGR)
    seg((x2, y1), (x2, y1 + cl), BBOX_CORNER_BGR)
    seg((x1, y2), (x1 + cl, y2), BBOX_CORNER_BGR)
    seg((x1, y2 - cl), (x1, y2), BBOX_CORNER_BGR)
    seg((x2 - cl, y2), (x2, y2), BBOX_CORNER_BGR)
    seg((x2, y2 - cl), (x2, y2), BBOX_CORNER_BGR)

    if x2 - x1 > 2 * cl + 4:
        seg((x1 + cl, y1), (x2 - cl, y1), BBOX_EDGE_BGR)
        seg((x1 + cl, y2), (x2 - cl, y2), BBOX_EDGE_BGR)
    if y2 - y1 > 2 * cl + 4:
        seg((x1, y1 + cl), (x1, y2 - cl), BBOX_EDGE_BGR)
        seg((x2, y1 + cl), (x2, y2 - cl), BBOX_EDGE_BGR)


def draw_person_boxes_on_frame(
    frame: np.ndarray,
    boxes: Iterable[tuple[int, int, int, int, float]],
) -> np.ndarray:
    """Retorna copia do frame com bordas estilo cantos + label opcional."""
    annotated = frame.copy()
    for x1, y1, x2, y2, conf in boxes:
        draw_corner_bracket_box(annotated, x1, y1, x2, y2)
        if not DETECTION_OVERLAY_LABELS:
            continue
        label = f"Pessoa {conf:.0%}"
        (tw, th), baseline = cv2.getTextSize(label, LABEL_FONT, LABEL_SCALE, LABEL_THICKNESS)
        ty1 = max(y1, th + baseline + 4)
        cv2.rectangle(
            annotated,
            (x1, ty1 - th - baseline - 4),
            (x1 + tw + 6, ty1),
            BBOX_CORNER_BGR,
            -1,
        )
        cv2.putText(
            annotated,
            label,
            (x1 + 3, ty1 - baseline - 2),
            LABEL_FONT,
            LABEL_SCALE,
            (0, 0, 0),
            LABEL_THICKNESS,
            cv2.LINE_AA,
        )
    return annotated


def select_boxes_for_overlay(
    results: Any,
    conf_min: float,
    zonas: list[dict],
    modo: str,
) -> list[tuple[int, int, int, int, float]]:
    if DETECTION_DRAW_ALL:
        return all_person_boxes(results, conf_min)
    return matching_person_boxes(results, conf_min, zonas, modo)


def annotate_event_snapshot_frame(
    frame: np.ndarray,
    results: Any,
    conf_min: float,
    zonas: list[dict],
    modo: str,
) -> np.ndarray:
    boxes = select_boxes_for_overlay(results, conf_min, zonas, modo)
    if not boxes:
        return frame.copy()
    return draw_person_boxes_on_frame(frame, boxes)


def overlay_boxes_and_bbox_json(
    frame: np.ndarray,
    results: Any,
    conf_min: float,
    zonas: list[dict],
    modo: str,
) -> tuple[np.ndarray, str]:
    """Frame anotado + bbox_json para persistencia no evento."""
    h, w = results.orig_shape
    boxes = select_boxes_for_overlay(results, conf_min, zonas, modo)
    if not boxes:
        return frame.copy(), "[]"
    return draw_person_boxes_on_frame(frame, boxes), boxes_to_bbox_json(boxes, w, h)
