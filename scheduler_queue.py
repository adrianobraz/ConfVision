"""Fila de jobs YOLO com prioridade (P1-P4) — Redis ou memoria."""

from __future__ import annotations

import heapq
import json
import threading
import time
from dataclasses import dataclass, field
from typing import Optional

from config import (
    REDIS_URL,
    SCHEDULER_BACKEND,
    SCHEDULER_QUEUE_KEY,
    YOLO_PRIORITY_ENABLED,
)


@dataclass(order=True)
class YoloJob:
    sort_key: tuple
    camera_id: int = field(compare=False, default=0)
    slot: str = field(compare=False, default="")
    priority: int = field(compare=False, default=4)
    session_id: int = field(compare=False, default=0)
    motion_at: float = field(compare=False, default=0.0)
    frame_path: str = field(compare=False, default="")
    enqueued_at: float = field(compare=False, default=0.0)

    @classmethod
    def create(
        cls,
        camera_id: int,
        slot: str,
        priority: int,
        session_id: int,
        motion_at: float,
        frame_path: str,
    ) -> "YoloJob":
        pri = int(priority) if YOLO_PRIORITY_ENABLED else 4
        enq = time.time()
        return cls(
            sort_key=(pri, enq),
            camera_id=int(camera_id),
            slot=str(slot),
            priority=pri,
            session_id=int(session_id),
            motion_at=float(motion_at),
            frame_path=str(frame_path),
            enqueued_at=enq,
        )

    def to_dict(self) -> dict:
        return {
            "camera_id": self.camera_id,
            "slot": self.slot,
            "priority": self.priority,
            "session_id": self.session_id,
            "motion_at": self.motion_at,
            "frame_path": self.frame_path,
            "enqueued_at": self.enqueued_at,
        }

    @classmethod
    def from_dict(cls, data: dict) -> "YoloJob":
        return cls.create(
            camera_id=int(data["camera_id"]),
            slot=str(data["slot"]),
            priority=int(data.get("priority") or 4),
            session_id=int(data.get("session_id") or 0),
            motion_at=float(data.get("motion_at") or 0.0),
            frame_path=str(data.get("frame_path") or ""),
        )


class SchedulerQueue:
    def enqueue(self, job: YoloJob) -> bool:
        raise NotImplementedError

    def pop_batch(self, max_size: int, timeout_sec: float) -> list[YoloJob]:
        raise NotImplementedError

    def cancel_camera_session(self, camera_id: int, session_id: int) -> None:
        pass


class MemorySchedulerQueue(SchedulerQueue):
    def __init__(self):
        self._heap: list[YoloJob] = []
        self._lock = threading.Lock()
        self._cond = threading.Condition(self._lock)

    def enqueue(self, job: YoloJob) -> bool:
        with self._cond:
            heapq.heappush(self._heap, job)
            self._cond.notify()
        return True

    def pop_batch(self, max_size: int, timeout_sec: float) -> list[YoloJob]:
        deadline = time.time() + timeout_sec
        batch: list[YoloJob] = []
        with self._cond:
            while len(batch) < max_size:
                if self._heap:
                    batch.append(heapq.heappop(self._heap))
                    if batch and batch[0].priority == 1 and len(batch) >= 1:
                        break
                    continue
                remaining = deadline - time.time()
                if remaining <= 0:
                    break
                self._cond.wait(timeout=remaining)
        return batch

    def cancel_camera_session(self, camera_id: int, session_id: int) -> None:
        with self._lock:
            kept: list[YoloJob] = []
            for job in self._heap:
                if job.camera_id == camera_id and job.session_id == session_id and job.slot.startswith("ev_"):
                    continue
                kept.append(job)
            self._heap = kept
            heapq.heapify(self._heap)


class RedisSchedulerQueue(SchedulerQueue):
    """Lista Redis — push LPUSH, pop jobs com ZPOPMIN simulado via sorted set."""

    def __init__(self, url: str, key: str):
        import redis

        self._client = redis.Redis.from_url(url, decode_responses=True)
        self._key = key

    def _score(self, job: YoloJob) -> float:
        return float(job.priority) * 1e12 + job.enqueued_at

    def enqueue(self, job: YoloJob) -> bool:
        try:
            self._client.zadd(self._key, {json.dumps(job.to_dict()): self._score(job)})
            return True
        except Exception as exc:
            print(f"[SCHEDULER] enqueue falhou camera={job.camera_id}: {exc}")
            return False

    def pop_batch(self, max_size: int, timeout_sec: float) -> list[YoloJob]:
        deadline = time.time() + timeout_sec
        batch: list[YoloJob] = []
        while len(batch) < max_size and time.time() < deadline:
            try:
                items = self._client.zpopmin(self._key, count=1)
            except Exception as exc:
                print(f"[SCHEDULER] pop falhou: {exc}")
                time.sleep(0.05)
                continue
            if not items:
                time.sleep(min(0.01, deadline - time.time()))
                continue
            raw, _score = items[0]
            try:
                batch.append(YoloJob.from_dict(json.loads(raw)))
            except Exception:
                continue
            if batch and batch[0].priority == 1:
                break
        return batch

    def cancel_camera_session(self, camera_id: int, session_id: int) -> None:
        try:
            members = self._client.zrange(self._key, 0, -1)
            for raw in members:
                try:
                    data = json.loads(raw)
                except Exception:
                    continue
                if (
                    int(data.get("camera_id") or 0) == camera_id
                    and int(data.get("session_id") or 0) == session_id
                    and str(data.get("slot") or "").startswith("ev_")
                ):
                    self._client.zrem(self._key, raw)
        except Exception as exc:
            print(f"[SCHEDULER] cancel session camera={camera_id}: {exc}")


_queue_singleton: Optional[SchedulerQueue] = None
_queue_lock = threading.Lock()


def get_scheduler_queue() -> SchedulerQueue:
    global _queue_singleton
    with _queue_lock:
        if _queue_singleton is not None:
            return _queue_singleton
        if SCHEDULER_BACKEND == "redis" and REDIS_URL:
            _queue_singleton = RedisSchedulerQueue(REDIS_URL, SCHEDULER_QUEUE_KEY)
            print(f"[SCHEDULER] backend=redis key={SCHEDULER_QUEUE_KEY}")
        else:
            _queue_singleton = MemorySchedulerQueue()
            print("[SCHEDULER] backend=memory")
        return _queue_singleton
