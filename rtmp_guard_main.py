"""Serviço ConfVision: auth RTMP (MediaMTX) + ban/desban + falhas.

EasyPanel:
  Arguments: -u rtmp_guard_main.py

Porta padrão: 8100
"""

from __future__ import annotations

import json
import os
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Optional
from urllib.parse import parse_qs, urlparse

from rtmp_ban import BanStore
from rtmp_guard import RtmpGuard
from rtmp_token import publish_secret
from rtmp_watch import RtmpLogParser, RtmpWatchStore, follow_file


def _env(name: str, default: str = "") -> str:
    return (os.getenv(name) or default).strip()


BANS = BanStore(
    _env("RTMP_BAN_JSON", "/recordings/rtmp_bans.json"),
    max_fails=int(_env("RTMP_BAN_MAX_FAILS", "20") or "20"),
    window_sec=int(_env("RTMP_BAN_WINDOW_SEC", "60") or "60"),
    ban_ttl_sec=int(_env("RTMP_BAN_TTL_SEC", "3600") or "3600"),
)
GUARD = RtmpGuard(BANS)
STORE = RtmpWatchStore(
    max_items=int(_env("RTMP_WATCH_MAX", "500") or "500"),
    dedupe_sec=int(_env("RTMP_WATCH_DEDUPE_SEC", "60") or "60"),
)
ADMIN_KEY = _env("RTMP_GUARD_ADMIN_KEY")


class _WatchWithBan(RtmpLogParser):
    """Ao detectar falha EOF sem publish, conta para auto-ban."""

    def _handle_rtmp_conn(self, ts: str, sub: str, msg: str):
        falha = super()._handle_rtmp_conn(ts, sub, msg)
        if falha and falha.motivo_codigo in (
            "eof_sem_publish",
            "path_barra_final",
            "auth_falhou",
            "path_invalido",
        ):
            BANS.registrar_falha(falha.ip, motivo=falha.motivo_codigo)
        return falha


PARSER = _WatchWithBan(STORE)


def _admin_ok(handler: BaseHTTPRequestHandler) -> bool:
    if not ADMIN_KEY:
        return True
    key = handler.headers.get("X-RTMP-Guard-Key") or ""
    if key == ADMIN_KEY:
        return True
    u = urlparse(handler.path)
    qs = parse_qs(u.query)
    return (qs.get("key") or [""])[0] == ADMIN_KEY


