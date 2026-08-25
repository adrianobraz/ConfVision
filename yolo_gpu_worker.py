"""Inferencia YOLO em batch — um modelo GPU, sem lock global por camera."""

from __future__ import annotations

import os
from typing import Optional

import cv2
from ultralytics import YOLO

from config import YOLO_DEVICE, YOLO_MODEL
from frame_store import FrameStore

os.environ.setdefault("OPENCV_FFMPEG_CAPTURE_OPTIONS", "rtsp_transport;tcp")


class YoloGpuWorker:
    def __init__(self, frame_store: FrameStore | None = None):
        self.model = YOLO(YOLO_MODEL)
        self.device = self._resolve_device()
        self.frame_store = frame_store or FrameStore()
        print(f"[YOLO-GPU] model={YOLO_MODEL} device={self.device} batch=sim")

    def _resolve_device(self) -> str:
        if YOLO_DEVICE:
            return YOLO_DEVICE
        try:
            import torch

            if torch.cuda.is_available():
                name = torch.cuda.get_device_name(0)
                print(f"[YOLO-GPU] GPU detectada: {name}")
                return "cuda:0"
        except Exception:
            pass
        return "cpu"

    def _load_frame(self, job) -> Optional[tuple]:
        if not self.frame_store.is_job_valid(
            job.camera_id, job.slot, job.session_id, job.motion_at
        ):
            return None
        path = job.frame_path or self.frame_store.frame_path(job.camera_id, job.slot)
        frame = cv2.imread(path)
        if frame is None:
            return None
        return job, frame

    def infer_batch(self, jobs: list) -> list[tuple]:
        """Retorna lista de (job, results, frame) para jobs validos."""
        loaded = []
        frames = []
        for job in jobs:
            item = self._load_frame(job)
            if item is None:
                continue
            loaded.append(item)
            frames.append(item[1])

        if not frames:
            return []

        results_list = self.model(frames, device=self.device, verbose=False)
        out = []
        for (job, frame), results in zip(loaded, results_list):
            out.append((job, results, frame))
        return out
