"""Desenho de bounding boxes nas fotos de evento (snapshot YOLO)."""

from __future__ import annotations

from typing import Any, Iterable

import cv2
import numpy as np

from area_utils import find_area_for_box

BBOX_COLOR_BGR = (0, 220, 255)
BBOX_THICKNESS = 2
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
        x1, y1, x2, y2 = box.xyxy[0].tolist()
        out.append((int(x1), int(y1), int(x2), int(y2), conf))
    return out


def draw_person_boxes_on_frame(
    frame: np.ndarray,
    boxes: Iterable[tuple[int, int, int, int, float]],
) -> np.ndarray:
    """Retorna copia do frame com retangulos e labels."""
    annotated = frame.copy()
    for x1, y1, x2, y2, conf in boxes:
        cv2.rectangle(annotated, (x1, y1), (x2, y2), BBOX_COLOR_BGR, BBOX_THICKNESS)
        label = f"Pessoa {conf:.0%}"
        (tw, th), baseline = cv2.getTextSize(label, LABEL_FONT, LABEL_SCALE, LABEL_THICKNESS)
        ty1 = max(y1, th + baseline + 4)
        cv2.rectangle(
            annotated,
            (x1, ty1 - th - baseline - 4),
            (x1 + tw + 6, ty1),
            BBOX_COLOR_BGR,
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


def annotate_event_snapshot_frame(
    frame: np.ndarray,
    results: Any,
    conf_min: float,
    zonas: list[dict],
    modo: str,
) -> np.ndarray:
    boxes = matching_person_boxes(results, conf_min, zonas, modo)
    if not boxes:
        return frame.copy()
    return draw_person_boxes_on_frame(frame, boxes)
