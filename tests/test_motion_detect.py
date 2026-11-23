"""MOG2: resize, buffers e decisao de movimento. Sem RTSP."""

from __future__ import annotations

import threading

import cv2
import numpy as np

import detection_metrics
from config import MOTION_ANALYSIS_WIDTH, MOTION_MIN_AREA
from motion_detect import (
    create_motion_detector,
    frame_has_motion,
    mask_has_motion,
)


def setup_function():
    detection_metrics.stop_reporter()
    detection_metrics.configure(True, 30)
    detection_metrics.reset()


def teardown_function():
    detection_metrics.stop_reporter()
    detection_metrics.configure(False, 30)
    detection_metrics.reset()


def _legacy(frame, fgbg, kernel) -> bool:
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


def _pair(frame):
    fg_old, kernel_old = create_motion_detector()
    fg_new, kernel_new = create_motion_detector()
    buffers = {}
    old = _legacy(frame, fg_old, kernel_old)
    new = frame_has_motion(frame, fg_new, kernel_new, buffers=buffers)
    return old, new, buffers


def test_largura_de_analise_continua_640():
    assert MOTION_ANALYSIS_WIDTH == 640
    assert MOTION_MIN_AREA == 1500


def test_frame_640_nao_chama_resize(monkeypatch):
    calls = []
    real = cv2.resize

    def wrapped(src, dsize, dst=None, *args, **kwargs):
        calls.append(dsize)
        return real(src, dsize, dst=dst, *args, **kwargs)

    monkeypatch.setattr("motion_detect.cv2.resize", wrapped)
    frame = np.zeros((480, 640, 3), np.uint8)
    frame[40:120, 40:120] = 255
    before = frame.copy()
    fgbg, kernel = create_motion_detector()
    frame_has_motion(frame, fgbg, kernel, buffers={})
    assert calls == []
    assert np.array_equal(frame, before)


def test_frame_1280_preserva_proporcao(monkeypatch):
    calls = []
    real = cv2.resize

    def wrapped(src, dsize, dst=None, *args, **kwargs):
        calls.append((dsize, dst is not None))
        return real(src, dsize, dst=dst, *args, **kwargs)

    monkeypatch.setattr("motion_detect.cv2.resize", wrapped)
    frame = np.zeros((720, 1280, 3), np.uint8)
    before = frame.copy()
    fgbg, kernel = create_motion_detector()
    buffers = {}
    frame_has_motion(frame, fgbg, kernel, buffers=buffers)
    assert calls == [((640, 360), True)]
    assert buffers["small"].shape == (360, 640, 3)
    assert np.array_equal(frame, before)


def test_frame_1920_preserva_proporcao():
    frame = np.zeros((1080, 1920, 3), np.uint8)
    before = frame.copy()
    fgbg, kernel = create_motion_detector()
    buffers = {}
    frame_has_motion(frame, fgbg, kernel, buffers=buffers)
    small = buffers["small"]
    assert small.shape == (360, 640, 3)
    assert abs((small.shape[1] / small.shape[0]) - (1920 / 1080)) < 1e-6
    reference = cv2.resize(frame, (640, 360))
    assert np.array_equal(small, reference)
    assert np.array_equal(frame, before)


def test_mascara_vazia_nao_tem_movimento():
    assert mask_has_motion(np.zeros((48, 64), np.uint8)) is False


def test_mascara_com_menos_de_1500_pixels_nao_dispara():
    mask = np.zeros((80, 80), np.uint8)
    mask[:20, :20] = 255
    assert int(cv2.countNonZero(mask)) < MOTION_MIN_AREA
    assert mask_has_motion(mask) is False


def test_contorno_acima_de_1500_dispara():
    mask = np.zeros((120, 120), np.uint8)
    mask[:80, :80] = 255
    assert mask_has_motion(mask) is True


