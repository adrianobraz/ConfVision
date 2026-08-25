"""Scheduler GPU: monta batches dinamicos P1-P4 e delega ao YoloGpuWorker."""

from __future__ import annotations

import threading
import time
from typing import Callable, Optional

from config import YOLO_BATCH_SIZE, YOLO_BATCH_TIMEOUT_MS
from detection_handler import DetectionHandler
from frame_store import FrameStore
from scheduler_queue import get_scheduler_queue
from yolo_gpu_worker import YoloGpuWorker


class GpuScheduler:
    def __init__(
        self,
        yolo_worker: YoloGpuWorker,
        detection_handler: DetectionHandler,
        scheduler_queue=None,
        frame_store: FrameStore | None = None,
    ):
        self.yolo_worker = yolo_worker
        self.detection_handler = detection_handler
        self.queue = scheduler_queue or get_scheduler_queue()
        self.frame_store = frame_store or FrameStore()
        self._stop = threading.Event()
        self._thread: Optional[threading.Thread] = None

    def start(self) -> None:
        if self._thread and self._thread.is_alive():
            return
        self._stop.clear()
        self._thread = threading.Thread(target=self._loop, daemon=True, name="gpu-scheduler")
        self._thread.start()
        print(
            f"[GPU-SCHED] iniciado batch_size={YOLO_BATCH_SIZE} "
            f"timeout_ms={YOLO_BATCH_TIMEOUT_MS}"
        )

    def stop(self) -> None:
        self._stop.set()

    def _loop(self) -> None:
        timeout_sec = max(0.001, YOLO_BATCH_TIMEOUT_MS / 1000.0)
        while not self._stop.is_set():
            try:
                jobs = self.queue.pop_batch(YOLO_BATCH_SIZE, timeout_sec)
                if not jobs:
                    continue
                valid_jobs = [
                    j for j in jobs
                    if self.frame_store.is_job_valid(j.camera_id, j.slot, j.session_id, j.motion_at)
                ]
                if not valid_jobs:
                    continue
                t0 = time.time()
                batch_results = self.yolo_worker.infer_batch(valid_jobs)
                infer_ms = (time.time() - t0) * 1000
                for job, results, frame in batch_results:
                    self.detection_handler.process_result(job, results, frame)
                if batch_results:
                    pri = min(j.priority for j, _, _ in batch_results)
                    print(
                        f"[GPU-SCHED] batch={len(batch_results)} infer_ms={infer_ms:.0f} "
                        f"min_pri=P{pri}"
                    )
            except Exception as exc:
                print(f"[GPU-SCHED] erro: {exc}")
                time.sleep(0.1)


def start_gpu_pipeline(
    event_queue,
    get_camera: Callable[[int], Optional[dict]],
    frame_store: FrameStore | None = None,
) -> GpuScheduler:
    store = frame_store or FrameStore()
    sched_q = get_scheduler_queue()
    yolo = YoloGpuWorker(store)
    handler = DetectionHandler(event_queue, store, sched_q, get_camera)
    scheduler = GpuScheduler(yolo, handler, sched_q, store)
    scheduler.start()
    return scheduler
