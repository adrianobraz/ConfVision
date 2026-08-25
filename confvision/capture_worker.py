"""Worker RTSP + MOG2 — grava evidencias e latest no FrameStore, enfileira YOLO."""

from __future__ import annotations

import os
import time
from typing import Callable, Optional

import cv2

from area_utils import areas_ativas, normalize_modo_deteccao
from config import (
    YOLO_EVIDENCE_FRAMES,
    YOLO_MOTION_FRAME_SKIP,
    YOLO_MOTION_MISS_FRAMES,
    YOLO_ONLY_ON_MOTION,
    YOLO_P2_FPS,
    YOLO_P3_FPS,
    YOLO_P4_FPS,
)
from frame_store import FrameStore
from motion_detect import create_motion_detector, frame_has_motion
from scheduler_queue import YoloJob, get_scheduler_queue
from urls import rtsp_url_for_camera

os.environ.setdefault("OPENCV_FFMPEG_CAPTURE_OPTIONS", "rtsp_transport;tcp")


def _open_capture(rtsp_url: str):
    if rtsp_url.startswith("rtmp://"):
        raise ValueError(f"Use RTSP para leitura, nao RTMP: {rtsp_url}")
    cap = cv2.VideoCapture(rtsp_url, cv2.CAP_FFMPEG)
    cap.set(cv2.CAP_PROP_BUFFERSIZE, 1)
    return cap


def _reset_gate() -> dict:
    return {
        "yolo_ligado": False,
        "motion_pending_verify": False,
        "miss_streak": 0,
        "evidence_idx": 0,
        "session_id": 0,
        "last_yolo_enqueue": 0.0,
        "priority": 4,
    }


def _fps_interval(priority: int) -> float:
    if priority == 2:
        return 1.0 / max(0.1, YOLO_P2_FPS)
    if priority == 3:
        return 1.0 / max(0.1, YOLO_P3_FPS)
    return 1.0 / max(0.1, YOLO_P4_FPS)


def _enqueue_job(
    store: FrameStore,
    queue,
    camera_id: int,
    slot: str,
    priority: int,
    session_id: int,
    motion_at: float,
) -> None:
    path = store.frame_path(camera_id, slot)
    job = YoloJob.create(
        camera_id=camera_id,
        slot=slot,
        priority=priority,
        session_id=session_id,
        motion_at=motion_at,
        frame_path=path,
    )
    queue.enqueue(job)


def run_capture_camera(
    camera: dict,
    should_continue: Callable[[], bool],
    frame_store: FrameStore | None = None,
    scheduler_queue=None,
) -> None:
    camera_id = int(camera.get("id") or 0)
    modo = normalize_modo_deteccao(camera.get("modo_deteccao"))
    zonas = areas_ativas(camera.get("areas"))
    url = rtsp_url_for_camera(camera)

    if modo != "ambos" and not zonas:
        print(f"[CAPTURE] camera id={camera_id} sem area — encerrada")
        return

    store = frame_store or FrameStore()
    queue = scheduler_queue or get_scheduler_queue()
    gate = _reset_gate()
    motion_fgbg = None
    motion_kernel = None
    if YOLO_ONLY_ON_MOTION:
        motion_fgbg, motion_kernel = create_motion_detector()

    cap = _open_capture(url)
    if not cap.isOpened():
        print(f"[CAPTURE] nao abriu RTSP camera={camera_id} url={url}")
        time.sleep(10)
        return

    print(f"[CAPTURE] camera={camera_id} url={url} modo={modo} areas={len(zonas)}")
    frame_index = 0
    falhas = 0

    while should_continue():
        ok, frame = cap.read()
        if not should_continue():
            break
        if not ok:
            falhas += 1
            cap.release()
            if not should_continue():
                break
            time.sleep(min(30, 2 * falhas))
            cap = _open_capture(url)
            if not cap.isOpened():
                return
            if YOLO_ONLY_ON_MOTION:
                motion_fgbg, motion_kernel = create_motion_detector()
                gate = _reset_gate()
            continue

        falhas = 0
        frame_index += 1
        agora = time.time()

        if YOLO_ONLY_ON_MOTION:
            if frame_index % max(1, YOLO_MOTION_FRAME_SKIP) == 0:
                if frame_has_motion(frame, motion_fgbg, motion_kernel):
                    if not gate["yolo_ligado"]:
                        gate["yolo_ligado"] = True
                        gate["motion_pending_verify"] = True
                        gate["miss_streak"] = 0
                        gate["evidence_idx"] = 0
                        meta = store.start_motion_session(camera_id, frame, motion_at=agora)
                        gate["session_id"] = meta.session_id
                        _enqueue_job(
                            store, queue, camera_id, "ev_0", 1,
                            meta.session_id, meta.motion_at,
                        )
                        gate["evidence_idx"] = 1
                        gate["last_yolo_enqueue"] = agora
                        continue
            if not gate["yolo_ligado"]:
                continue

        meta = store.read_meta(camera_id)
        if store.is_person_confirmed(camera_id, gate["session_id"]):
            gate["priority"] = 2
        elif meta.state == "evidence_pending":
            gate["priority"] = 1
        elif gate["yolo_ligado"]:
            gate["priority"] = 3
        else:
            gate["priority"] = 4

        if (
            gate["yolo_ligado"]
            and gate["evidence_idx"] < YOLO_EVIDENCE_FRAMES
            and not store.is_person_confirmed(camera_id, gate["session_id"])
        ):
            slot = f"ev_{gate['evidence_idx']}"
            store.append_evidence(camera_id, gate["evidence_idx"], frame)
            _enqueue_job(
                store, queue, camera_id, slot, 1,
                gate["session_id"], meta.motion_at or agora,
            )
            gate["evidence_idx"] += 1
            gate["last_yolo_enqueue"] = agora

        if gate["yolo_ligado"]:
            interval = _fps_interval(gate["priority"])
            if agora - gate["last_yolo_enqueue"] >= interval:
                store.put_latest(camera_id, frame, priority=gate["priority"])
                pri = gate["priority"] if gate["priority"] >= 2 else 3
                _enqueue_job(
                    store, queue, camera_id, "latest", pri,
                    gate["session_id"], meta.motion_at or agora,
                )
                gate["last_yolo_enqueue"] = agora

        if YOLO_ONLY_ON_MOTION and gate["yolo_ligado"]:
            if store.is_person_confirmed(camera_id, gate["session_id"]):
                gate["miss_streak"] = 0
                gate["motion_pending_verify"] = False
            elif gate["motion_pending_verify"] and meta.evidence_count >= YOLO_EVIDENCE_FRAMES and not meta.evidence_pending:
                gate["yolo_ligado"] = False
                gate["motion_pending_verify"] = False
                gate["miss_streak"] = 0
                store.reset_session(camera_id)
            elif not gate["motion_pending_verify"]:
                gate["miss_streak"] += 1
                if gate["miss_streak"] >= max(1, YOLO_MOTION_MISS_FRAMES * 5):
                    gate["yolo_ligado"] = False
                    gate["miss_streak"] = 0
                    store.reset_session(camera_id)

    cap.release()
    print(f"[CAPTURE] camera={camera_id} encerrada")
