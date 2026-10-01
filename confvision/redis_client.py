"""Cliente Redis compartilhado (fila, cache de config, auth RTMP)."""

from __future__ import annotations

import threading
from typing import Optional

from config import REDIS_URL

_client = None
_lock = threading.Lock()


def get_redis():
    global _client
    if not REDIS_URL:
        return None
    with _lock:
        if _client is None:
            import redis

            _client = redis.Redis.from_url(REDIS_URL, decode_responses=True)
        return _client


def redis_available() -> bool:
    client = get_redis()
    if client is None:
        return False
    try:
        client.ping()
        return True
    except Exception:
        return False
