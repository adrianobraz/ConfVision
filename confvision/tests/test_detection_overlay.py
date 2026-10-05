"""Testes do overlay de bbox (sem YOLO)."""

from __future__ import annotations

from types import SimpleNamespace

import cv2
import numpy as np

from detection_overlay import (
    all_person_boxes,
    boxes_to_bbox_json,
    draw_corner_bracket_box,
    draw_person_boxes_on_frame,
    matching_person_boxes,
)


class _FakeXY:
    def __init__(self, coords):
        self._coords = coords

    def tolist(self):
        return list(self._coords)


class _FakeBox:
    def __init__(self, xyxy, cls_id: int, conf: float):
        self.xyxy = [_FakeXY(xyxy)]
        self.cls = [cls_id]
        self.conf = [conf]


class _FakeBoxes:
    def __init__(self, items):
        self._items = items

    def __iter__(self):
        return iter(self._items)


def test_matching_person_ambos():
    results = SimpleNamespace(
        orig_shape=(480, 640),
        boxes=_FakeBoxes([_FakeBox([10, 20, 100, 200], 0, 0.9)]),
    )
    boxes = matching_person_boxes(results, 0.5, [], "ambos")
    assert len(boxes) == 1
    assert boxes[0][:4] == (10, 20, 100, 200)


def test_all_person_boxes_ignores_area():
    results = SimpleNamespace(
        orig_shape=(480, 640),
        boxes=_FakeBoxes(
            [
                _FakeBox([10, 20, 100, 200], 0, 0.9),
                _FakeBox([200, 20, 300, 200], 1, 0.95),
            ]
        ),
    )
    boxes = all_person_boxes(results, 0.5)
    assert len(boxes) == 1


def test_boxes_to_bbox_json_pct():
    js = boxes_to_bbox_json([(0, 0, 640, 480, 0.88)], 640, 480)
    assert "x2_pct" in js
    assert "100.0" in js or "100" in js


def test_draw_returns_copy_with_marks():
    frame = np.zeros((100, 100, 3), dtype=np.uint8)
    before = int(frame.sum())
    out = draw_person_boxes_on_frame(frame, [(5, 5, 50, 50, 0.8)])
    assert int(frame.sum()) == before
    assert out is not frame
    assert int(out.sum()) > before


def test_corner_bracket_draws_on_frame():
    frame = np.zeros((120, 120, 3), dtype=np.uint8)
    draw_corner_bracket_box(frame, 10, 10, 90, 100)
    assert int(frame.sum()) > 0
