"""Orchestrator modo distributed: sync cameras + capture MOG2 + pipeline GPU batch."""

from __future__ import annotations

import socket
import threading
import time
from typing import Optional

from bootstrap_check import validate_config
from camera_state import is_camera_active, set_active_camera_ids
from capture import cancel_camera_captures
from capture_workers import start_capture_workers
from capture_worker import run_capture_camera
from config import (
    CAPTURE_WORKERS,
    CLIP_DURACAO_SEG,
    CONFIG_CACHE_BACKEND,
    EVENT_QUEUE_BACKEND,
    EVENT_STORE,
    FRAME_STORE_DIR,
    MAX_CAMERAS,
    POSTGRES_URL,
    SYNC_INTERVAL_SEC,
    WORKER_SHARD_INDEX,
    WORKER_SHARD_TOTAL,
    YOLO_ARCH,
    YOLO_BATCH_SIZE,
    YOLO_EVIDENCE_FRAMES,
)
from area_utils import camera_deve_rodar_thread, camera_elegivel_analitico, normalize_modo_deteccao
from device_armed import prefetch_armado
from event_queue import get_event_queue
from frame_store import FrameStore
from gpu_scheduler import start_gpu_pipeline
from sharding import shard_label
from urls import rtsp_url_for_camera, stream_path
from xano_client import get_cameras_ativas, post_ping

_restart_ids: set[int] = set()
_thread_cfg: dict[int, str] = {}
_camera_configs: dict[int, dict] = {}
_camera_configs_lock = threading.Lock()


def _camera_cfg_key(camera: dict) -> str:
    areas = camera.get("areas") or []
    area_ids = sorted(
        str(a.get("id"))
        for a in areas
        if a.get("ativo", True) and a.get("id") is not None
    )
    return "|".join(
        [
            normalize_modo_deteccao(camera.get("modo_deteccao")),
            str(camera.get("confianca_min") or ""),
            str(camera.get("cooldown_seg") or ""),
            str(bool(camera.get("somente_armado"))),
            ",".join(area_ids),
        ]
    )


def get_camera_config(camera_id: int) -> Optional[dict]:
    with _camera_configs_lock:
        return _camera_configs.get(int(camera_id))


def _update_camera_configs(cameras: list[dict]) -> None:
    with _camera_configs_lock:
        _camera_configs.clear()
        for cam in cameras:
            cid = cam.get("id")
            if cid is not None:
                _camera_configs[int(cid)] = cam


def loop_capture_camera(camera, frame_store: FrameStore):
    camera_id = camera.get("id")

    def should_continue():
        if camera_id in _restart_ids:
            return False
        if not is_camera_active(camera_id):
            return False
        if camera.get("analitico_pausado"):
            return False
        if camera.get("somente_armado"):
            from device_armed import is_dispositivo_armado

            id_disp = str(camera.get("id_dispositivo") or "").strip()
            if not id_disp or not is_dispositivo_armado(id_disp):
                return False
        return True

    while should_continue():
        try:
            run_capture_camera(camera, should_continue, frame_store=frame_store)
        except Exception as exc:
            print(f"[CAPTURE] erro camera {camera_id}: {exc}")
            if not should_continue():
                break
            time.sleep(5)

    _restart_ids.discard(camera_id)
    print(f"[CAPTURE] thread camera id={camera_id} encerrada")


def main():
    print(
        f"[START] ConfVision distributed | arch={YOLO_ARCH} | host={socket.gethostname()} "
        f"| clip={CLIP_DURACAO_SEG}s | evidence={YOLO_EVIDENCE_FRAMES} | batch={YOLO_BATCH_SIZE} "
        f"| frame_store={FRAME_STORE_DIR} | {shard_label()}"
    )
    validate_config()

    if EVENT_STORE in ("postgres", "dual") and POSTGRES_URL:
        import postgres_store

        postgres_store.init_schema()
        postgres_store.start_sync_worker()

    if CONFIG_CACHE_BACKEND == "redis":
        print("[START] modo cache Redis — sync via sync_agent_main ou fallback inline")

    frame_store = FrameStore()
    event_queue = get_event_queue()
    start_capture_workers(CAPTURE_WORKERS, event_queue)
    start_gpu_pipeline(event_queue, get_camera_config, frame_store)

    threads: dict[int, threading.Thread] = {}

    while True:
        try:
            cameras = get_cameras_ativas()
            cameras_elegiveis = [c for c in cameras if camera_elegivel_analitico(c)]
            cameras_com_area = [c for c in cameras_elegiveis if camera_deve_rodar_thread(c)]
            _update_camera_configs(cameras_com_area)

            ids_somente_armado = {
                str(c.get("id_dispositivo")).strip()
                for c in cameras_elegiveis
                if c.get("somente_armado") and c.get("id_dispositivo")
            }
            if ids_somente_armado:
                prefetch_armado(ids_somente_armado)

            active_ids = [c["id"] for c in cameras_com_area]
            set_active_camera_ids(active_ids)
            post_ping(
                len(cameras_com_area),
                extra={
                    "shard_index": WORKER_SHARD_INDEX if WORKER_SHARD_INDEX >= 0 else None,
                    "shard_total": WORKER_SHARD_TOTAL if WORKER_SHARD_TOTAL > 0 else None,
                    "max_cameras": MAX_CAMERAS,
                    "yolo_arch": YOLO_ARCH,
                    "queue_backend": EVENT_QUEUE_BACKEND,
                },
            )
            print(
                f"[SYNC] {len(cameras_com_area)} camera(s) distributed ids={active_ids} "
                f"({len(cameras_elegiveis) - len(cameras_com_area)} pausadas/desarmadas, "
                f"{len(cameras) - len(cameras_elegiveis)} ignoradas)"
            )

            active_set = set(active_ids)
            for camera_id, thread in list(threads.items()):
                if camera_id not in active_set:
                    cancel_camera_captures(camera_id)
                    frame_store.reset_session(camera_id)
                    _thread_cfg.pop(camera_id, None)
                    if thread.is_alive():
                        print(f"[SYNC] camera id={camera_id} desativada — aguardando thread")
                elif not thread.is_alive():
                    del threads[camera_id]
                    _thread_cfg.pop(camera_id, None)

            for camera in cameras_com_area:
                camera_id = camera["id"]
                cfg_key = _camera_cfg_key(camera)
                thread = threads.get(camera_id)
                if thread is not None and thread.is_alive():
                    if _thread_cfg.get(camera_id) != cfg_key:
                        print(
                            f"[SYNC] camera id={camera_id} config alterada — reiniciando capture"
                        )
                        _restart_ids.add(camera_id)
                        continue
                if thread is None or not thread.is_alive():
                    _restart_ids.discard(camera_id)
                    thread = threading.Thread(
                        target=loop_capture_camera,
                        args=(camera, frame_store),
                        daemon=True,
                        name=f"capture-{camera_id}",
                    )
                    threads[camera_id] = thread
                    _thread_cfg[camera_id] = cfg_key
                    thread.start()
                    print(
                        f"[CAPTURE] camera id={camera_id} "
                        f"path={stream_path(camera_id, camera.get('id_franqueado'))} "
                        f"nome={camera.get('nome')} url={rtsp_url_for_camera(camera)}"
                    )
        except Exception as exc:
            print(f"[ERRO] sync: {exc}")

        time.sleep(SYNC_INTERVAL_SEC)


if __name__ == "__main__":
    main()
