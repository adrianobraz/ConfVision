from typing import Any, Optional
from urllib.parse import quote

import requests

from config import (
    DVR_MTX_SYNC_API,
    DVR_RECORD_DIR,
    DVR_SEGMENTO_MINUTOS_DEFAULT,
    MEDIAMTX_API_BASE,
    MEDIAMTX_API_PASS,
    MEDIAMTX_API_USER,
)
from urls import stream_path

# id -> segmento_minutos (estado sincronizado com MediaMTX)
RecordState = dict[int, int]


def _api_auth():
    if MEDIAMTX_API_USER and MEDIAMTX_API_PASS:
        return (MEDIAMTX_API_USER, MEDIAMTX_API_PASS)
    return None


def _api_request(method: str, url: str, **kwargs):
    auth = _api_auth()
    if auth:
        kwargs["auth"] = auth
    response = requests.request(method, url, **kwargs)
    return response


def _path_url(path_name: str) -> str:
    encoded = quote(path_name, safe="")
    return f"{MEDIAMTX_API_BASE}/v3/config/paths/patch/{encoded}"


def _segment_duration(minutes: int) -> str:
    mins = max(1, int(minutes or DVR_SEGMENTO_MINUTOS_DEFAULT))
    return f"{mins}m0s"


def record_config(segmento_minutos: int) -> dict[str, Any]:
    return {
        "record": True,
        "recordPath": f"{DVR_RECORD_DIR}/%path/%Y-%m-%d_%H-%M-%S",
        "recordFormat": "fmp4",
        "recordSegmentDuration": _segment_duration(segmento_minutos),
    }


def enable_record(camera_id: int, segmento_minutos: int = DVR_SEGMENTO_MINUTOS_DEFAULT):
    path_name = stream_path(camera_id)
    payload = record_config(segmento_minutos)
    response = _api_request("PATCH", _path_url(path_name), json=payload, timeout=10)
    if response.status_code == 404:
        add_url = f"{MEDIAMTX_API_BASE}/v3/config/paths/add/{quote(path_name, safe='')}"
        response = _api_request("POST", add_url, json=payload, timeout=10)
    response.raise_for_status()
    return response.json() if response.content else {}


def disable_record(camera_id: int):
    path_name = stream_path(camera_id)
    response = _api_request(
        "PATCH", _path_url(path_name), json={"record": False}, timeout=10
    )
    if response.status_code == 404:
        return
    response.raise_for_status()


def sync_record_paths(
    cameras: list[dict[str, Any]],
    previous: Optional[RecordState] = None,
) -> RecordState:
    """Liga/desliga record via API apenas quando a lista de cameras mudar."""
    if not DVR_MTX_SYNC_API:
        return _build_state(cameras)

    prev = previous or {}
    current = _build_state(cameras)

    for camera_id, segmento in current.items():
        if prev.get(camera_id) == segmento:
            continue
        try:
            enable_record(camera_id, segmento)
            print(
                f"[MTX] record ON camera={camera_id} "
                f"path={stream_path(camera_id)} segmento={segmento}min"
            )
        except Exception as exc:
            print(f"[MTX] ERRO record ON camera={camera_id}: {exc}")

    for camera_id in prev.keys() - current.keys():
        try:
            disable_record(camera_id)
            print(f"[MTX] record OFF camera={camera_id}")
        except Exception as exc:
            print(f"[MTX] ERRO record OFF camera={camera_id}: {exc}")

    return current


def _build_state(cameras: list[dict[str, Any]]) -> RecordState:
    state: RecordState = {}
    for camera in cameras:
        camera_id = int(camera["id"])
        segmento = int(camera.get("segmento_minutos") or DVR_SEGMENTO_MINUTOS_DEFAULT)
        state[camera_id] = segmento
    return state
