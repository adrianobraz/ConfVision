"""FrameStore: 3 frames de evidencia protegidos + 1 latest por camera."""

from __future__ import annotations

import json
import threading
import time
from dataclasses import asdict, dataclass, field
from pathlib import Path
from typing import Optional

import cv2

from config import FRAME_STORE_DIR, SNAPSHOT_JPEG_QUALITY, YOLO_EVIDENCE_FRAMES


@dataclass
class CameraFrameMeta:
    camera_id: int
    session_id: int = 0
    state: str = "idle"
    priority: int = 4
    motion_at: float = 0.0
    person_confirmed: bool = False
    evidence_count: int = 0
    evidence_pending: list[int] = field(default_factory=list)
    last_updated: float = 0.0

    def to_dict(self) -> dict:
        return asdict(self)

    @classmethod
    def from_dict(cls, data: dict) -> "CameraFrameMeta":
        return cls(
            camera_id=int(data["camera_id"]),
            session_id=int(data.get("session_id") or 0),
            state=str(data.get("state") or "idle"),
            priority=int(data.get("priority") or 4),
            motion_at=float(data.get("motion_at") or 0.0),
            person_confirmed=bool(data.get("person_confirmed")),
            evidence_count=int(data.get("evidence_count") or 0),
            evidence_pending=[int(x) for x in (data.get("evidence_pending") or [])],
            last_updated=float(data.get("last_updated") or 0.0),
        )


