"""Auth HTTP do MediaMTX + regras ConfVision (path cam/{hash})."""

from __future__ import annotations

import os
import time
from typing import Any, Optional

import requests

from rtmp_ban import BanStore
from vis_api_auth import vis_api_headers
from rtmp_token import RE_HASH_PATH, chave_valida, parse_chave_rtmp, publish_secret


def _env(name: str, default: str = "") -> str:
    return (os.getenv(name) or default).strip()


def _truthy(v: Any) -> bool:
    if v is True or v == 1:
        return True
    if isinstance(v, str) and v.strip().lower() in ("1", "true", "yes", "sim"):
        return True
    return False


def _meta(path: str = "", hash_: str = "", camera_id: Any = None, plano: str = "") -> dict[str, Any]:
    return {
        "path": path or "",
        "hash": hash_ or "",
        "camera_id": camera_id if camera_id is not None else "",
        "plano": plano or "",
    }


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
        # 0 = read/playback passam pela mesma regra de publish (bloqueado/ativo/plano)
        self.allow_read_open = _env("RTMP_ALLOW_READ_OPEN", "0") in ("1", "true", "yes")

    def authorize(self, payload: dict[str, Any]) -> tuple[int, str, dict[str, Any]]:
        """Retorna (http_status, motivo, meta). 200 = ok; 401/403 = nega."""
        action = str(payload.get("action") or "").strip().lower()
        ip = str(payload.get("ip") or "").strip()
        user = str(payload.get("user") or "").strip()
        password = str(payload.get("password") or "").strip()
        path = str(payload.get("path") or "").strip().lstrip("/").rstrip("/")
        base_meta = _meta(path=path)

        if ip and self.bans.is_banned(ip):
            return 403, "ip_banido", base_meta

        if action in ("api", "metrics", "pprof"):
            if self.dvr_user and user == self.dvr_user and password == self.dvr_pass:
                return 200, "api_ok", base_meta
            return 401, "api_credencial_invalida", base_meta

        if action in ("read", "playback"):
            if self.allow_read_open:
                return 200, "read_aberto", base_meta
            return self._authorize_camera_path(ip, path, ok_motivo="read_ok")

        if action == "publish":
            return self._authorize_camera_path(ip, path, ok_motivo="publish_ok")

        return 401, f"action_desconhecida:{action}", base_meta

    def _authorize_camera_path(
        self, ip: str, path: str, *, ok_motivo: str = "publish_ok"
    ) -> tuple[int, str, dict[str, Any]]:
        if not self.secret:
            print("[RTMP-GUARD] RTMP_PUBLISH_SECRET vazio — negando acesso", flush=True)
            return 403, "secret_nao_configurado", _meta(path=path)

        m = RE_HASH_PATH.match(path)
        if not m:
            # tenta extrair hash se veio só o token ou path legado
            nome = path.rsplit("/", 1)[-1] if path else ""
            self._fail(ip, "path_invalido")
            return 403, "path_invalido", _meta(path=path, hash_=nome)

        chave = m.group(1)
        meta = _meta(path=path, hash_=chave)

        if not chave_valida(chave, secret=self.secret):
            self._fail(ip, "chave_invalida")
            return 403, "chave_invalida", meta

        camera_id = parse_chave_rtmp(chave, secret=self.secret)
        assert camera_id is not None
        meta["camera_id"] = camera_id

        cam = self._fetch_camera(camera_id)
        if not cam:
            self._fail(ip, "camera_nao_encontrada")
            return 403, "camera_nao_encontrada", meta

        plano = str(cam.get("plano") or "").strip().lower()
        meta["plano"] = plano

        if _truthy(cam.get("bloqueado")):
            self._fail(ip, "camera_bloqueada")
            return 403, "camera_bloqueada", meta

        # Plano online grava ativo=false de propósito (sob demanda).
        # OK se: bloqueado=false E (ativo=true OU plano=online).
        if plano == "online" or _truthy(cam.get("ativo")):
            return 200, ok_motivo, meta

        self._fail(ip, "camera_inativa")
        return 403, "camera_inativa", meta

    def _fail(self, ip: str, motivo: str) -> None:
        if ip:
            self.bans.registrar_falha(ip, motivo=motivo)

    def _fetch_camera(self, camera_id: int) -> Optional[dict]:
        cached = self.cache.get(camera_id)
        if cached is not None:
            return cached
        try:
            from config_cache import read_rtmp_auth, write_rtmp_auth

            redis_cam = read_rtmp_auth(camera_id)
            if redis_cam is not None:
                self.cache.set(camera_id, redis_cam)
                return redis_cam
        except Exception:
            pass
        if not self.xano:
            print("[RTMP-GUARD] XANO_BASE_URL vazio", flush=True)
            return None
        url = f"{self.xano}/vis_camera/rtmp_auth/{camera_id}"
        try:
            r = requests.get(
                url,
                params={"vis_camera_id": camera_id},
                headers=vis_api_headers(),
                timeout=8,
            )
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
            try:
                from config_cache import write_rtmp_auth

                write_rtmp_auth(camera_id, cam, self.cache.ttl)
            except Exception:
                pass
            return cam
        except Exception as exc:
            print(f"[RTMP-GUARD] xano falhou camera={camera_id}: {exc}", flush=True)
            return None
