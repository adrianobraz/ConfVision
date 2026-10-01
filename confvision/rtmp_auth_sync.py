"""Pre-popula cache Redis de auth RTMP (sync-agent → mediamtx guard)."""

from __future__ import annotations

import requests

from config import MEDIAMTX_NODE_ID, RTMP_AUTH_CACHE_SEC, XANO_BASE_URL
from config_cache import write_rtmp_auth_batch
from vis_api_auth import vis_api_headers


def sync_rtmp_auth_cache() -> int:
    """Busca cameras elegiveis no Go API e grava no Redis. Retorna total gravado."""
    if not XANO_BASE_URL:
        return 0
    url = f"{XANO_BASE_URL}/vis_camera_rtmp_auth_sync"
    params: dict[str, str | int] = {}
    if MEDIAMTX_NODE_ID and int(MEDIAMTX_NODE_ID) > 0:
        params["vis_mediamtx_node_id"] = int(MEDIAMTX_NODE_ID)
    try:
        response = requests.get(
            url, params=params, headers=vis_api_headers(), timeout=45
        )
        response.raise_for_status()
        data = response.json()
    except Exception as exc:
        print(f"[SYNC-AGENT] rtmp_auth_sync falhou: {exc}")
        return 0
    records = data.get("dados") if isinstance(data, dict) else None
    if not isinstance(records, list):
        return 0
    count = write_rtmp_auth_batch(records, RTMP_AUTH_CACHE_SEC)
    if count:
        print(f"[SYNC-AGENT] rtmp_auth cache={count} ttl={RTMP_AUTH_CACHE_SEC}s node={MEDIAMTX_NODE_ID or 'all'}")
    return count
