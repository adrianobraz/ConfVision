import threading
import traceback
from pathlib import Path

from camera_state import is_camera_active
from event_capture import processar_deteccao
from event_queue import EventJob, EventQueue


def start_capture_workers(count: int, event_queue: EventQueue) -> list[threading.Thread]:
    threads: list[threading.Thread] = []
    for index in range(max(1, count)):
        thread = threading.Thread(
            target=_capture_worker_loop,
            args=(event_queue, index + 1),
            daemon=True,
            name=f"capture-worker-{index + 1}",
        )
        thread.start()
        threads.append(thread)
    print(f"[CAPTURE] pool iniciado workers={len(threads)}")
    return threads


def _capture_worker_loop(event_queue: EventQueue, worker_no: int):
    while True:
        job = event_queue.pop(timeout_sec=1.0)
        if job is None:
            continue
        camera_id = job.camera.get("id")
        if not is_camera_active(camera_id):
            print(f"[CAPTURA] worker={worker_no} camera={camera_id} inativa — job descartado")
            if job.snapshot_path:
                Path(job.snapshot_path).unlink(missing_ok=True)
            continue
        try:
            print(
                f"[CAPTURE] worker={worker_no} camera={camera_id} "
                f"conf={job.confianca:.2f} fila=ok"
            )
            processar_deteccao(job.camera, job.confianca, job.snapshot_path)
        except Exception as exc:
            print(f"[ERRO] capture worker={worker_no} camera={camera_id}: {exc}")
            traceback.print_exc()