def test_varios_contornos_pequenos_nao_disparam():
    mask = np.zeros((400, 400), np.uint8)
    for index in range(20):
        origin = 10 + index * 18
        mask[origin : origin + 10, 10:20] = 255
    white = int(cv2.countNonZero(mask))
    assert white > MOTION_MIN_AREA
    contours, _ = cv2.findContours(mask, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
    assert contours
    assert all(cv2.contourArea(c) <= MOTION_MIN_AREA for c in contours)
    assert mask_has_motion(mask) is False


def test_furo_mostra_por_que_o_atalho_de_pixels_e_incorreto():
    mask = np.zeros((480, 640), np.uint8)
    cv2.rectangle(mask, (20, 20), (60, 60), 255, -1)
    cv2.rectangle(mask, (23, 23), (57, 57), 0, -1)
    kernel = cv2.getStructuringElement(cv2.MORPH_ELLIPSE, (3, 3))
    opened = cv2.morphologyEx(mask, cv2.MORPH_OPEN, kernel)
    white = int(cv2.countNonZero(opened))
    assert white < MOTION_MIN_AREA
    assert mask_has_motion(opened) is True


def test_decisao_nova_igual_a_anterior():
    frames = []
    quiet = np.zeros((720, 1280, 3), np.uint8)
    for _ in range(4):
        frames.append(quiet.copy())
    moving = quiet.copy()
    moving[200:520, 300:900] = 255
    frames.append(moving)
    vga = np.zeros((480, 640, 3), np.uint8)
    vga[100:180, 100:180] = 180
    frames.append(vga)
    full = np.zeros((1080, 1920, 3), np.uint8)
    full[400:700, 600:1200] = 255
    frames.append(full)

    fg_old, kernel_old = create_motion_detector()
    fg_new, kernel_new = create_motion_detector()
    buffers = {}
    for frame in frames:
        old = _legacy(frame, fg_old, kernel_old)
        new = frame_has_motion(frame, fg_new, kernel_new, buffers=buffers)
        assert old is new


def test_buffers_de_cameras_diferentes_nao_se_misturam():
    fg1, k1 = create_motion_detector()
    fg2, k2 = create_motion_detector()
    b1, b2 = {}, {}
    wide = np.zeros((720, 1280, 3), np.uint8)
    wide[10:200, 10:400] = 255
    vga = np.zeros((480, 640, 3), np.uint8)
    vga[10:80, 10:80] = 255
    frame_has_motion(wide, fg1, k1, buffers=b1)
    frame_has_motion(vga, fg2, k2, buffers=b2)
    assert b1["mask"] is not b2["mask"]
    assert b1["opened"] is not b2["opened"]
    assert "small" in b1
    assert "small" not in b2


def test_frame_original_permanece_para_o_snapshot():
    frame = np.zeros((1080, 1920, 3), np.uint8)
    frame[10:40, 10:40] = (0, 255, 0)
    snapshot = frame.copy()
    fgbg, kernel = create_motion_detector()
    frame_has_motion(frame, fgbg, kernel, buffers={})
    assert np.array_equal(frame, snapshot)
    assert frame is not snapshot


def test_threads_usam_buffers_proprios():
    errors = []
    masks = []
    guard = threading.Lock()

    def worker(seed: int):
        try:
            fgbg, kernel = create_motion_detector()
            buffers = {}
            frame = np.zeros((240, 320, 3), np.uint8)
            frame[seed : seed + 30, 10:80] = seed * 20
            for _ in range(8):
                frame_has_motion(frame, fgbg, kernel, buffers=buffers)
            with guard:
                masks.append(id(buffers["mask"]))
        except Exception as exc:
            errors.append(exc)

    threads = [threading.Thread(target=worker, args=(index,)) for index in range(1, 7)]
    for thread in threads:
        thread.start()
    for thread in threads:
        thread.join()
    assert errors == []
    assert len(masks) == len(set(masks)) == 6


def test_metricas_separam_os_tempos_e_preservam_mog2():
    frame = np.zeros((720, 1280, 3), np.uint8)
    fgbg, kernel = create_motion_detector()
    timings = {}
    frame_has_motion(frame, fgbg, kernel, buffers={}, timings=timings)
    detection_metrics.record_rtsp_read(5, 0.064)
    detection_metrics.record_mog2(5, timings["total_s"])
    detection_metrics.record_motion_stages(5, timings)
    line = [
        item
        for item in detection_metrics.collect_lines(30)
        if item.startswith("[METRICS] camera=5 ")
    ][0]
    assert "mog2_ms_medio=" in line
    assert "resize_ms_medio=" in line
    assert "mog2_apply_ms_medio=" in line
    assert "threshold_ms_medio=" in line
    assert "morphology_ms_medio=" in line
    assert "contours_ms_medio=" in line
    assert "motion_total_ms_medio=" in line
    assert "rtsp://" not in line
