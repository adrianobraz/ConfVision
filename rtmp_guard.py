"""Auth HTTP do MediaMTX + regras ConfVision (chave 24 dígitos no path)."""

from __future__ import annotations

import os
import re
import time
from typing import Any, Optional

import requests

from rtmp_ban import BanStore
from rtmp_token import chave_valida, parse_chave_rtmp, publish_secret

RE_LIVE_CHAVE = re.compile(r"^live/(\d{24})/?$")


def _env(name: str, default: str = "") -> str:
    return (os.getenv(name) or default).strip()


class CameraCache:
    def __init__(self, ttl_sec: int = 45):
        self.ttl = max(5, ttl_sec)
        self._data: dict[int, tuple[float, dict]] = {}

    def get(self, camera_id: int) -> Optional[dict]:
        hit = self._data.get(camera_id)
        if not hit:
            return None
        ts, cam = hit
        if time.time() - ts > self.ttl:
            self._data.pop(camera_id, None)
            return None
        return cam

    def set(self, camera_id: int, cam: dict) -> None:
        self._data[camera_id] = (time.time(), cam)

    def invalidate(self, camera_id: int) -> None:
        self._data.pop(camera_id, None)


class RtmpGuard:
    def __init__(self, bans: BanStore):
        self.bans = bans
        self.xano = _env("XANO_BASE_URL").rstrip("/")
        self.secret = publish_secret()
        self.dvr_user = _env("MEDIAMTX_API_USER", "dvr")
        self.dvr_pass = _env("MEDIAMTX_API_PASS")
        self.cache = CameraCache(ttl_sec=int(_env("RTMP_AUTH_CACHE_SEC", "45") or "45"))
        self.allow_read_open = _env("RTMP_ALLOW_READ_OPEN", "1") in ("1", "true", "yes")

    def authorize(self, payload: dict[str, Any]) -> tuple[int, str]:
        """Retorna (http_status, motivo). 200 = ok; 401/403 = nega."""
        action = str(payload.get("action") or "").strip().lower()
        ip = str(payload.get("ip") or "").strip()
        user = str(payload.get("user") or "").strip()
        password = str(payload.get("password") or "").strip()
        path = str(payload.get("path") or "").strip().lstrip("/")

        if ip and self.bans.is_banned(ip):
            return 403, "ip_banido"

        if action in ("api", "metrics", "pprof"):
            if self.dvr_user and user == self.dvr_user and password == self.dvr_pass:
                return 200, "api_ok"
            return 401, "api_credencial_invalida"

        if action in ("read", "playback"):
            if self.allow_read_open:
                return 200, "read_aberto"
            return self._authorize_publish(ip, path)

        if action == "publish":
            return self._authorize_publish(ip, path)

        return 401, f"action_desconhecida:{action}"

    def _authorize_publish(self, ip: str, path: str) -> tuple[int, str]:
        if not self.secret:
            print("[RTMP-GUARD] RTMP_PUBLISH_SECRET vazio — negando publish", flush=True)
            return 403, "secret_nao_configurado"

        m = RE_LIVE_CHAVE.match(path)
        if not m:
            self._fail(ip, "path_invalido")
            return 403, "path_invalido"

        chave = m.group(1)
        if not chave_valida(chave, secret=self.secret):
            self._fail(ip, "chave_invalida")
            return 403, "chave_invalida"

        parsed = parse_chave_rtmp(chave)
        assert parsed is not None
        _mac, fra6, camera_id = parsed

        cam = self._fetch_camera(camera_id)
        if not cam:
            self._fail(ip, "camera_nao_encontrada")
            return 403, "camera_nao_encontrada"

        id_fra = str(cam.get("id_franqueado") or "").strip()
        digits = re.sub(r"\D", "", id_fra)
        if not digits.endswith(fra6):
            self._fail(ip, "franqueado_divergente")
            return 403, "franqueado_divergente"

        if cam.get("bloqueado") is True:
            self._fail(ip, "camera_bloqueada")
            return 403, "camera_bloqueada"

        # ativo=false NÃO bloqueia publish (plano "online" força ativo=false).
        return 200, "publish_ok"

    def _fail(self, ip: str, motivo: str) -> None:
        if ip:
            self.bans.registrar_falha(ip, motivo=motivo)

    def _fetch_camera(self, camera_id: int) -> Optional[dict]:
        cached = self.cache.get(camera_id)
        if cached is not None:
            return cached
        if not self.xano:
            print("[RTMP-GUARD] XANO_BASE_URL vazio", flush=True)
            return None
        url = f"{self.xano}/vis_camera/rtmp_auth/{camera_id}"
        try:
            r = requests.get(url, params={"vis_camera_id": camera_id}, timeout=8)
            if r.status_code == 404:
                return None
            r.raise_for_status()
            data = r.json()
            if not isinstance(data, dict):
                return None
            cam = data.get("dados") if isinstance(data.get("dados"), dict) else data
            if not cam or cam.get("id") is None:
                return None
            self.cache.set(camera_id, cam)
            return cam
        except Exception as exc:
            print(f"[RTMP-GUARD] xano falhou camera={camera_id}: {exc}", flush=True)
            return None
