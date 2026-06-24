import subprocess
import threading
from pathlib import Path

from config import CAPTURE_DIR, CLIP_DURACAO_SEG

_locks: dict[int, threading.Lock] = {}
_locks_guard = threading.Lock()


def _lock_camera(camera_id) -> threading.Lock:
    cid = int(camera_id)
    with _locks_guard:
        if cid not in _locks:
            _locks[cid] = threading.Lock()
        return _locks[cid]


def try_iniciar_captura(camera_id) -> bool:
    lock = _lock_camera(camera_id)
    return lock.acquire(blocking=False)


def finalizar_captura(camera_id):
    lock = _lock_camera(camera_id)
    if lock.locked():
        lock.release()


def evento_work_dir(evento_id: int) -> Path:
    dest = Path(CAPTURE_DIR) / str(evento_id)
    dest.mkdir(parents=True, exist_ok=True)
    return dest


def cleanup_work_dir(work_dir: Path):
    import shutil

    shutil.rmtree(work_dir, ignore_errors=True)


def _run_ffmpeg(args, timeout_sec=60):
    result = subprocess.run(
        ["ffmpeg", "-hide_banner", "-loglevel", "error", *args],
        capture_output=True,
        text=True,
        timeout=timeout_sec,
    )
    if result.returncode != 0:
        msg = (result.stderr or result.stdout or "").strip()
        raise RuntimeError(msg or "ffmpeg falhou")
    return result


def capture_snapshot_jpeg_file(rtsp_url: str, dest_path: str | Path):
    dest = Path(dest_path)
    dest.parent.mkdir(parents=True, exist_ok=True)

    _run_ffmpeg(
        [
            "-rtsp_transport",
            "tcp",
            "-i",
            rtsp_url,
            "-frames:v",
            "1",
            "-update",
            "1",
            "-y",
            str(dest),
        ],
        timeout_sec=30,
    )
    if not dest.exists() or dest.stat().st_size == 0:
        raise RuntimeError("ffmpeg nao retornou imagem")


def record_clip_mp4_file(rtsp_url: str, dest_path: str | Path, duration_sec: int | None = None):
    dest = Path(dest_path)
    dest.parent.mkdir(parents=True, exist_ok=True)
    duracao = duration_sec if duration_sec and duration_sec > 0 else CLIP_DURACAO_SEG
    timeout = max(90, duracao + 60)

    _run_ffmpeg(
        [
            "-rtsp_transport",
            "tcp",
            "-i",
            rtsp_url,
            "-t",
            str(duracao),
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
            str(dest),
        ],
        timeout_sec=timeout,
    )
    if not dest.exists() or dest.stat().st_size == 0:
        raise RuntimeError("ffmpeg nao gerou video")
