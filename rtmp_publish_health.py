"""Último publish OK por path/câmera — persistência leve para monitoramento."""

from __future__ import annotations

import json
import threading
import time
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Optional

from rtmp_token import parse_chave_rtmp


def _iso(ts: float) -> str:
    return datetime.fromtimestamp(ts, tz=timezone.utc).isoformat()


class PublishHealthStore:
    def __init__(self, path: str = ""):
        self.path = Path(path) if path else None
        self._lock = threading.Lock()
        # path -> {ip, camera_id, ultimo_publish, ultimo_publish_iso}
        self._items: dict[str, dict[str, Any]] = {}
        self._load()

    def _load(self) -> None:
        if not self.path or not self.path.exists():
            return
        try:
            raw = json.loads(self.path.read_text(encoding="utf-8"))
            items = raw.get("publishers") if isinstance(raw, dict) else raw
            if not isinstance(items, dict):
                return
            self._items = {str(k): v for k, v in items.items() if isinstance(v, dict)}
        except Exception as exc:
            print(f"[RTMP-HEALTH] falha ao ler {self.path}: {exc}", flush=True)

    def _save(self) -> None:
        if not self.path:
            return
        self.path.parent.mkdir(parents=True, exist_ok=True)
        with self._lock:
            payload = {
                "atualizado_em": _iso(time.time()),
                "publishers": dict(self._items),
            }
        tmp = self.path.with_suffix(".tmp")
        tmp.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")
        tmp.replace(self.path)

    def registrar_publish(
        self,
        path: str,
        *,
        ip: str = "",
        camera_id: Any = None,
        fonte: str = "log",
    ) -> None:
        p = (path or "").strip().rstrip("/")
        if not p:
            return
        now = time.time()
        cid = camera_id
        if cid is None:
            cid = parse_chave_rtmp(p)
        with self._lock:
            self._items[p] = {
                "path": p,
                "ip": (ip or "").strip(),
                "camera_id": cid,
                "ultimo_publish": now,
                "ultimo_publish_iso": _iso(now),
                "fonte": fonte,
            }
        self._save()

    def remover_path(self, path: str) -> None:
        p = (path or "").strip().rstrip("/")
        if not p:
            return
        with self._lock:
            self._items.pop(p, None)
        self._save()

    def listar(
        self,
        *,
        online_paths: Optional[set[str]] = None,
        offline_apos_sec: int = 300,
    ) -> list[dict[str, Any]]:
        now = time.time()
        online_paths = online_paths or set()
        with self._lock:
            rows = []
            for p, item in self._items.items():
                row = dict(item)
                row["online_agora"] = p in online_paths
                ultimo = float(row.get("ultimo_publish") or 0)
                row["segundos_desde_publish"] = max(0, int(now - ultimo)) if ultimo else None
                if row["online_agora"]:
                    row["status"] = "online"
                elif ultimo and (now - ultimo) <= offline_apos_sec:
                    row["status"] = "recente_offline"
                else:
                    row["status"] = "offline"
                rows.append(row)
            rows.sort(key=lambda x: (x.get("camera_id") is None, x.get("camera_id") or 0, x.get("path") or ""))
            return rows

    def resumo(self, *, online_paths: Optional[set[str]] = None) -> dict[str, Any]:
        online_paths = online_paths or set()
        rows = self.listar(online_paths=online_paths)
        return {
            "total_rastreados": len(rows),
            "online_agora": sum(1 for r in rows if r.get("online_agora")),
            "offline": sum(1 for r in rows if r.get("status") == "offline"),
            "recente_offline": sum(1 for r in rows if r.get("status") == "recente_offline"),
        }
