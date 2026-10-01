"""Envia pause_analytic ao ConfVision (Go) via POST /vis_worker_ping."""

from __future__ import annotations

import os
import threading
import time
from typing import Optional

import requests

# Postgres `stream_motivo_pausa` — prefixo sistema_stream obrigatório.
PAUSE_REASON_VIDEO_TRACK = "sistema_stream_video_track_not_set_up"
ERROR_CLASS_VIDEO_TRACK = "VIDEO_TRACK_NOT_SET_UP"

_lock = threading.Lock()
_last_pause_at: dict[int, float] = {}


def _env(name: str, default: str = "") -> str:
    return (os.getenv(name) or default).strip()


def confvision_api_base() -> str:
    base = (
        _env("CONFVISION_API_URL")
        or _env("CONFVISION_VISDATA_URL")
        or _env("VISDATA_BASE_URL")
        or _env("XANO_BASE_URL")
    )
    return base.rstrip("/")


def vis_worker_auth_headers() -> dict[str, str]:
    key = _env("VIS_WORKER_API_KEY")
    if not key:
        return {}
    return {
        "Authorization": f"Bearer {key}",
        "X-Vis-Worker-Key": key,
    }


def rtmp_auth_request_headers(api_base: str) -> dict[str, str]:
    """Worker auth exigido pelo Go em GET /vis_camera/rtmp_auth/{id}."""
    go = confvision_api_base()
    if not go or api_base.rstrip("/") != go:
        return {}
    return vis_worker_auth_headers()


def _dedupe_sec() -> float:
    try:
        return max(60.0, float(_env("RTMP_STREAM_PAUSE_DEDUPE_SEC", "3600") or "3600"))
    except ValueError:
        return 3600.0


def _pause_dedupe_allows(camera_id: int) -> bool:
    now = time.time()
    with _lock:
        last = _last_pause_at.get(camera_id, 0.0)
        return now - last >= _dedupe_sec()


def _mark_pause_sent(camera_id: int) -> None:
    with _lock:
        _last_pause_at[camera_id] = time.time()


def reset_pause_dedupe_for_tests() -> None:
    with _lock:
        _last_pause_at.clear()


def stream_health_configured() -> bool:
    return bool(confvision_api_base() and _env("VIS_WORKER_API_KEY"))


def pause_camera_video_track(
    camera_id: int,
    *,
    last_error: str,
    path: str = "",
) -> bool:
    """Pausa analítico por video track não configurado. Retorna True se POST enviado."""
    if camera_id < 1:
        return False
    if not _pause_dedupe_allows(camera_id):
        return False

    base = confvision_api_base()
    key = _env("VIS_WORKER_API_KEY")
    if not base or not key:
        print(
            "[RTMP-GUARD] stream_health: ERRO — CONFVISION_API_URL ou VIS_WORKER_API_KEY "
            f"ausente; camera_id={camera_id} (pause_analytic não enviado)",
            flush=True,
        )
        return False

    msg = (last_error or "").strip()
    if path:
        msg = f"{msg} path={path}".strip()

    body = {
        "worker_id": _env("RTMP_GUARD_WORKER_ID", "rtmp-guard"),
        "worker_tipo": "rtmp_guard",
        "hostname": _env("HOSTNAME", "rtmp-guard"),
        "versao": _env("RTMP_GUARD_VERSION", "1"),
        "cameras_ativas": 0,
        "camera_stream_health": [
            {
                "camera_id": int(camera_id),
                "event": "pause_analytic",
                "pause_reason": PAUSE_REASON_VIDEO_TRACK,
                "error_class": ERROR_CLASS_VIDEO_TRACK,
                "last_error": msg[:500] if msg else (
                    "received a packet for video track 0, but track is not set up"
                ),
            }
        ],
    }
    url = f"{base}/vis_worker_ping"
    headers = {"Content-Type": "application/json", **vis_worker_auth_headers()}
    try:
        r = requests.post(url, json=body, headers=headers, timeout=12)
        if r.status_code >= 300:
            print(
                f"[RTMP-GUARD] stream_health: ERRO HTTP {r.status_code} POST /vis_worker_ping "
                f"camera_id={camera_id} body={r.text[:200]}",
                flush=True,
            )
            return False
        _mark_pause_sent(camera_id)
        print(
            f"[RTMP-GUARD] stream_health: pause_analytic OK camera_id={camera_id} "
            f"motivo={PAUSE_REASON_VIDEO_TRACK} error_class={ERROR_CLASS_VIDEO_TRACK}",
            flush=True,
        )
        return True
    except Exception as exc:
        print(
            f"[RTMP-GUARD] stream_health: ERRO HTTP POST /vis_worker_ping "
            f"camera_id={camera_id}: {exc}",
            flush=True,
        )
        return False
