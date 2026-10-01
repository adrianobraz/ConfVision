"""Rotação segura do log MediaMTX (copytruncate) em container long-running."""

from __future__ import annotations

import os
import shutil
import time
from pathlib import Path


def _env_int(name: str, default: int) -> int:
    raw = (os.getenv(name) or "").strip()
    if not raw:
        return default
    try:
        return max(0, int(raw))
    except ValueError:
        return default


def log_rotate_config() -> tuple[int, int]:
    """(max_bytes, keep_rotated). 0 max_bytes = desligado."""
    max_bytes = _env_int("RTMP_LOG_MAX_BYTES", 52_428_800)  # 50 MiB
    keep = _env_int("RTMP_LOG_KEEP_ROTATED", 3)
    return max_bytes, max(1, min(keep, 10))


def maybe_rotate_log(path: str, *, min_interval_sec: float = 120.0) -> bool:
    """
    Se o arquivo passar de RTMP_LOG_MAX_BYTES, copia para .1..N e trunca o original.
    MediaMTX mantém o fd aberto — copytruncate é o padrão seguro aqui.
    """
    max_bytes, keep = log_rotate_config()
    if max_bytes <= 0:
        return False

    p = Path(path)
    if not p.is_file():
        return False

    stamp_path = p.with_name(p.name + ".rotate_stamp")
    now = time.time()
    try:
        if stamp_path.is_file():
            last = float(stamp_path.read_text(encoding="utf-8").strip() or "0")
            if now - last < min_interval_sec:
                return False
    except (OSError, ValueError):
        pass

    try:
        size = p.stat().st_size
    except OSError:
        return False
    if size < max_bytes:
        return False

    for i in range(keep, 0, -1):
        dst = p.with_name(f"{p.name}.{i}")
        if i == keep and dst.is_file():
            try:
                dst.unlink()
            except OSError:
                pass
        if i > 1:
            src = p.with_name(f"{p.name}.{i - 1}")
            if src.is_file():
                try:
                    src.replace(dst)
                except OSError:
                    pass

    backup = p.with_name(f"{p.name}.1")
    try:
        shutil.copy2(p, backup)
        with p.open("w", encoding="utf-8"):
            pass
        stamp_path.write_text(str(now), encoding="utf-8")
        print(
            f"[LOG-ROTATE] {p} truncado (era {size} bytes); backup={backup.name} keep={keep}",
            flush=True,
        )
        return True
    except OSError as exc:
        print(f"[LOG-ROTATE] falha em {p}: {exc}", flush=True)
        return False
