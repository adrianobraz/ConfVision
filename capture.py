import subprocess
import tempfile
import threading
from pathlib import Path

from config import CLIP_DURACAO_SEG

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


def capture_snapshot_jpeg(rtsp_url: str) -> bytes:
    with tempfile.NamedTemporaryFile(suffix=".jpg", delete=False) as tmp:
        path = tmp.name

    try:
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
                path,
            ],
            timeout_sec=30,
        )
        data = Path(path).read_bytes()
        if not data:
            raise RuntimeError("ffmpeg nao retornou imagem")
        return data
    finally:
        Path(path).unlink(missing_ok=True)


def record_clip_mp4(rtsp_url: str, duration_sec: int | None = None) -> bytes:
    duracao = duration_sec if duration_sec and duration_sec > 0 else CLIP_DURACAO_SEG
    with tempfile.NamedTemporaryFile(suffix=".mp4", delete=False) as tmp:
        path = tmp.name

    timeout = max(90, duracao + 60)
    args_encode = [
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
        path,
    ]

    try:
        _run_ffmpeg(args_encode, timeout_sec=timeout)
        data = Path(path).read_bytes()
        if not data:
            raise RuntimeError("ffmpeg nao gerou video")
        return data
    finally:
        Path(path).unlink(missing_ok=True)
