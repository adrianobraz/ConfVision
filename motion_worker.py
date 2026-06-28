import os
import signal
import subprocess
import threading
import time
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Optional

import cv2

from config import (
    MOTION_ANALYSIS_WIDTH,
    MOTION_CLIP_MAX_SEC,
    MOTION_FRAME_SKIP,
    MOTION_MIN_AREA,
    MOTION_MOG2_HISTORY,
    MOTION_MOG2_THRESHOLD,
    MOTION_POST_ROLL_SEC,
    MOTION_RECONNECT_SEC,
    MOTION_RECORD_DIR,
)
from dvr_segment import process_segment_file
from urls import rtsp_url_for_camera

os.environ.setdefault("OPENCV_FFMPEG_CAPTURE_OPTIONS", "rtsp_transport;tcp")


class MotionCameraWorker:
    def __init__(self, camera: dict[str, Any]):
        self.camera = camera
        self.camera_id = int(camera["id"])
        self._stop = threading.Event()
        self._thread: Optional[threading.Thread] = None

    def start(self):
        if self._thread and self._thread.is_alive():
            return
        self._stop.clear()
        self._thread = threading.Thread(
            target=self._run,
            daemon=True,
            name=f"motion-{self.camera_id}",
        )
        self._thread.start()

    def stop(self, timeout: float = 20):
        self._stop.set()
        if self._thread:
            self._thread.join(timeout=timeout)

    def is_alive(self) -> bool:
        return bool(self._thread and self._thread.is_alive())

    def _run(self):
        while not self._stop.is_set():
            try:
                self._process_stream()
            except Exception as exc:
                print(f"[MOTION] camera={self.camera_id} ERRO loop: {exc}")
            if not self._stop.is_set():
                self._stop.wait(MOTION_RECONNECT_SEC)

    def _open_capture(self, url: str):
        cap = cv2.VideoCapture(url, cv2.CAP_FFMPEG)
        cap.set(cv2.CAP_PROP_BUFFERSIZE, 1)
        return cap

    def _detect_motion(self, frame, fgbg, kernel) -> bool:
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

    def _start_ffmpeg(self, rtsp_url: str):
        out_dir = Path(MOTION_RECORD_DIR) / str(self.camera_id)
        out_dir.mkdir(parents=True, exist_ok=True)
        clip_start = datetime.now(timezone.utc)
        ts = clip_start.strftime("%Y-%m-%d_%H-%M-%S")
        clip_path = out_dir / f"{ts}.mp4"
        cmd = [
            "ffmpeg",
            "-hide_banner",
            "-loglevel",
            "error",
            "-rtsp_transport",
            "tcp",
            "-i",
            rtsp_url,
            "-c:v",
            "libx264",
            "-preset",
            "veryfast",
            "-crf",
            "28",
            "-an",
            "-movflags",
            "+faststart",
            "-y",
            str(clip_path),
        ]
        proc = subprocess.Popen(
            cmd,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        return clip_start, clip_path, proc

    def _stop_ffmpeg(self, proc: subprocess.Popen, clip_path: Path, timeout: int = 45):
        if proc.poll() is not None:
            return
        try:
            proc.send_signal(signal.SIGINT)
            proc.wait(timeout=timeout)
        except subprocess.TimeoutExpired:
            proc.kill()
            proc.communicate(timeout=5)
        except Exception as exc:
            print(f"[MOTION] camera={self.camera_id} erro ao parar ffmpeg: {exc}")
            try:
                proc.kill()
            except Exception:
                pass

        if not clip_path.exists() or clip_path.stat().st_size == 0:
            print(f"[MOTION] camera={self.camera_id} clip vazio: {clip_path.name}")
            clip_path.unlink(missing_ok=True)
            return

    def _upload_clip(self, clip_path: Path, clip_start: datetime, clip_end: datetime):
        def _task():
            try:
                ok = process_segment_file(
                    clip_path,
                    self.camera,
                    tipo="movimento",
                    inicio_em=clip_start,
                    fim_em=clip_end,
                )
                if ok:
                    try:
                        clip_path.unlink(missing_ok=True)
                    except OSError:
                        pass
            except Exception as exc:
                print(
                    f"[MOTION] camera={self.camera_id} upload falhou "
                    f"file={clip_path.name}: {exc}"
                )

        threading.Thread(
            target=_task,
            daemon=True,
            name=f"motion-upload-{self.camera_id}",
        ).start()

    def _process_stream(self):
        url = rtsp_url_for_camera(self.camera)
        cap = self._open_capture(url)
        if not cap.isOpened():
            print(
                f"[MOTION] camera={self.camera_id} falhou abrir RTSP: {url} "
                f"(verifique RTMP live/{self.camera_id} no MediaMTX ou rtsp_url_sec na camera)"
            )
            time.sleep(MOTION_RECONNECT_SEC)
            return

        print(f"[MOTION] camera={self.camera_id} stream OK url={url}")
        fgbg = cv2.createBackgroundSubtractorMOG2(
            history=MOTION_MOG2_HISTORY,
            varThreshold=MOTION_MOG2_THRESHOLD,
            detectShadows=True,
        )
        kernel = cv2.getStructuringElement(cv2.MORPH_ELLIPSE, (3, 3))

        frame_idx = 0
        recording = False
        last_motion = 0.0
        clip_start: Optional[datetime] = None
        clip_path: Optional[Path] = None
        ffmpeg_proc: Optional[subprocess.Popen] = None
        falhas = 0

        def finalize_recording(force_motion: bool = False):
            nonlocal recording, clip_start, clip_path, ffmpeg_proc, last_motion
            if not recording or ffmpeg_proc is None or clip_path is None or clip_start is None:
                return False

            proc = ffmpeg_proc
            path = clip_path
            inicio = clip_start
            ffmpeg_proc = None
            clip_path = None
            clip_start = None
            recording = False

            clip_end = datetime.now(timezone.utc)
            print(
                f"[MOTION] REC STOP camera={self.camera_id} "
                f"file={path.name} dur={(clip_end - inicio).total_seconds():.0f}s"
            )
            self._stop_ffmpeg(proc, path)
            self._upload_clip(path, inicio, clip_end)

            if force_motion:
                last_motion = time.time()
            return True

        def start_recording():
            nonlocal recording, clip_start, clip_path, ffmpeg_proc
            if recording:
                return
            clip_start, clip_path, ffmpeg_proc = self._start_ffmpeg(url)
            recording = True
            print(
                f"[MOTION] REC START camera={self.camera_id} file={clip_path.name}"
            )

        while not self._stop.is_set():
            ok, frame = cap.read()
            if not ok:
                falhas += 1
                print(
                    f"[MOTION] camera={self.camera_id} frame perdido ({falhas}x), "
                    f"reconectando"
                )
                if recording:
                    finalize_recording()
                cap.release()
                time.sleep(MOTION_RECONNECT_SEC)
                return

            falhas = 0
            frame_idx += 1
            now = time.time()

            motion = False
            if frame_idx % MOTION_FRAME_SKIP == 0:
                motion = self._detect_motion(frame, fgbg, kernel)
                if motion:
                    last_motion = now

            if motion and not recording:
                start_recording()

            if recording and clip_start is not None:
                elapsed = now - clip_start.timestamp()
                if elapsed >= MOTION_CLIP_MAX_SEC:
                    still_moving = motion or (now - last_motion) < MOTION_POST_ROLL_SEC
                    finalize_recording(force_motion=still_moving)
                    if still_moving and not self._stop.is_set():
                        start_recording()
                elif not motion and (now - last_motion) >= MOTION_POST_ROLL_SEC:
                    finalize_recording()

        cap.release()
        if recording:
            finalize_recording()


class MotionWorkerManager:
    def __init__(self):
        self._workers: dict[int, MotionCameraWorker] = {}
        self._lock = threading.Lock()

    def sync(self, cameras: list[dict[str, Any]]):
        active_ids = {int(c["id"]) for c in cameras}

        with self._lock:
            for camera_id, worker in list(self._workers.items()):
                if camera_id not in active_ids:
                    print(f"[MOTION] camera={camera_id} removida — parando worker")
                    worker.stop()
                    del self._workers[camera_id]

            for camera in cameras:
                camera_id = int(camera["id"])
                worker = self._workers.get(camera_id)
                if worker is None or not worker.is_alive():
                    if worker is not None:
                        worker.stop(timeout=5)
                    worker = MotionCameraWorker(camera)
                    self._workers[camera_id] = worker
                    worker.start()
                    print(
                        f"[MOTION] worker iniciado camera={camera_id} "
                        f"nome={camera.get('nome')}"
                    )

    def stop_all(self):
        with self._lock:
            for worker in self._workers.values():
                worker.stop()
            self._workers.clear()

    def count(self) -> int:
        with self._lock:
            return len(self._workers)
