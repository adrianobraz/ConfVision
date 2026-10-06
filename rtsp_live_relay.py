"""Relay on-demand RTSP (cliente) → RTMP cam/{hash} para ao vivo HLS."""

from __future__ import annotations

import os
import re
import subprocess
import threading
import time
import urllib.parse
from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Any, Optional

import requests

from config import (
    MEDIAMTX_RTMP_PUBLISH_BASE,
    RTSP_LIVE_RELAY_ENABLED,
    RTSP_LIVE_RELAY_IDLE_SEC,
    RTSP_LIVE_RELAY_POLL_SEC,
    WORKER_ID,
    XANO_BASE_URL,
)
from rtmp_online import _api_get as mtx_api_get
from rtmp_token import chave_rtmp, stream_path_from_chave
from vis_api_auth import vis_api_headers


def _env(name: str, default: str = "") -> str:
    return (os.getenv(name) or default).strip()


def _parse_wake_until(raw: Any) -> Optional[float]:
    if raw is None or raw == "":
        return None
    s = str(raw).strip()
    if not s or s == "<nil>":
        return None
    if s.endswith("Z"):
        s = s[:-1] + "+00:00"
    try:
        dt = datetime.fromisoformat(s)
        if dt.tzinfo is None:
            dt = dt.replace(tzinfo=timezone.utc)
        return dt.timestamp()
    except ValueError:
        return None


def _wake_active(cam: dict[str, Any], now: float) -> bool:
    until = _parse_wake_until(cam.get("rtsp_live_wake_until"))
    return until is not None and until > now


def _rtmp_dest(camera: dict[str, Any]) -> str:
    base = str(camera.get("mediamtx_rtmp_public") or "").strip().rstrip("/")
    if not base:
        base = MEDIAMTX_RTMP_PUBLISH_BASE.rstrip("/")
    cid = int(camera.get("id") or 0)
    path = stream_path_from_chave(chave_rtmp(cid))
    if not base or not path:
        return ""
    url = f"{base}/{path}"
    if str(camera.get("protocolo") or "").strip().lower() == "dvr":
        url += "/"
    return url


def _path_name(camera: dict[str, Any]) -> str:
    cid = int(camera.get("id") or 0)
    return stream_path_from_chave(chave_rtmp(cid))


def path_reader_count(path: str) -> int:
    if not path:
        return 0
    enc = urllib.parse.quote(path, safe="")
    data = mtx_api_get(f"/v3/paths/get/{enc}")
    if not data:
        return 0
    item = data.get("item") if isinstance(data.get("item"), dict) else data
    if not isinstance(item, dict):
        return 0
    readers = item.get("readers") or []
    if isinstance(readers, list):
        return len(readers)
    return 0


def _sanitize_ffmpeg_stderr(raw: bytes | str, limit: int = 800) -> str:
    text = raw.decode("utf-8", errors="replace") if isinstance(raw, bytes) else (raw or "")
    text = re.sub(r"rtsp://[^@\s/]+@", "rtsp://***@", text, flags=re.I)
    text = re.sub(r"rtmp://[^\s]+", "rtmp://***", text, flags=re.I)
    text = " ".join(text.split())
    if len(text) > limit:
        return text[:limit] + "..."
    return text


def fetch_rtsp_pull_cameras() -> list[dict[str, Any]]:
    base = XANO_BASE_URL.rstrip("/")
    if not base:
        print("[RTSP-RELAY] XANO_BASE_URL vazio", flush=True)
        return []
    url = f"{base}/vis_camera_query_rtsp_pull"
    params: dict[str, str] = {}
    if WORKER_ID:
        params["worker_id"] = WORKER_ID
    try:
        resp = requests.get(
            url, params=params, headers=vis_api_headers(), timeout=30
        )
        resp.raise_for_status()
        data = resp.json()
    except Exception as exc:
        print(f"[RTSP-RELAY] listar cameras falhou: {exc}", flush=True)
        return []
    dados = data.get("dados") if isinstance(data, dict) else None
    if not isinstance(dados, list):
        return []
    return [c for c in dados if isinstance(c, dict)]


@dataclass
class RelayProc:
    camera_id: int
    path: str
    process: subprocess.Popen
    started_at: float = field(default_factory=time.time)
    last_demand_at: float = field(default_factory=time.time)


