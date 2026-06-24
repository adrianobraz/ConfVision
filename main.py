import socket
import threading
import time
from pathlib import Path

from bootstrap_check import validate_config
from camera_state import is_camera_active, set_active_camera_ids
from capture_workers import start_capture_workers
from config import (
    CAPTURE_WORKERS,
    CLIP_DURACAO_SEG,
    EVENT_QUEUE_BACKEND,
    MAX_CAMERAS,
    SYNC_INTERVAL_SEC,
    WORKER_SHARD_INDEX,
    WORKER_SHARD_TOTAL,
)
from capture import cancel_camera_captures, write_detection_snapshot
from detector import PersonDetector
from event_queue import EventJob, get_event_queue
from sharding import shard_label
from urls import rtsp_url, stream_path
from xano_client import get_cameras_ativas, post_ping


def loop_camera(camera, detector: PersonDetector, event_queue):
    camera_id = camera.get("id")
    conf_min = float(camera.get("confianca_min") or 0.5)
    cooldown = int(camera.get("cooldown_seg") or 30)
    url = rtsp_url(camera_id)

    def on_person(conf, frame):
        if not is_camera_active(camera_id):
            return
        snapshot_path = None
        try:
            snapshot_path = str(write_detection_snapshot(camera_id, frame))
            print(f"[EVENTO] snapshot instantaneo camera={camera_id} path={snapshot_path}")
        except Exception as exc:
            print(f"[WARN] snapshot instantaneo camera={camera_id} falhou: {exc}")
        job = EventJob(
            camera=camera,
            confianca=conf,
            detected_at=time.time(),
            snapshot_path=snapshot_path,
        )
        if not event_queue.publish(job):
            print(f"[FILA] cheia — evento descartado camera={camera_id} conf={conf:.2f}")
            if snapshot_path:
                Path(snapshot_path).unlink(missing_ok=True)

    def should_continue():
        return is_camera_active(camera_id)

    while should_continue():
        try:
            detector.process_camera(url, conf_min, cooldown, on_person, should_continue)
        except Exception as exc:
            print(f"[ERRO] camera {camera_id}: {exc}")
            if not should_continue():
                break
            time.sleep(5)

    print(f"[THREAD] camera id={camera_id} encerrada (desativada ou sem deteccao)")


def main():
    print(
        f"[START] ConfVision worker | host={socket.gethostname()} "
        f"| clip={CLIP_DURACAO_SEG}s | {shard_label()}"
    )
    validate_config()

    event_queue = get_event_queue()
    start_capture_workers(CAPTURE_WORKERS, event_queue)

    detector = PersonDetector()
    threads: dict[int, threading.Thread] = {}

    while True:
        try:
            cameras = get_cameras_ativas()
            active_ids = [c["id"] for c in cameras]
            set_active_camera_ids(active_ids)
            post_ping(
                len(cameras),
                extra={
                    "shard_index": WORKER_SHARD_INDEX if WORKER_SHARD_INDEX >= 0 else None,
                    "shard_total": WORKER_SHARD_TOTAL if WORKER_SHARD_TOTAL > 0 else None,
                    "max_cameras": MAX_CAMERAS,
                    "yolo_device": detector.device,
                    "queue_backend": EVENT_QUEUE_BACKEND,
                },
            )
            print(f"[SYNC] {len(cameras)} camera(s) ativa(s) ids={active_ids}")

            active_set = set(active_ids)
            for camera_id, thread in list(threads.items()):
                if camera_id not in active_set:
                    cancel_camera_captures(camera_id)
                    if thread.is_alive():
                        print(
                            f"[SYNC] camera id={camera_id} desativada — "
                            f"aguardando thread encerrar"
                        )
                elif not thread.is_alive():
                    del threads[camera_id]

            for camera in cameras:
                camera_id = camera["id"]
                thread = threads.get(camera_id)
                if thread is None or not thread.is_alive():
                    thread = threading.Thread(
                        target=loop_camera,
                        args=(camera, detector, event_queue),
                        daemon=True,
                        name=f"camera-{camera_id}",
                    )
                    threads[camera_id] = thread
                    thread.start()
                    print(
                        f"[THREAD] camera id={camera_id} "
                        f"path={stream_path(camera_id)} "
                        f"nome={camera.get('nome')} url={rtsp_url(camera_id)}"
                    )
        except Exception as exc:
            print(f"[ERRO] sync: {exc}")

        time.sleep(SYNC_INTERVAL_SEC)


if __name__ == "__main__":
    main()
