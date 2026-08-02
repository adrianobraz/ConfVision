"""Lista de IPs banidos (auto + manual) persistida em JSON."""

from __future__ import annotations

import json
import threading
import time
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Optional

# Falhas de config/rede do cliente — ban mais tolerante (ou desligado).
MOTIVOS_SOFT = frozenset({"eof_sem_publish", "path_barra_final", "closed_outro"})


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


@dataclass(frozen=True)
class _Limites:
    max_fails: int
    window_sec: int
    ban_ttl_sec: int


class BanStore:
    def __init__(
        self,
        path: str,
        *,
        max_fails: int = 20,
        window_sec: int = 60,
        ban_ttl_sec: int = 3600,
        soft_max_fails: int = 0,
        soft_window_sec: int = 300,
        soft_ban_ttl_sec: int = 600,
        auto_unban_on_publish: bool = True,
    ):
        self.path = Path(path)
        self.hard = _Limites(
            max_fails=max(1, max_fails),
            window_sec=max(10, window_sec),
            ban_ttl_sec=max(60, ban_ttl_sec),
        )
        self.soft = _Limites(
            max_fails=max(0, soft_max_fails),
            window_sec=max(10, soft_window_sec),
            ban_ttl_sec=max(60, soft_ban_ttl_sec),
        )
        self.auto_unban_on_publish = auto_unban_on_publish
        self._lock = threading.Lock()
        self._bans: dict[str, BanEntry] = {}
        # ip -> motivo -> timestamps
        self._fails: dict[str, dict[str, list[float]]] = {}
        self._load()

    def _limites(self, motivo: str) -> _Limites:
        if (motivo or "").strip() in MOTIVOS_SOFT:
            return self.soft
        return self.hard

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
                    expira_em=exp or (now + self.hard.ban_ttl_sec),
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
        for ip, por_motivo in list(self._fails.items()):
            kept_m: dict[str, list[float]] = {}
            for motivo, times in por_motivo.items():
                lim = self._limites(motivo)
                kept = [t for t in times if now - t <= lim.window_sec]
                if kept:
                    kept_m[motivo] = kept
            if kept_m:
                self._fails[ip] = kept_m
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
        ttl = self.hard.ban_ttl_sec if ttl_sec is None else max(60, int(ttl_sec))
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

    def registrar_sucesso(self, ip: str) -> bool:
        """Limpa falhas; remove auto-ban (mantém ban manual)."""
        ip = (ip or "").strip()
        if not ip or not self.auto_unban_on_publish:
            return False
        with self._lock:
            self._purge_locked()
            self._fails.pop(ip, None)
            entry = self._bans.get(ip)
            if entry and entry.manual:
                return False
            if entry:
                del self._bans[ip]
                self._save()
                print(f"[RTMP-BAN] AUTO-UNBAN ip={ip} motivo=publish_ok", flush=True)
                return True
        return False

    def registrar_falha(self, ip: str, motivo: str = "auth_falhou") -> Optional[BanEntry]:
        """Conta falha por motivo; se estourar limiar, bane e retorna o ban."""
        ip = (ip or "").strip()
        motivo = (motivo or "auth_falhou").strip() or "auth_falhou"
        if not ip or ip in ("127.0.0.1", "::1"):
            return None
        lim = self._limites(motivo)
        if lim.max_fails < 1:
            return None
        now = time.time()
        with self._lock:
            self._purge_locked()
            if ip in self._bans:
                return self._bans[ip]
            por_motivo = self._fails.setdefault(ip, {})
            times = [t for t in por_motivo.get(motivo, []) if now - t <= lim.window_sec]
            times.append(now)
            por_motivo[motivo] = times
            if len(times) < lim.max_fails:
                print(
                    f"[RTMP-BAN] falha ip={ip} tentativa={len(times)}/{lim.max_fails} "
                    f"motivo={motivo} ({'soft' if motivo in MOTIVOS_SOFT else 'hard'})",
                    flush=True,
                )
                return None
            entry = BanEntry(
                ip=ip,
                motivo=f"auto:{motivo} ({len(times)} falhas/{lim.window_sec}s)",
                criado_em=now,
                expira_em=now + lim.ban_ttl_sec,
                falhas=len(times),
                manual=False,
            )
            self._bans[ip] = entry
            por_motivo.pop(motivo, None)
            if not por_motivo:
                self._fails.pop(ip, None)
            self._save()
        print(
            f"[RTMP-BAN] AUTO-BAN ip={ip} falhas={entry.falhas}/{lim.max_fails} motivo={motivo} "
            f"ttl={lim.ban_ttl_sec}s — próximas tentativas bloqueadas",
            flush=True,
        )
        return entry
