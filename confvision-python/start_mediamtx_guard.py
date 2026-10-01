"""Sobe Guard (auth/ban) + MediaMTX no mesmo container.

Usado por Dockerfile.mediamtx e confvision-mediamtx-node (V2).
Sequencia: Guard healthy -> MediaMTX -> API healthy -> NODE READY.
Auth via http://127.0.0.1:8100/auth
"""

from __future__ import annotations

import json
import os
import signal
import socket
import subprocess
import sys
import time
import urllib.error
import urllib.request
from base64 import b64encode

_procs: list[subprocess.Popen] = []


def _shutdown(*_args) -> None:
    print("[START] encerrando MediaMTX + Guard...", flush=True)
    for p in _procs:
        if p.poll() is None:
            p.terminate()
    for p in _procs:
        try:
            p.wait(timeout=10)
        except subprocess.TimeoutExpired:
            p.kill()
    sys.exit(0)


def _env(name: str, default: str = "") -> str:
    return (os.getenv(name) or default).strip()


def _port_in_use(host: str, port: int) -> bool:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        sock.settimeout(0.5)
        return sock.connect_ex((host, port)) == 0


def _wait_guard_http(guard: subprocess.Popen, timeout_sec: float = 30.0) -> None:
    port = _env("RTMP_GUARD_HTTP_PORT", "8100")
    url = f"http://127.0.0.1:{port}/health"
    deadline = time.time() + timeout_sec
    while time.time() < deadline:
        if guard.poll() is not None:
            print(f"[START] Guard saiu cedo code={guard.returncode}", flush=True)
            sys.exit(guard.returncode or 1)
        try:
            with urllib.request.urlopen(url, timeout=1.5) as resp:
                if 200 <= resp.status < 500:
                    print(f"[START] Guard HTTP pronto ({url})", flush=True)
                    return
        except (urllib.error.URLError, TimeoutError, OSError):
            time.sleep(0.25)
    print("[START] FALHA Guard HTTP timeout — abortando (MediaMTX nao deve subir sem auth)", flush=True)
    _shutdown()
    sys.exit(1)


def _mediamtx_api_get(path: str, timeout: float = 2.0) -> bool:
    base = _env("MEDIAMTX_API_BASE", "http://127.0.0.1:9997").rstrip("/")
    if not base:
        return False
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
            json.loads(raw or "{}")
            return 200 <= resp.status < 300
    except (urllib.error.URLError, urllib.error.HTTPError, TimeoutError, OSError, json.JSONDecodeError):
        return False


def _wait_mediamtx_api(mtx: subprocess.Popen, timeout_sec: float = 45.0) -> None:
    deadline = time.time() + timeout_sec
    while time.time() < deadline:
        if mtx.poll() is not None:
            print(f"[START] MediaMTX saiu cedo code={mtx.returncode}", flush=True)
            sys.exit(mtx.returncode or 1)
        if _mediamtx_api_get("/v3/config/global/get"):
            print("[START] MediaMTX API pronta (/v3/config/global/get)", flush=True)
            return
        time.sleep(0.35)
    print("[START] WARN MediaMTX API timeout — RTMP/RTSP podem subir mesmo assim; verifique logs", flush=True)


def _preflight_ports() -> None:
    port = int(_env("RTMP_GUARD_HTTP_PORT", "8100") or "8100")
    if _port_in_use("127.0.0.1", port):
        print(
            f"[START] FALHA porta {port} em uso — pare confvision-rtmp-guard legado ou container duplicado",
            flush=True,
        )
        sys.exit(1)
    for label, p in (("RTMP", 1935), ("RTSP", 8554)):
        if _port_in_use("0.0.0.0", p) or _port_in_use("127.0.0.1", p):
            print(f"[START] WARN porta {p} ({label}) ja em uso — possivel MediaMTX duplicado", flush=True)


def main() -> None:
    signal.signal(signal.SIGTERM, _shutdown)
    signal.signal(signal.SIGINT, _shutdown)

    cfg = _env("MEDIAMTX_CONFIG", "/mediamtx.yml")
    mtx_bin = _env("MEDIAMTX_BIN", "/mediamtx")

    _preflight_ports()

    print("[START] Guard RTMP ...", flush=True)
    guard = subprocess.Popen([sys.executable, "-u", "rtmp_guard_main.py"])
    _procs.append(guard)
    _wait_guard_http(guard)

    print(f"[START] MediaMTX {mtx_bin} {cfg} ...", flush=True)
    mtx = subprocess.Popen([mtx_bin, cfg])
    _procs.append(mtx)
    _wait_mediamtx_api(mtx)

    print("[START] NODE READY — Guard + MediaMTX (RTMP :1935 RTSP :8554 API :9997 Guard :8100)", flush=True)

    while True:
        for name, p in (("guard", guard), ("mediamtx", mtx)):
            rc = p.poll()
            if rc is not None:
                print(f"[START] {name} saiu code={rc} — derrubando o outro", flush=True)
                _shutdown()
        time.sleep(0.5)


if __name__ == "__main__":
    main()
