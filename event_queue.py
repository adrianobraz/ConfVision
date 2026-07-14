import json
import queue
import threading
import time
from dataclasses import dataclass
from typing import Any, Optional

from config import (
    EVENT_QUEUE_BACKEND,
    EVENT_QUEUE_KEY,
    EVENT_QUEUE_MAX_SIZE,
    REDIS_URL,
)


@dataclass
class EventJob:
    camera: dict[str, Any]
    confianca: float
    detected_at: float
    snapshot_path: Optional[str] = None

    def to_dict(self) -> dict[str, Any]:
        data = {
            "camera": self.camera,
            "confianca": self.confianca,
            "detected_at": self.detected_at,
        }
        if self.snapshot_path:
            data["snapshot_path"] = self.snapshot_path
        return data

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "EventJob":
        return cls(
            camera=data["camera"],
            confianca=float(data["confianca"]),
            detected_at=float(data.get("detected_at") or time.time()),
            snapshot_path=data.get("snapshot_path"),
        )


class EventQueue:
    def publish(self, job: EventJob) -> bool:
        raise NotImplementedError

    def pop(self, timeout_sec: float = 1.0) -> Optional[EventJob]:
        raise NotImplementedError


class MemoryEventQueue(EventQueue):
    def __init__(self, max_size: int):
        self._queue: queue.Queue[EventJob] = queue.Queue(maxsize=max_size)

    def publish(self, job: EventJob) -> bool:
        try:
            self._queue.put_nowait(job)
            return True
        except queue.Full:
            return False

    def pop(self, timeout_sec: float = 1.0) -> Optional[EventJob]:
        try:
            return self._queue.get(timeout=timeout_sec)
        except queue.Empty:
            return None


class RedisEventQueue(EventQueue):
    def __init__(self, url: str, key: str, max_size: int):
        import redis

        self._client = redis.Redis.from_url(url, decode_responses=True)
        self._key = key
        self._max_size = max_size

    def publish(self, job: EventJob) -> bool:
        try:
            if self._client.llen(self._key) >= self._max_size:
                return False
            self._client.lpush(self._key, json.dumps(job.to_dict(), ensure_ascii=False))
            return True
        except Exception as exc:
            print(f"[FILA] erro ao publicar evento camera={job.camera.get('id')}: {exc}")
            return False

    def pop(self, timeout_sec: float = 1.0) -> Optional[EventJob]:
        try:
            item = self._client.brpop(self._key, timeout=max(1, int(timeout_sec)))
            if not item:
                return None
            _, payload = item
            return EventJob.from_dict(json.loads(payload))
        except Exception as exc:
            print(f"[FILA] erro ao consumir evento: {exc}")
            return None


_queue: EventQueue | None = None
_queue_lock = threading.Lock()


def get_event_queue() -> EventQueue:
    global _queue
    with _queue_lock:
        if _queue is None:
            _queue = _build_queue()
        return _queue


def _build_queue() -> EventQueue:
    backend = EVENT_QUEUE_BACKEND
    if backend == "redis" and REDIS_URL:
        print(f"[FILA] backend=redis key={EVENT_QUEUE_KEY}")
        return RedisEventQueue(REDIS_URL, EVENT_QUEUE_KEY, EVENT_QUEUE_MAX_SIZE)
    if backend == "redis" and not REDIS_URL:
        print("[FILA] AVISO: EVENT_QUEUE_BACKEND=redis sem REDIS_URL — usando memory")
    print(f"[FILA] backend=memory max={EVENT_QUEUE_MAX_SIZE}")
    return MemoryEventQueue(EVENT_QUEUE_MAX_SIZE)
