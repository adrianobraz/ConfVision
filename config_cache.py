"""Cache Redis da config de cameras (sync coordinator → workers)."""

from __future__ import annotations

import hashlib
import json
import time
from typing import Any, Optional

from config import (
    CONFIG_CACHE_BACKEND,
    CONFIG_CACHE_TTL_SEC,
    MEDIAMTX_NODE_ID,
    REDIS_URL,
    SYNC_USE_UNIFIED_API,
)
from redis_client import get_redis, redis_available
from sharding import filter_cameras, filter_gravacao_cameras, query_params

KEY_PREFIX = "confvision:sync"
RTMP_AUTH_PREFIX = "confvision:rtmp_auth"


def _backend() -> str:
    if CONFIG_CACHE_BACKEND == "redis" and REDIS_URL and redis_available():
        return "redis"
    return "memory"


_memory_store: dict[str, tuple[float, dict]] = {}


def _cache_key(kind: str) -> str:
    params = query_params()
    node = params.get("vis_mediamtx_node_id") or MEDIAMTX_NODE_ID or 0
    worker = params.get("worker_id") or "all"
    return f"{KEY_PREFIX}:{kind}:n{node}:w{worker}"


def _version_key(kind: str) -> str:
    return f"{_cache_key(kind)}:version"


def _hash_payload(payload: dict) -> str:
    canonical = json.dumps(payload, sort_keys=True, ensure_ascii=False, default=str)
    return hashlib.sha256(canonical.encode("utf-8")).hexdigest()[:16]


def write_sync(kind: str, payload: dict[str, Any]) -> str:
    """Grava sync no Redis/memoria. Retorna config_version."""
    version = payload.get("config_version") or _hash_payload(payload)
    envelope = {
        "config_version": version,
        "fetched_at": time.time(),
        "payload": payload,
    }
    key = _cache_key(kind)
    if _backend() == "redis":
        client = get_redis()
        assert client is not None
        pipe = client.pipeline()
        pipe.setex(key, CONFIG_CACHE_TTL_SEC, json.dumps(envelope, ensure_ascii=False))
        pipe.setex(_version_key(kind), CONFIG_CACHE_TTL_SEC, version)
        pipe.execute()
    else:
        _memory_store[key] = (time.time(), envelope)
    return version


def read_sync(kind: str) -> Optional[dict[str, Any]]:
    key = _cache_key(kind)
    envelope = None
    if _backend() == "redis":
        client = get_redis()
        assert client is not None
        raw = client.get(key)
        if raw:
            envelope = json.loads(raw)
    else:
        hit = _memory_store.get(key)
        if hit and time.time() - hit[0] <= CONFIG_CACHE_TTL_SEC:
            envelope = hit[1]
    if not envelope:
        return None
    return envelope.get("payload")


def get_cached_version(kind: str) -> Optional[str]:
    if _backend() == "redis":
        client = get_redis()
        assert client is not None
        return client.get(_version_key(kind))
    key = _cache_key(kind)
    hit = _memory_store.get(key)
    if hit:
        return hit[1].get("config_version")
    return None


def write_rtmp_auth(camera_id: int, cam: dict, ttl_sec: int) -> None:
    if _backend() != "redis":
        return
    client = get_redis()
    assert client is not None
    key = f"{RTMP_AUTH_PREFIX}:{camera_id}"
    client.setex(key, max(5, ttl_sec), json.dumps(cam, ensure_ascii=False))


def read_rtmp_auth(camera_id: int) -> Optional[dict]:
    if _backend() != "redis":
        return None
    client = get_redis()
    assert client is not None
    raw = client.get(f"{RTMP_AUTH_PREFIX}:{camera_id}")
    if not raw:
        return None
    try:
        return json.loads(raw)
    except json.JSONDecodeError:
        return None


def invalidate_rtmp_auth(camera_id: int) -> None:
    if _backend() != "redis":
        return
    client = get_redis()
    assert client is not None
    client.delete(f"{RTMP_AUTH_PREFIX}:{camera_id}")


def attach_areas(cameras: list, areas: list) -> list:
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


def cameras_from_sync_payload(payload: dict) -> list:
    cameras = payload.get("cameras") or []
    areas = payload.get("areas") or []
    return filter_cameras(attach_areas(cameras, areas))


def gravacao_from_sync_payload(payload: dict) -> list:
    gravacao = payload.get("gravacao") or []
    return filter_gravacao_cameras(gravacao, motion=False)


def use_unified_api() -> bool:
    return SYNC_USE_UNIFIED_API