class RtspLiveRelay:
    def __init__(self) -> None:
        self._lock = threading.Lock()
        self._running: dict[int, RelayProc] = {}
        self._ffmpeg = _env("FFMPEG_PATH", "ffmpeg")

    def _log_ffmpeg_exit(self, rp: RelayProc) -> None:
        code = rp.process.poll()
        if code is None:
            return
        err_raw = b""
        if rp.process.stderr:
            try:
                err_raw = rp.process.stderr.read() or b""
            except OSError:
                pass
        msg = _sanitize_ffmpeg_stderr(err_raw) if err_raw else "(sem stderr)"
        print(
            f"[RTSP-RELAY] ffmpeg exit camera={rp.camera_id} path={rp.path} "
            f"code={code} err={msg}",
            flush=True,
        )

    def _should_run(self, cam: dict[str, Any], now: float) -> bool:
        if _wake_active(cam, now):
            return True
        path = _path_name(cam)
        if path_reader_count(path) > 0:
            return True
        with self._lock:
            proc = self._running.get(int(cam.get("id") or 0))
        if proc and (now - proc.last_demand_at) < RTSP_LIVE_RELAY_IDLE_SEC:
            return True
        return False

    def _start(self, cam: dict[str, Any]) -> None:
        cid = int(cam.get("id") or 0)
        rtsp = str(cam.get("rtsp_url_sec") or "").strip()
        dest = _rtmp_dest(cam)
        path = _path_name(cam)
        if cid < 1 or not rtsp.lower().startswith("rtsp") or not dest:
            if cid >= 1:
                if not rtsp.lower().startswith("rtsp"):
                    print(f"[RTSP-RELAY] skip camera={cid} rtsp_url_sec invalido", flush=True)
                elif not dest:
                    print(
                        f"[RTSP-RELAY] skip camera={cid} destino RTMP vazio "
                        f"(RTMP_PUBLISH_SECRET / mediamtx_rtmp_public)",
                        flush=True,
                    )
            return
        args = [
            self._ffmpeg,
            "-hide_banner",
            "-loglevel",
            "error",
            "-rtsp_transport",
            "tcp",
            "-i",
            rtsp,
            "-c:v",
            "copy",
            "-an",
            "-f",
            "flv",
            dest,
        ]
        try:
            proc = subprocess.Popen(
                args,
                stdout=subprocess.DEVNULL,
                stderr=subprocess.PIPE,
            )
        except OSError as exc:
            print(f"[RTSP-RELAY] start camera={cid} falhou: {exc}", flush=True)
            return
        now = time.time()
        with self._lock:
            old = self._running.pop(cid, None)
            if old:
                if old.process.poll() is not None:
                    self._log_ffmpeg_exit(old)
                elif old.process.poll() is None:
                    old.process.terminate()
            self._running[cid] = RelayProc(
                camera_id=cid, path=path, process=proc, started_at=now, last_demand_at=now
            )
        print(f"[RTSP-RELAY] start camera={cid} path={path}", flush=True)

    def _stop(self, camera_id: int, reason: str = "idle") -> None:
        with self._lock:
            rp = self._running.pop(camera_id, None)
        if not rp:
            return
        if rp.process.poll() is None:
            rp.process.terminate()
            try:
                rp.process.wait(timeout=8)
            except subprocess.TimeoutExpired:
                rp.process.kill()
        if rp.process.poll() not in (None, 0):
            self._log_ffmpeg_exit(rp)
        print(f"[RTSP-RELAY] stop camera={camera_id} path={rp.path} reason={reason}", flush=True)

    def tick(self) -> None:
        if not RTSP_LIVE_RELAY_ENABLED:
            return
        cameras = fetch_rtsp_pull_cameras()
        now = time.time()
        want: set[int] = set()
        for cam in cameras:
            cid = int(cam.get("id") or 0)
            if cid < 1:
                continue
            if self._should_run(cam, now):
                want.add(cid)
                with self._lock:
                    running = self._running.get(cid)
                if running is not None and running.process.poll() is not None:
                    self._log_ffmpeg_exit(running)
                    with self._lock:
                        self._running.pop(cid, None)
                    running = None
                if running is None:
                    self._start(cam)
                else:
                    with self._lock:
                        if cid in self._running:
                            self._running[cid].last_demand_at = now
            elif cid in self._running:
                with self._lock:
                    rp = self._running.get(cid)
                if rp and (now - rp.last_demand_at) >= RTSP_LIVE_RELAY_IDLE_SEC:
                    self._stop(cid, "idle")

        active_ids = {int(c.get("id") or 0) for c in cameras if int(c.get("id") or 0) > 0}
        with self._lock:
            stale = [cid for cid in self._running if cid not in want]
            removed = [cid for cid in self._running if cid not in active_ids]
        for cid in removed:
            self._stop(cid, "removed")
        for cid in stale:
            with self._lock:
                rp = self._running.get(cid)
            if rp and (now - rp.last_demand_at) >= RTSP_LIVE_RELAY_IDLE_SEC:
                self._stop(cid, "idle")

    def run_forever(self) -> None:
        if not RTSP_LIVE_RELAY_ENABLED:
            print("[RTSP-RELAY] desabilitado (RTSP_LIVE_RELAY_ENABLED=0)", flush=True)
            while True:
                time.sleep(60)
        print(
            f"[RTSP-RELAY] ativo worker_id={WORKER_ID or '?'} poll={RTSP_LIVE_RELAY_POLL_SEC}s "
            f"idle={RTSP_LIVE_RELAY_IDLE_SEC}s",
            flush=True,
        )
        while True:
            try:
                self.tick()
            except Exception as exc:
                print(f"[RTSP-RELAY] tick erro: {exc}", flush=True)
            time.sleep(max(2, RTSP_LIVE_RELAY_POLL_SEC))


def main() -> None:
    RtspLiveRelay().run_forever()
