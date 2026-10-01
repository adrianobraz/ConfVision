"""Auth HTTP do MediaMTX + regras ConfVision (path cam/{hash})."""

from __future__ import annotations

import os
import time
from typing import Any, Optional

import requests

from rtmp_ban import BanStore
from rtmp_token import RE_HASH_PATH, chave_valida, parse_chave_rtmp, publish_secret
from stream_health_client import confvision_api_base, rtmp_auth_request_headers


def _env(name: str, default: str = "") -> str:
    return (os.getenv(name) or default).strip()


def _truthy(v: Any) -> bool:
    if v is True or v == 1:
        return True
    if isinstance(v, str) and v.strip().lower() in ("1", "true", "yes", "sim", "s"):
        return True
    return False


def _stream_pausado_pelo_sistema(cam: dict[str, Any]) -> bool:
    motivo = str(cam.get("stream_motivo_pausa") or "").strip()
    return motivo.startswith("sistema_stream")


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


# Motivo Postgres / Go — prefixo sistema_stream obrigatório.
STREAM_PAUSE_VIDEO_TRACK = "sistema_stream_video_track_not_set_up"
STREAM_ERRO_VIDEO_TRACK = "VIDEO_TRACK_NOT_SET_UP"


def _cam_has_stream_policy_fields(cam: dict[str, Any]) -> bool:
    return "stream_motivo_pausa" in cam or "analitico_pausado" in cam


class RtmpGuard:
    def __init__(self, bans: BanStore):
        self.bans = bans
        self.xano = _env("XANO_BASE_URL").rstrip("/")
        self.auth_bases = self._rtmp_auth_bases()
        self.secret = publish_secret()
        self.dvr_user = _env("MEDIAMTX_API_USER", "dvr")
        self.dvr_pass = _env("MEDIAMTX_API_PASS")
        self.cache = CameraCache(ttl_sec=int(_env("RTMP_AUTH_CACHE_SEC", "10") or "10"))
        # 0 = read/playback passam pela mesma regra de publish (bloqueado/ativo/plano)
        self.allow_read_open = _env("RTMP_ALLOW_READ_OPEN", "0") in ("1", "true", "yes")

    def authorize(self, payload: dict[str, Any]) -> tuple[int, str, dict[str, Any]]:
        """Retorna (http_status, motivo, meta). 200 = ok; 401/403 = nega."""
        action = str(payload.get("action") or "").strip().lower()
        ip = str(payload.get("ip") or "").strip()
        user = str(payload.get("user") or "").strip()
        password = str(payload.get("password") or "").strip()
        path = str(payload.get("path") or "").strip().lstrip("/")
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

        cam, fetch_err = self._fetch_camera(camera_id)
        if not cam:
            if fetch_err == "rtmp_auth_nao_autorizado":
                print(
                    "[RTMP-GUARD] ERRO rtmp_auth 401 — confira VIS_WORKER_API_KEY no foxpro "
                    "e no core-4 (sem contar falha de IP)",
                    flush=True,
                )
                return 503, "rtmp_auth_nao_autorizado", meta
            self._fail(ip, "camera_nao_encontrada")
            return 403, "camera_nao_encontrada", meta

        plano = str(cam.get("plano") or "").strip().lower()
        meta["plano"] = plano

        if _truthy(cam.get("bloqueado")):
            return 403, "camera_bloqueada", meta

        if _stream_pausado_pelo_sistema(cam):
            meta["stream_motivo_pausa"] = str(
                cam.get("stream_motivo_pausa") or STREAM_PAUSE_VIDEO_TRACK
            ).strip()
            return 403, "stream_pausado_sistema", meta

        # Plano online grava ativo=false de propósito (sob demanda).
        # OK se: bloqueado=false E (ativo=true OU plano=online).
        if plano == "online" or _truthy(cam.get("ativo")):
            return 200, ok_motivo, meta

        return 403, "camera_inativa", meta

    def _fail(self, ip: str, motivo: str) -> None:
        if ip:
            self.bans.registrar_falha(ip, motivo=motivo)

    @staticmethod
    def _rtmp_auth_bases() -> list[str]:
        """Go (vis_*) primeiro — endpoint legado Xano não traz stream_motivo_pausa."""
        out: list[str] = []
        go = confvision_api_base()
        if go:
            out.append(go)
        xano = _env("XANO_BASE_URL").rstrip("/")
        if xano and xano not in out:
            out.append(xano)
        return out

    @staticmethod
    def _parse_rtmp_auth_payload(data: dict) -> Optional[dict]:
        if not isinstance(data, dict):
            return None
        cam = data.get("dados") if isinstance(data.get("dados"), dict) else data
        if not cam or cam.get("id") is None:
            return None
        return cam

    def _fetch_camera(self, camera_id: int) -> tuple[Optional[dict], Optional[str]]:
        cached = self.cache.get(camera_id)
        if cached is not None:
            return cached, None
        bases = self.auth_bases or self._rtmp_auth_bases()
        if not bases:
            print(
                "[RTMP-GUARD] ERRO rtmp_auth: CONFVISION_API_URL e XANO_BASE_URL vazios",
                flush=True,
            )
            return None, "camera_nao_encontrada"
        last_exc: Optional[Exception] = None
        saw_auth_error = False
        for base in bases:
            url = f"{base.rstrip('/')}/vis_camera/rtmp_auth/{camera_id}"
            headers = rtmp_auth_request_headers(base)
            if confvision_api_base() and base.rstrip("/") == confvision_api_base() and not headers:
                print(
                    "[RTMP-GUARD] AVISO rtmp_auth Go sem VIS_WORKER_API_KEY — "
                    f"camera_id={camera_id}",
                    flush=True,
                )
            try:
                r = requests.get(
                    url,
                    params={"vis_camera_id": camera_id},
                    headers=headers or None,
                    timeout=8,
                )
                if r.status_code == 401:
                    saw_auth_error = True
                    print(
                        f"[RTMP-GUARD] rtmp_auth 401 base={base} camera={camera_id} "
                        "(VIS_WORKER_API_KEY ausente ou inválida)",
                        flush=True,
                    )
                    continue
                if r.status_code == 404:
                    continue
                r.raise_for_status()
                cam = self._parse_rtmp_auth_payload(r.json())
                if not cam:
                    continue
                if not _cam_has_stream_policy_fields(cam):
                    print(
                        "[RTMP-GUARD] AVISO rtmp_auth sem campos stream_policy "
                        f"(endpoint legado?) base={base} camera_id={camera_id} — tentando proxima base",
                        flush=True,
                    )
                    continue
                self.cache.set(camera_id, cam)
                return cam, None
            except Exception as exc:
                last_exc = exc
                print(
                    f"[RTMP-GUARD] rtmp_auth falhou base={base} camera={camera_id}: {exc}",
                    flush=True,
                )
        if saw_auth_error and not last_exc:
            return None, "rtmp_auth_nao_autorizado"
        if last_exc:
            print(
                f"[RTMP-GUARD] ERRO rtmp_auth esgotou bases camera_id={camera_id}: {last_exc}",
                flush=True,
            )
        return None, "camera_nao_encontrada"
