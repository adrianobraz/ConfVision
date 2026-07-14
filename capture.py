import subprocess
import threading
import time
from pathlib import Path

import cv2

from config import CAPTURE_DIR, CLIP_DURACAO_SEG, SNAPSHOT_JPEG_QUALITY

_locks: dict[int, threading.Lock] = {}
_locks_guard = threading.Lock()
_clip_procs: dict[int, subprocess.Popen] = {}
_clip_procs_guard = threading.Lock()


class CaptureCancelled(RuntimeError):
    pass


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


def cancel_camera_captures(camera_id):
    """Encerra ffmpeg de clip da camera (ex.: desativada no cadastro)."""
    cid = int(camera_id)
    with _clip_procs_guard:
        proc = _clip_procs.pop(cid, None)
    if proc is None or proc.poll() is not None:
        return
    print(f"[CAPTURA] cancelando ffmpeg camera={cid}")
    proc.terminate()
    try:
        proc.wait(timeout=5)
    except subprocess.TimeoutExpired:
        proc.kill()
        proc.wait(timeout=3)


def evento_work_dir(evento_id: int) -> Path:
    dest = Path(CAPTURE_DIR) / str(evento_id)
    dest.mkdir(parents=True, exist_ok=True)
    return dest


def cleanup_work_dir(work_dir: Path):
    import shutil

    shutil.rmtree(work_dir, ignore_errors=True)


def write_detection_snapshot(camera_id, frame) -> Path:
    """Salva o frame YOLO no instante da deteccao — sem ffmpeg, ~10ms."""
    detect_dir = Path(CAPTURE_DIR) / "detect"
    detect_dir.mkdir(parents=True, exist_ok=True)
    path = detect_dir / f"cam{camera_id}_{int(time.time() * 1000)}.jpg"
    ok = cv2.imwrite(
        str(path),
        frame,
        [int(cv2.IMWRITE_JPEG_QUALITY), SNAPSHOT_JPEG_QUALITY],
    )
    if not ok or not path.exists() or path.stat().st_size == 0:
        raise RuntimeError("cv2.imwrite falhou ao salvar snapshot da deteccao")
    return path


def install_detection_snapshot(dest_path: Path, source_path: str | Path | None) -> bool:
    """Copia snapshot pre-gravado da deteccao para a pasta do evento."""
    import shutil

    if not source_path:
        return False
    src = Path(source_path)
    if not src.exists() or src.stat().st_size == 0:
        return False
    dest_path.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(src, dest_path)
    src.unlink(missing_ok=True)
    return True


def start_clip_capture(
    rtsp_url: str,
    clip_path: str | Path,
    duration_sec: int | None = None,
    camera_id=None,
) -> tuple[threading.Thread, dict[str, Exception]]:
    errors: dict[str, Exception] = {}

    def _clip():
        try:
            record_clip_mp4_file(rtsp_url, clip_path, duration_sec, camera_id=camera_id)
        except Exception as exc:
            errors["clip"] = exc

    thread = threading.Thread(target=_clip, name="ffmpeg-clip", daemon=True)
    thread.start()
    return thread, errors


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


def record_clip_mp4_file(
    rtsp_url: str,
    dest_path: str | Path,
    duration_sec: int | None = None,
    camera_id=None,
):
    dest = Path(dest_path)
    dest.parent.mkdir(parents=True, exist_ok=True)
    duracao = duration_sec if duration_sec and duration_sec > 0 else CLIP_DURACAO_SEG
    timeout = max(90, duracao + 60)
    cmd = [
        "ffmpeg",
        "-hide_banner",
        "-loglevel",
        "error",
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
    ]
    proc = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    cid = int(camera_id) if camera_id is not None else None
    if cid is not None:
        with _clip_procs_guard:
            _clip_procs[cid] = proc
    try:
        try:
            _, stderr = proc.communicate(timeout=timeout)
        except subprocess.TimeoutExpired:
            proc.kill()
            proc.communicate(timeout=5)
            raise RuntimeError("ffmpeg timeout na gravacao do clip")
        if proc.returncode != 0:
            msg = (stderr or "").strip()
            if proc.returncode < 0:
                raise CaptureCancelled(msg or "gravacao cancelada")
            raise RuntimeError(msg or "ffmpeg falhou")
    finally:
        if cid is not None:
            with _clip_procs_guard:
                if _clip_procs.get(cid) is proc:
                    _clip_procs.pop(cid, None)
    if not dest.exists() or dest.stat().st_size == 0:
        raise RuntimeError("ffmpeg nao gerou video")


def start_parallel_capture(
    rtsp_url: str,
    snapshot_path: str | Path,
    clip_path: str | Path,
    duration_sec: int | None = None,
    camera_id=None,
) -> tuple[threading.Thread, threading.Thread, dict[str, Exception]]:
    """Inicia snapshot e gravacao do clip em paralelo (2 ffmpeg no mesmo RTSP)."""
    errors: dict[str, Exception] = {}

    def _snapshot():
        try:
            capture_snapshot_jpeg_file(rtsp_url, snapshot_path)
        except Exception as exc:
            errors["snapshot"] = exc

    def _clip():
        try:
            record_clip_mp4_file(rtsp_url, clip_path, duration_sec, camera_id=camera_id)
        except Exception as exc:
            errors["clip"] = exc

    t_snapshot = threading.Thread(target=_snapshot, name="ffmpeg-snapshot", daemon=True)
    t_clip = threading.Thread(target=_clip, name="ffmpeg-clip", daemon=True)
    t_snapshot.start()
    t_clip.start()
    return t_snapshot, t_clip, errors