class Handler(BaseHTTPRequestHandler):
    def log_message(self, fmt, *args):
        return

    def _json(self, code: int, payload: dict):
        body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
        self.send_response(code)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Access-Control-Allow-Origin", "*")
        self.end_headers()
        self.wfile.write(body)

    def _read_json(self) -> dict:
        try:
            n = int(self.headers.get("Content-Length") or "0")
        except ValueError:
            n = 0
        raw = self.rfile.read(n) if n > 0 else b"{}"
        try:
            data = json.loads(raw.decode("utf-8") or "{}")
            return data if isinstance(data, dict) else {}
        except Exception:
            return {}

    def do_OPTIONS(self):
        self.send_response(204)
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type, X-RTMP-Guard-Key")
        self.end_headers()

    def do_GET(self):
        u = urlparse(self.path)
        qs = parse_qs(u.query)
        path = u.path.rstrip("/") or "/"

        if path in ("/", "/health"):
            self._json(
                200,
                {
                    "status": "ok",
                    "service": "confvision-rtmp-guard",
                    "secret_configured": bool(publish_secret()),
                    "bans": len(BANS.listar()),
                },
            )
            return

        if path == "/falhas":
            if not _admin_ok(self):
                self._json(401, {"status": "nao autorizado"})
                return
            limit = int((qs.get("limit") or ["100"])[0] or "100")
            ip = (qs.get("ip") or [""])[0]
            stream = (qs.get("path") or [""])[0]
            self._json(
                200,
                {
                    "status": "ok",
                    "dados": STORE.listar(limit=limit, ip=ip, path=stream),
                    "resumo": STORE.resumo(),
                },
            )
            return

        if path == "/resumo":
            if not _admin_ok(self):
                self._json(401, {"status": "nao autorizado"})
                return
            self._json(200, {"status": "ok", "dados": STORE.resumo()})
            return

        if path == "/bans":
            if not _admin_ok(self):
                self._json(401, {"status": "nao autorizado"})
                return
            self._json(200, {"status": "ok", "dados": BANS.listar()})
            return

        self._json(404, {"status": "nao encontrado"})

    def do_POST(self):
        u = urlparse(self.path)
        path = u.path.rstrip("/") or "/"

        if path == "/auth":
            payload = self._read_json()
            code, motivo = GUARD.authorize(payload)
            # MediaMTX: 2xx = ok; qualquer outro = fail
            body = {"status": "ok" if code < 300 else "negado", "motivo": motivo}
            raw = json.dumps(body).encode("utf-8")
            self.send_response(code)
            self.send_header("Content-Type", "application/json; charset=utf-8")
            self.send_header("Content-Length", str(len(raw)))
            self.end_headers()
            self.wfile.write(raw)
            action = payload.get("action")
            ip = payload.get("ip")
            if code >= 400:
                print(
                    f"[RTMP-GUARD] NEGADO action={action} ip={ip} path={payload.get('path')} motivo={motivo}",
                    flush=True,
                )
            elif action == "publish":
                print(
                    f"[RTMP-GUARD] OK publish ip={ip} path={payload.get('path')}",
                    flush=True,
                )
            return

        if path == "/ban":
            if not _admin_ok(self):
                self._json(401, {"status": "nao autorizado"})
                return
            data = self._read_json()
            ip = str(data.get("ip") or "").strip()
            if not ip:
                self._json(400, {"status": "ip obrigatorio"})
                return
            motivo = str(data.get("motivo") or "manual").strip() or "manual"
            ttl = data.get("ttl_sec")
            entry = BANS.ban(
                ip,
                motivo=motivo,
                manual=True,
                ttl_sec=int(ttl) if ttl is not None else None,
            )
            self._json(200, {"status": "ok", "dados": entry.to_dict()})
            return

        if path == "/unban":
            if not _admin_ok(self):
                self._json(401, {"status": "nao autorizado"})
                return
            data = self._read_json()
            ip = str(data.get("ip") or "").strip()
            if not ip:
                self._json(400, {"status": "ip obrigatorio"})
                return
            ok = BANS.unban(ip)
            self._json(200, {"status": "ok", "desbanido": ok, "ip": ip})
            return

        if path == "/cache/invalidate":
            if not _admin_ok(self):
                self._json(401, {"status": "nao autorizado"})
                return
            data = self._read_json()
            cid = data.get("vis_camera_id") or data.get("camera_id")
            if cid is None:
                self._json(400, {"status": "vis_camera_id obrigatorio"})
                return
            GUARD.cache.invalidate(int(cid))
            self._json(200, {"status": "ok"})
            return

        self._json(404, {"status": "nao encontrado"})


def main():
    log_file = _env("MTX_LOG_FILE", "/recordings/mediamtx.log")
    json_out = _env("RTMP_WATCH_JSON", "/recordings/rtmp_falhas.json")
    port = int(_env("RTMP_GUARD_HTTP_PORT", "8100") or "8100")

    print(
        f"[RTMP-GUARD] START | auth+ban+watch | log={log_file} | http=:{port} | secret={'sim' if publish_secret() else 'NAO'}",
        flush=True,
    )

    t = threading.Thread(
        target=follow_file,
        args=(log_file, PARSER, STORE, json_out),
        name="rtmp-log-follow",
        daemon=True,
    )
    t.start()

    server = ThreadingHTTPServer(("0.0.0.0", port), Handler)
    print(
        f"[RTMP-GUARD] HTTP 0.0.0.0:{port}  POST /auth /ban /unban  GET /falhas /bans /health",
        flush=True,
    )
    server.serve_forever()


if __name__ == "__main__":
    main()
