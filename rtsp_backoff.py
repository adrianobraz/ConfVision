"""Backoff local apos falhas consecutivas de abrir RTSP (404 / offline)."""

from __future__ import annotations

import threading
import time
from datetime import datetime
from typing import Callable, Optional

from config import (
    RTSP_BACKOFF_ENABLED,
    RTSP_OPEN_BACKOFF_SEC,
    RTSP_OPEN_FAIL_MAX,
    RTSP_OPEN_RETRY_SEC,
)

_lock = threading.Lock()
_state: dict[int, dict] = {}


def _ensure(camera_id: int) -> dict:
    st = _state.get(camera_id)
    if st is None:
        st = {
            "fail_streak": 0,
            "backoff_until": 0.0,
            "backoff_logged_until": 0.0,
        }
        _state[camera_id] = st
    return st


def wait_if_backoff(
    camera_id: int,
    rtsp_url: str,
    should_continue: Optional[Callable[[], bool]] = None,
) -> bool:
    """
    Se a camera esta em backoff, aguarda ate expirar (sem abrir RTSP).
    Retorna True se saiu por should_continue=False durante a espera.
    """
    if not RTSP_BACKOFF_ENABLED or camera_id <= 0:
        return False
    if should_continue is None:
        should_continue = lambda: True

    while should_continue():
        now = time.time()
        with _lock:
            st = _ensure(camera_id)
            until = float(st["backoff_until"])
            if until <= now:
                st["backoff_logged_until"] = 0.0
                return False
            if st["backoff_logged_until"] < until:
                st["backoff_logged_until"] = until
                ate = datetime.fromtimestamp(until).strftime("%H:%M:%S")
                print(
                    f"[RTSP] camera={camera_id} em backoff ate {ate} "
                    f"({int(until - now)}s) — sem abrir url={rtsp_url}"
                )
        while should_continue() and time.time() < until:
            time.sleep(min(5.0, max(0.1, until - time.time())))
    return True


def after_open_failure(camera_id: int, rtsp_url: str) -> float:
    """Registra falha ao abrir/reabrir RTSP. Retorna segundos para sleep."""
    if not RTSP_BACKOFF_ENABLED or camera_id <= 0:
        print(f"[ERRO] Nao abriu stream RTSP: {rtsp_url}")
        return float(RTSP_OPEN_RETRY_SEC)

    now = time.time()
    entered_backoff = False
    streak = 0
    with _lock:
        st = _ensure(camera_id)
        st["fail_streak"] = int(st["fail_streak"]) + 1
        streak = st["fail_streak"]
        if streak >= RTSP_OPEN_FAIL_MAX:
            st["backoff_until"] = now + RTSP_OPEN_BACKOFF_SEC
            st["fail_streak"] = 0
            st["backoff_logged_until"] = 0.0
            entered_backoff = True
            until = st["backoff_until"]
        else:
            until = 0.0

    if entered_backoff:
        ate = datetime.fromtimestamp(until).strftime("%H:%M:%S")
        print(
            f"[RTSP] camera={camera_id} {RTSP_OPEN_FAIL_MAX} falhas seguidas — "
            f"backoff {RTSP_OPEN_BACKOFF_SEC}s ate {ate} url={rtsp_url}"
        )
        return float(RTSP_OPEN_RETRY_SEC)

    print(
        f"[ERRO] Nao abriu stream RTSP ({streak}/{RTSP_OPEN_FAIL_MAX}): "
        f"camera={camera_id} url={rtsp_url}"
    )
    return float(RTSP_OPEN_RETRY_SEC)


def after_open_success(camera_id: int, rtsp_url: str = "") -> None:
    if camera_id <= 0:
        return
    was_backoff = False
    with _lock:
        st = _state.get(camera_id)
        if not st:
            return
        was_backoff = float(st.get("backoff_until") or 0) > time.time() or int(
            st.get("fail_streak") or 0
        ) > 0
        st["fail_streak"] = 0
        st["backoff_until"] = 0.0
        st["backoff_logged_until"] = 0.0
    if was_backoff and RTSP_BACKOFF_ENABLED:
        suffix = f" url={rtsp_url}" if rtsp_url else ""
        print(f"[RTSP] camera={camera_id} stream ok — backoff encerrado{suffix}")
