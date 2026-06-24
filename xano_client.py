from datetime import datetime, timezone
from typing import Any, Optional

import requests

from config import WORKER_ID, WORKER_VERSION, XANO_BASE_URL


def _parse_json(response):
    response.raise_for_status()
    return response.json()


def get_cameras_ativas():
    url = f"{XANO_BASE_URL}/vis_camera_query_ativas"
    response = requests.get(url, timeout=15)
    data = _parse_json(response)
    cameras = data.get("dados", data) if isinstance(data, dict) else data
    return [
        camera
        for camera in cameras
        if camera.get("ativo") and camera.get("deteccao_humano")
    ]


def create_evento(camera, confianca, tipo="humano", status="capturando"):
    url = f"{XANO_BASE_URL}/vis_evento"
    payload = {
        "vis_camera_id": camera["id"],
        "id_franqueado": camera.get("id_franqueado"),
        "id_cliente": camera.get("id_cliente"),
        "id_dispositivo": camera.get("id_dispositivo"),
        "conta": camera.get("conta"),
        "particao": camera.get("particao"),
        "canal": camera.get("canal"),
        "tipo_deteccao": tipo,
        "confianca": confianca,
        "processado": False,
        "ignorado": False,
        "status": status,
        "clip_count": 0,
    }
    response = requests.post(url, json=payload, timeout=15)
    return _parse_json(response)


def put_evento(evento_id: int, campos: dict[str, Any], base: Optional[dict[str, Any]] = None):
    payload = _merge_evento_payload(base or {}, campos)
    url = f"{XANO_BASE_URL}/vis_evento/{evento_id}"
    params = {"id": evento_id}
    response = requests.put(url, params=params, json=payload, timeout=15)
    return _parse_json(response)


def _merge_evento_payload(base: dict[str, Any], campos: dict[str, Any]) -> dict[str, Any]:
    keys = [
        "vis_camera_id",
        "id_franqueado",
        "id_cliente",
        "id_dispositivo",
        "conta",
        "particao",
        "canal",
        "tipo_deteccao",
        "confianca",
        "snapshot_url",
        "video_url",
        "bbox_json",
        "processado",
        "alarm_events_id",
        "ignorado",
        "status",
        "id_evento",
        "id_processo",
        "started_at",
        "ended_at",
        "clip_count",
    ]
    merged = {}
    for key in keys:
        if key in campos:
            merged[key] = campos[key]
        elif key in base:
            merged[key] = base[key]
    return merged


def post_evento_clip(payload: dict[str, Any]):
    url = f"{XANO_BASE_URL}/vis_evento_clip"
    response = requests.post(url, json=payload, timeout=15)
    return _parse_json(response)


def post_ping(cameras_ativas: int):
    url = f"{XANO_BASE_URL}/vis_worker_ping"
    payload = {
        "worker_id": WORKER_ID,
        "hostname": WORKER_ID,
        "versao": WORKER_VERSION,
        "cameras_ativas": cameras_ativas,
        "ultimo_ping_em": datetime.now(timezone.utc).isoformat(),
        "ativo": True,
    }
    response = requests.post(url, json=payload, timeout=15)
    return _parse_json(response)


def post_evento(camera, confianca, tipo="humano"):
    """Compatibilidade — preferir create_evento + captura."""
    return create_evento(camera, confianca, tipo=tipo, status="capturando")