class FrameStore:
    """Armazena JPEGs em disco (ideal: /dev/shm) — compartilhavel entre processos."""

    def __init__(self, base_dir: str | None = None):
        self.base_dir = Path(base_dir or FRAME_STORE_DIR)
        self.base_dir.mkdir(parents=True, exist_ok=True)
        self._locks_guard = threading.Lock()
        self._locks: dict[int, threading.Lock] = {}

    def _lock(self, camera_id: int) -> threading.Lock:
        cid = int(camera_id)
        with self._locks_guard:
            if cid not in self._locks:
                self._locks[cid] = threading.Lock()
            return self._locks[cid]

    def _cam_dir(self, camera_id: int) -> Path:
        d = self.base_dir / str(int(camera_id))
        d.mkdir(parents=True, exist_ok=True)
        return d

    def _meta_path(self, camera_id: int) -> Path:
        return self._cam_dir(camera_id) / "meta.json"

    def _slot_path(self, camera_id: int, slot: str) -> Path:
        return self._cam_dir(camera_id) / f"{slot}.jpg"

    def read_meta(self, camera_id: int) -> CameraFrameMeta:
        path = self._meta_path(camera_id)
        if not path.exists():
            return CameraFrameMeta(camera_id=int(camera_id))
        try:
            data = json.loads(path.read_text(encoding="utf-8"))
            return CameraFrameMeta.from_dict(data)
        except Exception:
            return CameraFrameMeta(camera_id=int(camera_id))

    def _write_meta(self, meta: CameraFrameMeta) -> None:
        meta.last_updated = time.time()
        path = self._meta_path(meta.camera_id)
        tmp = path.with_suffix(".json.tmp")
        tmp.write_text(json.dumps(meta.to_dict()), encoding="utf-8")
        tmp.replace(path)

    def _write_jpeg(self, path: Path, frame) -> None:
        tmp = path.with_suffix(".jpg.tmp")
        ok = cv2.imwrite(
            str(tmp),
            frame,
            [int(cv2.IMWRITE_JPEG_QUALITY), SNAPSHOT_JPEG_QUALITY],
        )
        if not ok or not tmp.exists() or tmp.stat().st_size == 0:
            raise RuntimeError(f"cv2.imwrite falhou: {path}")
        tmp.replace(path)

    def start_motion_session(self, camera_id: int, frame, motion_at: float | None = None) -> CameraFrameMeta:
        """Movimento novo: nova sessao, primeiro frame de evidencia (ev_0)."""
        cid = int(camera_id)
        now = motion_at if motion_at is not None else time.time()
        with self._lock(cid):
            meta = self.read_meta(cid)
            meta.session_id += 1
            meta.state = "collecting"
            meta.priority = 1
            meta.motion_at = now
            meta.person_confirmed = False
            meta.evidence_count = 1
            meta.evidence_pending = [0]
            self._write_jpeg(self._slot_path(cid, "ev_0"), frame)
            self._write_jpeg(self._slot_path(cid, "latest"), frame)
            self._write_meta(meta)
            return meta

    def append_evidence(self, camera_id: int, slot_index: int, frame) -> CameraFrameMeta:
        cid = int(camera_id)
        slot = f"ev_{int(slot_index)}"
        with self._lock(cid):
            meta = self.read_meta(cid)
            if meta.person_confirmed:
                return meta
            if slot_index in meta.evidence_pending:
                return meta
            self._write_jpeg(self._slot_path(cid, slot), frame)
            meta.evidence_count = max(meta.evidence_count, slot_index + 1)
            if slot_index not in meta.evidence_pending:
                meta.evidence_pending.append(slot_index)
            if meta.evidence_count >= YOLO_EVIDENCE_FRAMES:
                meta.state = "evidence_pending"
            self._write_jpeg(self._slot_path(cid, "latest"), frame)
            self._write_meta(meta)
            return meta

    def put_latest(self, camera_id: int, frame, priority: int | None = None) -> CameraFrameMeta:
        cid = int(camera_id)
        with self._lock(cid):
            meta = self.read_meta(cid)
            self._write_jpeg(self._slot_path(cid, "latest"), frame)
            if priority is not None:
                meta.priority = int(priority)
            self._write_meta(meta)
            return meta

    def mark_evidence_processed(self, camera_id: int, slot_index: int, session_id: int) -> None:
        cid = int(camera_id)
        with self._lock(cid):
            meta = self.read_meta(cid)
            if meta.session_id != session_id:
                return
            if slot_index in meta.evidence_pending:
                meta.evidence_pending.remove(slot_index)
            if not meta.evidence_pending and not meta.person_confirmed:
                meta.state = "motion_continuous"
                meta.priority = 3
            self._write_meta(meta)

    def mark_person_confirmed(self, camera_id: int, session_id: int) -> None:
        cid = int(camera_id)
        with self._lock(cid):
            meta = self.read_meta(cid)
            if meta.session_id != session_id:
                return
            meta.person_confirmed = True
            meta.evidence_pending = []
            meta.state = "person_track"
            meta.priority = 2
            self._write_meta(meta)

    def is_person_confirmed(self, camera_id: int, session_id: int | None = None) -> bool:
        meta = self.read_meta(camera_id)
        if session_id is not None and meta.session_id != session_id:
            return False
        return meta.person_confirmed

    def reset_session(self, camera_id: int) -> None:
        cid = int(camera_id)
        with self._lock(cid):
            meta = CameraFrameMeta(camera_id=cid)
            self._write_meta(meta)

    def frame_path(self, camera_id: int, slot: str) -> str:
        return str(self._slot_path(int(camera_id), slot))

    def load_frame(self, camera_id: int, slot: str):
        path = self._slot_path(int(camera_id), slot)
        if not path.exists():
            return None
        return cv2.imread(str(path))

    def is_job_valid(self, camera_id: int, slot: str, session_id: int, motion_at: float) -> bool:
        meta = self.read_meta(camera_id)
        if meta.session_id != session_id:
            return False
        if meta.person_confirmed and slot.startswith("ev_"):
            return False
        if slot.startswith("ev_"):
            try:
                idx = int(slot.split("_", 1)[1])
            except (IndexError, ValueError):
                return False
            if idx not in meta.evidence_pending and meta.person_confirmed:
                return False
        path = self._slot_path(camera_id, slot)
        if not path.exists():
            return False
        age_ms = (time.time() - path.stat().st_mtime) * 1000
        from config import YOLO_MAX_FRAME_AGE_MS

        if age_ms > YOLO_MAX_FRAME_AGE_MS:
            return False
        return True
