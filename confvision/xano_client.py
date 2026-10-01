from datetime import datetime, timezone
from typing import Any, Optional

import requests

from config import (
    CONFIG_CACHE_BACKEND,
    MEDIAMTX_NODE_ID,
    REDIS_URL,
    SYNC_USE_UNIFIED_API,
    WORKER_ID,
    WORKER_VERSION,
    XANO_BASE_URL,
)
from sharding import filter_cameras, query_params
from vis_api_auth import vis_api_headers


def _api_get(url, **kwargs):
    headers = vis_api_headers(kwargs.pop("headers", None))
    return requests.get(url, headers=headers, **kwargs)


def _api_post(url, **kwargs):
    headers = vis_api_headers(kwargs.pop("headers", None))
    return requests.post(url, headers=headers, **kwargs)


def _api_put(url, **kwargs):
    headers = vis_api_headers(kwargs.pop("headers", None))
    return requests.put(url, headers=headers, **kwargs)


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


def _cache_enabled() -> bool:
    return CONFIG_CACHE_BACKEND == "redis" and bool(REDIS_URL)


def get_areas_ativas():
    url = f"{XANO_BASE_URL}/vis_camera_area_query_ativas"
    try:
        response = _api_get(url, timeout=30)
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


def get_cameras_ativas_direct():
    """Busca direta no Xano (legacy — preferir cache/sync unificado)."""
    if SYNC_USE_UNIFIED_API:
        url = f"{XANO_BASE_URL}/vis_camera_sync_ativas"
        params = dict(query_params())
        response = _api_get(url, params=params, timeout=45)
        data = _parse_json(response)
        if isinstance(data, dict) and data.get("dados"):
            data = data["dados"]
        cameras = data.get("cameras") or []
        areas = data.get("areas") or []
        return filter_cameras(_attach_areas(cameras, areas))

    url = f"{XANO_BASE_URL}/vis_camera_query_ativas"
    params = query_params()
    response = _api_get(url, params=params, timeout=30)
    data = _parse_json(response)
    cameras = _as_list(data)
    areas = get_areas_ativas()
    return filter_cameras(_attach_areas(cameras, areas))


def get_cameras_ativas():
    if _cache_enabled():
        from sync_agent import get_cameras_ativas_cached

        return get_cameras_ativas_cached()
    return get_cameras_ativas_direct()


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
    response = _api_post(url, json=payload, timeout=15)
    return _parse_json(response)


def put_evento(evento_id: int, campos: dict[str, Any], base: Optional[dict[str, Any]] = None):
    payload = _merge_evento_payload(base or {}, campos)
    url = f"{XANO_BASE_URL}/vis_evento/{evento_id}"
    params = {"id": evento_id}
    response = _api_put(url, params=params, json=payload, timeout=15)
    return _parse_json(response)


def finalizar_evento(
    evento_id: int,
    *,
    snapshot_url: Optional[str] = None,
    video_url: Optional[str] = None,
    status: str = "pronto",
    clip_count: int = 0,
    clip_duracao_seg: Optional[int] = None,
    processado: bool = False,
):
    """1 request: atualiza evento + cria clip (API 2400). Fallback legacy se 404."""
    url = f"{XANO_BASE_URL}/vis_evento_finalizar"
    payload = {
        "vis_evento_id": evento_id,
        "snapshot_url": snapshot_url,
        "video_url": video_url,
        "status": status,
        "clip_count": clip_count,
        "processado": processado,
        "clip_seq": 1,
        "clip_duracao_seg": clip_duracao_seg,
        "clip_snapshot_url": snapshot_url,
    }
    try:
        response = _api_post(url, json=payload, timeout=20)
        if response.status_code == 404:
            raise requests.HTTPError("404")
        return _parse_json(response)
    except requests.HTTPError:
        if video_url:
            post_evento_clip(
                {
                    "vis_evento_id": evento_id,
                    "seq": 1,
                    "video_url": video_url,
                    "duracao_seg": clip_duracao_seg,
                    "snapshot_url": snapshot_url,
                }
            )
        return put_evento(
            evento_id,
            {
                "snapshot_url": snapshot_url,
                "video_url": video_url,
                "clip_count": clip_count,
                "status": status,
                "processado": processado,
            },
        )


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
    response = _api_post(url, json=payload, timeout=15)
    return _parse_json(response)


def post_ping(cameras_ativas: int, extra: Optional[dict[str, Any]] = None):
    url = f"{XANO_BASE_URL}/vis_worker_ping"
    payload = {
        "worker_id": WORKER_ID,
        "worker_tipo": "analitico",
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
    if not payload.get("worker_tipo"):
        payload["worker_tipo"] = "analitico"
    response = _api_post(url, json=payload, timeout=15)
    return _parse_json(response)


def post_evento(camera, confianca, tipo="humano"):
    """Compatibilidade — preferir create_evento + captura."""
    return create_evento(camera, confianca, tipo=tipo, status="capturando")


def get_cameras_gravacao_ativas_direct():
    url = f"{XANO_BASE_URL}/vis_camera_query_gravacao_ativas"
    params = query_params()
    response = _api_get(url, params=params, timeout=30)
    return _as_list(_parse_json(response))


def get_cameras_gravacao_ativas():
    if _cache_enabled():
        from sync_agent import get_cameras_gravacao_ativas_cached

        return get_cameras_gravacao_ativas_cached()
    return get_cameras_gravacao_ativas_direct()


def get_gravacao_storage_credenciais(id_franqueado: str):
    url = f"{XANO_BASE_URL}/vis_gravacao_storage_credenciais_by_franqueado"
    response = _api_get(
        url, params={"id_franqueado": id_franqueado}, timeout=15
    )
    return _parse_json(response)


def post_gravacao_segmento(payload: dict[str, Any]):
    url = f"{XANO_BASE_URL}/vis_gravacao_segmento"
    response = _api_post(url, json=payload, timeout=30)
    return _parse_json(response)


def ack_flush_pedido(camera_id: int) -> bool:
    try:
        url_ack = f"{XANO_BASE_URL}/vis_camera/gravacao/flush/ack/{camera_id}?vis_camera_id={camera_id}"
        response = _api_post(url_ack, json={}, timeout=15)
        return response.ok
    except Exception as exc:
        print(f"[XANO] ack_flush_pedido camera={camera_id} erro: {exc}")
        return False


def get_eventos_sensor_pendentes(limit: int = 10):
    url = f"{XANO_BASE_URL}/vis_evento_query_sensor_pendentes"
    response = _api_get(url, params={"limit": limit}, timeout=15)
    return _as_list(_parse_json(response))


def get_camera_by_id(camera_id: int):
    url = f"{XANO_BASE_URL}/vis_camera/{camera_id}"
    response = _api_get(url, params={"vis_camera_id": camera_id}, timeout=15)
    return _parse_json(response)
