"""Detecção barata de movimento (MOG2) — gate antes do YOLO analítico."""

from __future__ import annotations

import time

import cv2
import numpy as np

from config import (
    MOTION_ANALYSIS_WIDTH,
    MOTION_MIN_AREA,
    MOTION_MOG2_HISTORY,
    MOTION_MOG2_THRESHOLD,
)


def create_motion_detector():
    fgbg = cv2.createBackgroundSubtractorMOG2(
        history=MOTION_MOG2_HISTORY,
        varThreshold=MOTION_MOG2_THRESHOLD,
        detectShadows=False,
    )
    kernel = cv2.getStructuringElement(cv2.MORPH_ELLIPSE, (3, 3))
    return fgbg, kernel


def mask_has_motion(opened) -> bool:
    """True se algum contorno externo tem area maior que MOTION_MIN_AREA.

    Nao usa a soma dos pixels brancos. O contorno externo inclui buracos,
    entao a area pode passar de 1500 com menos de 1500 pixels brancos.
    """
    contours, _ = cv2.findContours(
        opened, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE
    )
    return any(cv2.contourArea(c) > MOTION_MIN_AREA for c in contours)


def _analysis_size(width: int, height: int) -> tuple[int, int]:
    target_w = min(MOTION_ANALYSIS_WIDTH, width)
    target_h = max(1, int(height * target_w / width))
    return target_w, target_h


def _buffer(buffers, key: str, shape, dtype):
    current = None if buffers is None else buffers.get(key)
    if current is None or current.shape != shape or current.dtype != dtype:
        current = np.empty(shape, dtype=dtype)
        if buffers is not None:
            buffers[key] = current
    return current


def _prepare_small(frame, buffers):
    height, width = frame.shape[:2]
    target_w, target_h = _analysis_size(width, height)
    if (
        width == MOTION_ANALYSIS_WIDTH
        and target_w == width
        and target_h == height
    ):
        return frame
    channels = frame.shape[2] if frame.ndim == 3 else 1
    if channels == 1:
        shape = (target_h, target_w)
    else:
        shape = (target_h, target_w, channels)
    small = _buffer(buffers, "small", shape, frame.dtype)
    cv2.resize(frame, (target_w, target_h), dst=small)
    return small


def frame_has_motion(frame, fgbg, kernel, buffers=None, timings=None) -> bool:
    measure = timings is not None
    started = time.perf_counter() if measure else 0.0

    mark = time.perf_counter() if measure else 0.0
    small = _prepare_small(frame, buffers)
    resize_s = (time.perf_counter() - mark) if measure else 0.0

    mark = time.perf_counter() if measure else 0.0
    mask_shape = (small.shape[0], small.shape[1])
    if buffers is None:
        fgmask = fgbg.apply(small)
    else:
        fgmask = _buffer(buffers, "mask", mask_shape, np.uint8)
        fgbg.apply(small, fgmask)
    apply_s = (time.perf_counter() - mark) if measure else 0.0

    mark = time.perf_counter() if measure else 0.0
    _, fgmask = cv2.threshold(fgmask, 200, 255, cv2.THRESH_BINARY, dst=fgmask)
    threshold_s = (time.perf_counter() - mark) if measure else 0.0

    mark = time.perf_counter() if measure else 0.0
    if buffers is None:
        opened = cv2.morphologyEx(fgmask, cv2.MORPH_OPEN, kernel)
    else:
        opened = _buffer(buffers, "opened", fgmask.shape, fgmask.dtype)
        cv2.morphologyEx(fgmask, cv2.MORPH_OPEN, kernel, dst=opened)
    morph_s = (time.perf_counter() - mark) if measure else 0.0

    mark = time.perf_counter() if measure else 0.0
    moved = mask_has_motion(opened)
    contour_s = (time.perf_counter() - mark) if measure else 0.0

    if measure:
        timings["resize_s"] = resize_s
        timings["apply_s"] = apply_s
        timings["threshold_s"] = threshold_s
        timings["morph_s"] = morph_s
        timings["contour_s"] = contour_s
        timings["total_s"] = time.perf_counter() - started
    return moved
