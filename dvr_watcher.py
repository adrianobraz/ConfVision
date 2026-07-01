import threading
import time
from pathlib import Path
from typing import Any, Callable, Optional

from config import DVR_RECORD_DIR, DVR_STABLE_SEC
from dvr_segment import process_segment_file
from urls import stream_path


class DvrWatcher:
    def __init__(
        self,
        get_cameras: Callable[[], list[dict[str, Any]]],
        poll_interval: float = 5.0,
    ):
        self._get_cameras = get_cameras
        self._poll_interval = poll_interval
        self._stop = threading.Event()
        self._thread: Optional[threading.Thread] = None
        self._stable: dict[str, tuple[int, float]] = {}
        self._processing: set[str] = set()
        self._lock = threading.Lock()

    def start(self):
        if self._thread and self._thread.is_alive():
            return
        self._thread = threading.Thread(
            target=self._loop, daemon=True, name="dvr-watcher"
        )
        self._thread.start()
        print(f"[DVR] watcher iniciado dir={DVR_RECORD_DIR}")

    def stop(self):
        self._stop.set()
        if self._thread:
            self._thread.join(timeout=10)

    def _cameras_by_id(self) -> dict[int, dict[str, Any]]:
        result: dict[int, dict[str, Any]] = {}
        for camera in self._get_cameras():
            camera_id = camera.get("id")
            if camera_id is not None:
                result[int(camera_id)] = camera
        return result

    def _is_stable(self, path: Path) -> bool:
        key = str(path)
        try:
            size = path.stat().st_size
        except OSError:
            return False

        now = time.time()
        with self._lock:
            prev = self._stable.get(key)
            if prev and prev[0] == size:
                if now - prev[1] >= DVR_STABLE_SEC:
                    return True
            else:
                self._stable[key] = (size, now)
        return False

    def _scan_camera_dir(self, camera_id: int, camera: dict[str, Any]):
        cam_dir = Path(DVR_RECORD_DIR) / stream_path(camera_id)
        if not cam_dir.is_dir():
            return

        segmento = int(camera.get("segmento_minutos") or 5)
        for file_path in sorted(cam_dir.iterdir()):
            if not file_path.is_file():
                continue
            suffix = file_path.suffix.lower()
            if suffix in (".part", ".tmp", ".uploaded"):
                continue
            if suffix not in (".mp4", ".fmp4", ".m4s", ""):
                continue

            key = str(file_path)
            with self._lock:
                if key in self._processing:
                    continue

            if not self._is_stable(file_path):
                continue

            with self._lock:
                self._processing.add(key)

            try:
                process_segment_file(file_path, camera, segmento_minutos=segmento)
            except Exception as exc:
                print(f"[DVR] ERRO processar {file_path}: {exc}")
            finally:
                with self._lock:
                    self._processing.discard(key)
                    self._stable.pop(key, None)

    def _loop(self):
        while not self._stop.is_set():
            try:
                cameras = self._cameras_by_id()
                for camera_id, camera in cameras.items():
                    self._scan_camera_dir(camera_id, camera)
            except Exception as exc:
                print(f"[DVR] watcher ERRO: {exc}")
            self._stop.wait(self._poll_interval)
