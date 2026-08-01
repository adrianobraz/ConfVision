from datetime import datetime, timezone
from typing import Any, Optional

import requests

from config import WORKER_ID, WORKER_VERSION, XANO_BASE_URL, MEDIAMTX_NODE_ID
from sharding import filter_cameras, query_params


def _parse_json(response):
    response.raise_for_status()
    return response.json()


def _as_list(data):
    if isinstance(data, list):
        return data
    if isinstance(data, dict):
        payload = data.get("dados", data)
        if isinstance(payload, list):
            return payload
    return []


def get_areas_ativas():
    url = f"{XANO_BASE_URL}/vis_camera_area_query_ativas"
    try:
        response = requests.get(url, timeout=30)
        return _as_list(_parse_json(response))
    except Exception as exc:
        print(f"[WARN] vis_camera_area_query_ativas falhou: {exc}")
        return []


def _attach_areas(cameras, areas):
    by_camera: dict[int, list] = {}
    for area in areas:
        camera_id = area.get("vis_camera_id")
        if camera_id is None:
            continue
        by_camera.setdefault(int(camera_id), []).append(area)
    for camera in cameras:
        camera_id = camera.get("id")
        if camera_id is None:
            camera["areas"] = []
            continue
        camera["areas"] = by_camera.get(int(camera_id), [])
    return cameras


def get_cameras_ativas():
    url = f"{XANO_BASE_URL}/vis_camera_query_ativas"
    params = query_params()
    response = requests.get(url, params=params, timeout=30)
    data = _parse_json(response)

    cameras = _as_list(data)
    areas = get_areas_ativas()
    cameras = _attach_areas(cameras, areas)

    return filter_cameras(cameras)


def create_evento(camera, confianca, tipo="humano", status="capturando", extra=None):
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
    if extra:
        payload.update(extra)
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


def post_ping(cameras_ativas: int, extra: Optional[dict[str, Any]] = None):
    url = f"{XANO_BASE_URL}/vis_worker_ping"
    payload = {
        "worker_id": WORKER_ID,
        "hostname": WORKER_ID,
        "versao": WORKER_VERSION,
        "cameras_ativas": cameras_ativas,
        "ultimo_ping_em": datetime.now(timezone.utc).isoformat(),
        "ativo": True,
    }
    if MEDIAMTX_NODE_ID > 0:
        payload["vis_mediamtx_node_id"] = MEDIAMTX_NODE_ID
    if extra:
        payload.update(extra)
    response = requests.post(url, json=payload, timeout=15)
    return _parse_json(response)


def post_evento(camera, confianca, tipo="humano"):
    """Compatibilidade — preferir create_evento + captura."""
    return create_evento(camera, confianca, tipo=tipo, status="capturando")


def get_cameras_gravacao_ativas():
    url = f"{XANO_BASE_URL}/vis_camera_query_gravacao_ativas"
    params = query_params()
    response = requests.get(url, params=params, timeout=30)
    return _as_list(_parse_json(response))


def get_gravacao_storage_credenciais(id_franqueado: str):
    url = f"{XANO_BASE_URL}/vis_gravacao_storage_credenciais_by_franqueado"
    response = requests.get(
        url, params={"id_franqueado": id_franqueado}, timeout=15
    )
    return _parse_json(response)


def post_gravacao_segmento(payload: dict[str, Any]):
    url = f"{XANO_BASE_URL}/vis_gravacao_segmento"
    response = requests.post(url, json=payload, timeout=30)
    return _parse_json(response)


def ack_flush_pedido(camera_id: int) -> bool:
    """Limpa o flag gravacao_flush_pedido após o worker executar o flush."""
    try:
        url = f"{XANO_BASE_URL}/vis_camera/gravacao/flush/{camera_id}"
        # Chama a mesma API de flush com ack=true para limpar o flag
        url_ack = f"{XANO_BASE_URL}/vis_camera/gravacao/flush/ack/{camera_id}?vis_camera_id={camera_id}"
        response = requests.post(url_ack, json={}, timeout=15)
        return response.ok
    except Exception as exc:
        print(f"[XANO] ack_flush_pedido camera={camera_id} erro: {exc}")
        return False


def get_eventos_sensor_pendentes(limit: int = 10):
    url = f"{XANO_BASE_URL}/vis_evento_query_sensor_pendentes"
    response = requests.get(url, params={"limit": limit}, timeout=15)
    return _as_list(_parse_json(response))


def get_camera_by_id(camera_id: int):
    url = f"{XANO_BASE_URL}/vis_camera/{camera_id}"
    response = requests.get(url, params={"vis_camera_id": camera_id}, timeout=15)
    return _parse_json(response)
