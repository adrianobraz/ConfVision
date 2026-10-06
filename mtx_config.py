"""API de configuração de paths MediaMTX (v3)."""

from __future__ import annotations

import json
import urllib.error
import urllib.parse
import urllib.request
from typing import Any, Optional

from rtmp_online import _api_get, mtx_api_base


def _auth_headers() -> dict[str, str]:
    import os
    from base64 import b64encode

    headers: dict[str, str] = {"Content-Type": "application/json"}
    user = (os.getenv("MEDIAMTX_API_USER") or "").strip()
    password = (os.getenv("MEDIAMTX_API_PASS") or "").strip()
    if user:
        token = b64encode(f"{user}:{password}".encode("utf-8")).decode("ascii")
        headers["Authorization"] = f"Basic {token}"
    return headers


def _encode_path_name(path: str) -> str:
    p = (path or "").strip().strip("/")
    if not p:
        return ""
    return urllib.parse.quote(p, safe="/")


def mtx_path_config_get(path: str) -> Optional[dict[str, Any]]:
    enc = _encode_path_name(path)
    if not enc:
        return None
    return _api_get(f"/v3/config/paths/get/{enc}")


def _api_write(method: str, api_path: str, body: dict[str, Any], timeout: float = 15.0) -> tuple[bool, str]:
    base = mtx_api_base()
    if not base:
        return False, "MEDIAMTX_API_BASE vazio"
    url = f"{base}{api_path}"
    data = json.dumps(body).encode("utf-8")
    req = urllib.request.Request(url, data=data, method=method, headers=_auth_headers())
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            _ = resp.read()
            return True, ""
    except urllib.error.HTTPError as exc:
        try:
            detail = exc.read().decode("utf-8", errors="replace")[:300]
        except OSError:
            detail = str(exc)
        return False, f"HTTP {exc.code}: {detail}"
    except (urllib.error.URLError, TimeoutError, OSError) as exc:
        return False, str(exc)


def mtx_path_config_add(path: str, conf: dict[str, Any]) -> tuple[bool, str]:
    enc = _encode_path_name(path)
    if not enc:
        return False, "path vazio"
    return _api_write("POST", f"/v3/config/paths/add/{enc}", conf)


def mtx_path_config_patch(path: str, conf: dict[str, Any]) -> tuple[bool, str]:
    enc = _encode_path_name(path)
    if not enc:
        return False, "path vazio"
    return _api_write("PATCH", f"/v3/config/paths/patch/{enc}", conf)


def mtx_path_config_delete(path: str) -> tuple[bool, str]:
    enc = _encode_path_name(path)
    if not enc:
        return False, "path vazio"
    base = mtx_api_base()
    if not base:
        return False, "MEDIAMTX_API_BASE vazio"
    url = f"{base}/v3/config/paths/delete/{enc}"
    req = urllib.request.Request(url, method="DELETE", headers=_auth_headers())
    try:
        with urllib.request.urlopen(req, timeout=15.0) as resp:
            _ = resp.read()
            return True, ""
    except urllib.error.HTTPError as exc:
        if exc.code == 404:
            return True, ""
        try:
            detail = exc.read().decode("utf-8", errors="replace")[:300]
        except OSError:
            detail = str(exc)
        return False, f"HTTP {exc.code}: {detail}"
    except (urllib.error.URLError, TimeoutError, OSError) as exc:
        return False, str(exc)


def rtsp_on_demand_path_conf(rtsp_url: str) -> dict[str, Any]:
    """PathConf para pull RTSP só com leitor HLS (Fase 3)."""
    return {
        "source": rtsp_url.strip(),
        "sourceOnDemand": True,
        "sourceOnDemandStartTimeout": "20s",
        "sourceOnDemandCloseAfter": "30s",
        "disablePublisherOverride": True,
    }
