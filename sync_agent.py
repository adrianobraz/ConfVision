"""Agente de sync: busca Xano 1x e distribui via Redis para todos os workers."""

from __future__ import annotations

import time
from typing import Any, Optional

import requests

from config import (
    CONFIG_CACHE_BACKEND,
    MEDIAMTX_NODE_ID,
    SYNC_INTERVAL_SEC,
    SYNC_USE_UNIFIED_API,
    XANO_BASE_URL,
)
from config_cache import (
    cameras_from_sync_payload,
    get_cached_version,
    gravacao_from_sync_payload,
    normalize_sync_payload,
    use_unified_api,
    write_sync,
)
from sharding import query_params


def _parse_json(response):
    response.raise_for_status()
    return response.json()


def fetch_unified_sync(since_version: Optional[str] = None) -> dict[str, Any]:
    url = f"{XANO_BASE_URL}/vis_camera_sync_ativas"
    params = dict(query_params())
    params["include_gravacao"] = "true"
    if since_version:
        params["since_version"] = since_version
    try:
        response = requests.get(url, params=params, timeout=45)
        if response.status_code == 404:
            print("[SYNC-AGENT] API 2399 nao encontrada — fallback legacy")
            return fetch_legacy_sync()
        data = _parse_json(response)
    except Exception as exc:
        print(f"[SYNC-AGENT] unified falhou ({exc}) — fallback legacy")
        return fetch_legacy_sync()
    if isinstance(data, dict) and data.get("dados"):
        data = data["dados"]
    if isinstance(data, dict):
        return normalize_sync_payload(data)
    return {}


def fetch_legacy_sync() -> dict[str, Any]:
    from xano_client import _as_list, _attach_areas, get_areas_ativas

    params = query_params()
    url_cam = f"{XANO_BASE_URL}/vis_camera_query_ativas"
    response = requests.get(url_cam, params=params, timeout=45)
    cameras = _as_list(_parse_json(response))
    areas = get_areas_ativas()
    cameras = _attach_areas(cameras, areas)

    url_grav = f"{XANO_BASE_URL}/vis_camera_query_gravacao_ativas"
    response_g = requests.get(url_grav, params=params, timeout=45)
    gravacao = _as_list(_parse_json(response_g))

    payload = normalize_sync_payload(
        {
            "unchanged": False,
            "config_version": f"{len(cameras)}:{len(areas)}:{len(gravacao)}",
            "cameras": cameras,
            "areas": areas,
            "gravacao": gravacao,
        }
    )
    return payload


def run_sync_cycle() -> dict[str, Any]:
    since = get_cached_version("full")
    if use_unified_api():
        payload = fetch_unified_sync(since_version=since)
    else:
        payload = fetch_legacy_sync()

    if payload.get("unchanged"):
        return payload

    write_sync("full", payload)
    write_sync("analitico", payload)
    write_sync("gravacao", payload)
    return payload


def get_cameras_ativas_cached() -> list:
    from config_cache import read_sync

    payload = read_sync("analitico")
    if payload:
        return cameras_from_sync_payload(payload)

    payload = run_sync_cycle()
    if payload.get("unchanged"):
        payload = read_sync("analitico") or {}
    return cameras_from_sync_payload(payload)


def get_cameras_gravacao_ativas_cached() -> list:
    from config_cache import read_sync

    payload = read_sync("gravacao")
    if payload:
        return gravacao_from_sync_payload(payload)

    payload = run_sync_cycle()
    if payload.get("unchanged"):
        payload = read_sync("gravacao") or {}
    return gravacao_from_sync_payload(payload)


def agent_loop(interval_sec: Optional[int] = None):
    interval = interval_sec or SYNC_INTERVAL_SEC
    node = MEDIAMTX_NODE_ID or "all"
    api_mode = "2399-unified" if SYNC_USE_UNIFIED_API else "legacy"
    cache_mode = CONFIG_CACHE_BACKEND
    print(
        f"[SYNC-AGENT] iniciado interval={interval}s node={node} "
        f"api={api_mode} cache={cache_mode}"
    )
    while True:
        try:
            payload = run_sync_cycle()
            if payload.get("unchanged"):
                print(f"[SYNC-AGENT] unchanged version={payload.get('config_version')}")
            else:
                cams = len(payload.get("cameras") or [])
                grav = len(payload.get("gravacao") or [])
                print(
                    f"[SYNC-AGENT] atualizado version={payload.get('config_version')} "
                    f"cameras={cams} gravacao={grav}"
                )
        except Exception as exc:
            print(f"[SYNC-AGENT] erro: {exc}")
        time.sleep(interval)
