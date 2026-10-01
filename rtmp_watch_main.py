"""Serviço ConfVision: monitora falhas RTMP do MediaMTX.

Uso EasyPanel:
  Arguments: -u rtmp_watch_main.py

Variáveis:
  MTX_LOG_FILE=/recordings/mediamtx.log
  RTMP_WATCH_JSON=/recordings/rtmp_falhas.json
  RTMP_WATCH_HTTP_PORT=8099
  RTMP_WATCH_MAX=500
"""

from __future__ import annotations

import json
import os
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

from rtmp_watch import RtmpLogParser, RtmpWatchStore, follow_file


def _env(name: str, default: str = "") -> str:
    return (os.getenv(name) or default).strip()


STORE = RtmpWatchStore(
    max_items=int(_env("RTMP_WATCH_MAX", "500") or "500"),
    dedupe_sec=int(_env("RTMP_WATCH_DEDUPE_SEC", "60") or "60"),
)
PARSER = RtmpLogParser(STORE)


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

    def do_GET(self):
        u = urlparse(self.path)
        qs = parse_qs(u.query)
        path = u.path.rstrip("/") or "/"

        if path in ("/", "/health"):
            self._json(200, {"status": "ok", "service": "confvision-rtmp-watch"})
            return

        if path == "/falhas":
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
            self._json(200, {"status": "ok", "dados": STORE.resumo()})
            return

        self._json(404, {"status": "nao encontrado"})


def main():
    log_file = _env("MTX_LOG_FILE", "/recordings/mediamtx.log")
    json_out = _env("RTMP_WATCH_JSON", "/recordings/rtmp_falhas.json")
    port = int(_env("RTMP_WATCH_HTTP_PORT", "8099") or "8099")

    print(
        f"[RTMP-WATCH] START | log={log_file} | json={json_out} | http=:{port}",
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
    print(f"[RTMP-WATCH] HTTP em 0.0.0.0:{port}  GET /falhas /resumo /health", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
