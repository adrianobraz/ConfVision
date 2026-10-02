"""Testes do overlay de bbox (sem YOLO)."""

from __future__ import annotations

from types import SimpleNamespace

import cv2

from detection_overlay import draw_person_boxes_on_frame, matching_person_boxes


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


def test_draw_returns_copy_with_marks():
    import numpy as np

    frame = np.zeros((100, 100, 3), dtype=np.uint8)
    before = int(frame.sum())
    out = draw_person_boxes_on_frame(frame, [(5, 5, 50, 50, 0.8)])
    assert int(frame.sum()) == before
    assert out is not frame
    assert int(out.sum()) > before
