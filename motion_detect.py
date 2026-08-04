"""Detecção barata de movimento (MOG2) — gate antes do YOLO analítico."""

from __future__ import annotations

import cv2

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


def frame_has_motion(frame, fgbg, kernel) -> bool:
    height, width = frame.shape[:2]
    target_w = min(MOTION_ANALYSIS_WIDTH, width)
    target_h = max(1, int(height * target_w / width))
    small = cv2.resize(frame, (target_w, target_h))
    fgmask = fgbg.apply(small)
    _, fgmask = cv2.threshold(fgmask, 200, 255, cv2.THRESH_BINARY)
    fgmask = cv2.morphologyEx(fgmask, cv2.MORPH_OPEN, kernel)
    contours, _ = cv2.findContours(
        fgmask, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE
    )
    return any(cv2.contourArea(c) > MOTION_MIN_AREA for c in contours)
