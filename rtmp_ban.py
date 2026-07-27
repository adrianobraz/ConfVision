"""Lista de IPs banidos (auto + manual) persistida em JSON."""

from __future__ import annotations

import json
import threading
import time
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Optional


@dataclass
class BanEntry:
    ip: str
    motivo: str
    criado_em: float
    expira_em: float
    falhas: int = 0
    manual: bool = False

    def to_dict(self) -> dict:
        d = asdict(self)
        d["expirado"] = time.time() >= self.expira_em
        d["restante_sec"] = max(0, int(self.expira_em - time.time()))
        return d


class BanStore:
    def __init__(
        self,
        path: str,
        *,
        max_fails: int = 20,
        window_sec: int = 60,
        ban_ttl_sec: int = 3600,
    ):
        self.path = Path(path)
        self.max_fails = max(3, max_fails)
        self.window_sec = max(10, window_sec)
        self.ban_ttl_sec = max(60, ban_ttl_sec)
        self._lock = threading.Lock()
        self._bans: dict[str, BanEntry] = {}
        self._fails: dict[str, list[float]] = {}
        self._load()

    def _load(self) -> None:
        if not self.path.exists():
            return
        try:
            raw = json.loads(self.path.read_text(encoding="utf-8"))
            items = raw.get("bans") if isinstance(raw, dict) else raw
            if not isinstance(items, list):
                return
            now = time.time()
            for it in items:
                ip = str(it.get("ip") or "").strip()
                if not ip:
                    continue
                exp = float(it.get("expira_em") or 0)
                if exp and exp < now:
                    continue
                self._bans[ip] = BanEntry(
                    ip=ip,
                    motivo=str(it.get("motivo") or "ban"),
                    criado_em=float(it.get("criado_em") or now),
                    expira_em=exp or (now + self.ban_ttl_sec),
                    falhas=int(it.get("falhas") or 0),
                    manual=bool(it.get("manual")),
                )
        except Exception as exc:
            print(f"[RTMP-BAN] falha ao ler {self.path}: {exc}", flush=True)

    def _save(self) -> None:
        self.path.parent.mkdir(parents=True, exist_ok=True)
        payload = {"bans": [b.to_dict() for b in self._bans.values()]}
        tmp = self.path.with_suffix(".tmp")
        tmp.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")
        tmp.replace(self.path)

    def _purge_locked(self) -> None:
        now = time.time()
        expired = [ip for ip, b in self._bans.items() if b.expira_em <= now]
        for ip in expired:
            del self._bans[ip]
        # limpa falhas antigas
        for ip, times in list(self._fails.items()):
            kept = [t for t in times if now - t <= self.window_sec]
            if kept:
                self._fails[ip] = kept
            else:
                del self._fails[ip]

    def is_banned(self, ip: str) -> bool:
        ip = (ip or "").strip()
        if not ip:
            return False
        with self._lock:
            self._purge_locked()
            return ip in self._bans

    def get(self, ip: str) -> Optional[BanEntry]:
        with self._lock:
            self._purge_locked()
            return self._bans.get(ip)

    def listar(self) -> list[dict]:
        with self._lock:
            self._purge_locked()
            return [b.to_dict() for b in sorted(self._bans.values(), key=lambda x: -x.criado_em)]

    def ban(
        self,
        ip: str,
        *,
        motivo: str,
        falhas: int = 0,
        manual: bool = False,
        ttl_sec: Optional[int] = None,
    ) -> BanEntry:
        ip = (ip or "").strip()
        now = time.time()
        ttl = self.ban_ttl_sec if ttl_sec is None else max(60, int(ttl_sec))
        entry = BanEntry(
            ip=ip,
            motivo=motivo,
            criado_em=now,
            expira_em=now + ttl,
            falhas=falhas,
            manual=manual,
        )
        with self._lock:
            self._bans[ip] = entry
            self._fails.pop(ip, None)
            self._save()
        print(f"[RTMP-BAN] BAN ip={ip} motivo={motivo} ttl={ttl}s", flush=True)
        return entry

    def unban(self, ip: str) -> bool:
        ip = (ip or "").strip()
        with self._lock:
            self._purge_locked()
            if ip not in self._bans:
                return False
            del self._bans[ip]
            self._fails.pop(ip, None)
            self._save()
        print(f"[RTMP-BAN] UNBAN ip={ip}", flush=True)
        return True

    def registrar_falha(self, ip: str, motivo: str = "auth_falhou") -> Optional[BanEntry]:
        """Conta falha; se estourar limiar, bane e retorna o ban."""
        ip = (ip or "").strip()
        if not ip or ip in ("127.0.0.1", "::1"):
            return None
        now = time.time()
        with self._lock:
            self._purge_locked()
            if ip in self._bans:
                return self._bans[ip]
            times = [t for t in self._fails.get(ip, []) if now - t <= self.window_sec]
            times.append(now)
            self._fails[ip] = times
            if len(times) < self.max_fails:
                return None
            entry = BanEntry(
                ip=ip,
                motivo=f"auto:{motivo} ({len(times)} falhas/{self.window_sec}s)",
                criado_em=now,
                expira_em=now + self.ban_ttl_sec,
                falhas=len(times),
                manual=False,
            )
            self._bans[ip] = entry
            self._fails.pop(ip, None)
            self._save()
        print(
            f"[RTMP-BAN] AUTO-BAN ip={ip} falhas={entry.falhas} motivo={motivo}",
            flush=True,
        )
        return entry
