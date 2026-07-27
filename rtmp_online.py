"""Lista streams RTMP online via API MediaMTX + parse da chave 24."""

from __future__ import annotations

import json
import os
import urllib.error
import urllib.request
from base64 import b64encode
from typing import Any, Optional
from urllib.parse import quote

from rtmp_token import parse_chave_rtmp
from rtmp_watch import path_label


def _env(name: str, default: str = "") -> str:
    return (os.getenv(name) or default).strip()


def mtx_api_base() -> str:
    return _env("MEDIAMTX_API_BASE", "http://foxpro_confvision:9997").rstrip("/")


def _api_get(path: str, timeout: float = 5.0) -> Optional[dict]:
    base = mtx_api_base()
    if not base:
        return None
    url = f"{base}{path}"
    req = urllib.request.Request(url, method="GET")
    user = _env("MEDIAMTX_API_USER")
    password = _env("MEDIAMTX_API_PASS")
    if user:
        token = b64encode(f"{user}:{password}".encode("utf-8")).decode("ascii")
        req.add_header("Authorization", f"Basic {token}")
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            raw = resp.read().decode("utf-8")
            data = json.loads(raw or "{}")
            return data if isinstance(data, dict) else None
    except (urllib.error.URLError, urllib.error.HTTPError, TimeoutError, json.JSONDecodeError, OSError) as exc:
        print(f"[RTMP-ONLINE] API {path} falhou: {exc}", flush=True)
        return None


def _path_online(item: dict) -> bool:
    if item.get("online") is True or item.get("available") is True or item.get("ready") is True:
        return True
    return False


def _ip_from_remote(remote: str) -> str:
    s = (remote or "").strip()
    if not s:
        return ""
    # IPv4:port ou [IPv6]:port
    if s.startswith("["):
        end = s.find("]")
        return s[1:end] if end > 1 else s
    if ":" in s:
        return s.rsplit(":", 1)[0]
    return s


def _decode_path(path: str) -> dict[str, Any]:
    p = (path or "").strip().rstrip("/")
    nome = p.rsplit("/", 1)[-1] if p else ""
    out: dict[str, Any] = {
        "path": p,
        "path_label": path_label(p),
        "camera_id": None,
        "id_franqueado_sufixo": "",
        "legado": False,
    }
    parsed = parse_chave_rtmp(nome)
    if parsed:
        _, fra6, cam = parsed
        out["camera_id"] = cam
        out["id_franqueado_sufixo"] = fra6
        return out
    if nome.isdigit():
        out["camera_id"] = int(nome)
        out["legado"] = True
    return out


def listar_online(fallback_publishers: Optional[dict[str, dict]] = None) -> dict[str, Any]:
    """Retorna {status, dados: [...], fonte, aviso?}."""
    paths_data = _api_get("/v3/paths/list?itemsPerPage=500")
    conns_data = _api_get("/v3/rtmpconns/list?itemsPerPage=500")

    publishers_by_path: dict[str, str] = {}
    if conns_data:
        for c in conns_data.get("items") or []:
            if not isinstance(c, dict):
                continue
            state = str(c.get("state") or "").lower()
            # publish | publishIdle etc.
            if "publish" not in state:
                continue
            path = str(c.get("path") or "").strip().rstrip("/")
            ip = _ip_from_remote(str(c.get("remoteAddr") or ""))
            if path and ip:
                publishers_by_path[path] = ip

    items: list[dict[str, Any]] = []
    fonte = "mediamtx-api"

    if paths_data:
        for item in paths_data.get("items") or []:
            if not isinstance(item, dict):
                continue
            name = str(item.get("name") or "").strip().rstrip("/")
            if not name or not name.startswith("live/"):
                continue
            if not _path_online(item):
                continue
            meta = _decode_path(name)
            ip = publishers_by_path.get(name, "")
            if not ip and fallback_publishers:
                fb = fallback_publishers.get(name) or {}
                ip = str(fb.get("ip") or "")
            tracks = item.get("tracks") or []
            if tracks and isinstance(tracks[0], dict):
                tracks = [t.get("codec") or t.get("type") or str(t) for t in tracks]
            readers = item.get("readers") or []
            items.append(
                {
                    **meta,
                    "ip": ip or "—",
                    "online_desde": item.get("onlineTime")
                    or item.get("availableTime")
                    or item.get("readyTime")
                    or "",
                    "tracks": tracks,
                    "leitores": len(readers) if isinstance(readers, list) else 0,
                }
            )
    elif fallback_publishers:
        fonte = "log-watch"
        for path, info in fallback_publishers.items():
            meta = _decode_path(path)
            items.append(
                {
                    **meta,
                    "ip": str(info.get("ip") or "—"),
                    "online_desde": str(info.get("since") or ""),
                    "tracks": [],
                    "leitores": 0,
                }
            )
    else:
        return {
            "status": "ok",
            "dados": [],
            "fonte": "indisponivel",
            "aviso": "API MediaMTX inacessível. Defina MEDIAMTX_API_BASE no confvision-rtmp-guard "
            "(ex.: http://foxpro_confvision:9997) e as credenciais MEDIAMTX_API_USER/PASS.",
        }

    items.sort(key=lambda x: (x.get("camera_id") is None, x.get("camera_id") or 0, x.get("path") or ""))
    return {"status": "ok", "dados": items, "fonte": fonte, "total": len(items)}
