"""Sobe Guard (auth/ban) + MediaMTX no mesmo container.

Usado pelo Dockerfile.mediamtx. Auth via http://127.0.0.1:8100/auth
"""

from __future__ import annotations

import os
import signal
import subprocess
import sys
import time
import urllib.error
import urllib.request

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


def _wait_guard_http(guard: subprocess.Popen, timeout_sec: float | None = None) -> None:
    if timeout_sec is None:
        timeout_sec = float(os.getenv("GUARD_BOOT_WAIT_SEC", "30") or "30")
    port = (os.getenv("RTMP_GUARD_HTTP_PORT") or "8100").strip()
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
    print("[START] WARN Guard HTTP timeout — MediaMTX pode ver auth refused no boot", flush=True)


def main() -> None:
    signal.signal(signal.SIGTERM, _shutdown)
    signal.signal(signal.SIGINT, _shutdown)

    cfg = (os.getenv("MEDIAMTX_CONFIG") or "/mediamtx.yml").strip()
    mtx_bin = (os.getenv("MEDIAMTX_BIN") or "/mediamtx").strip()

    print("[START] Guard RTMP na :8100 ...", flush=True)
    guard = subprocess.Popen([sys.executable, "-u", "rtmp_guard_main.py"])
    _procs.append(guard)
    _wait_guard_http(guard)

    print(f"[START] MediaMTX {mtx_bin} {cfg} ...", flush=True)
    mtx = subprocess.Popen([mtx_bin, cfg])
    _procs.append(mtx)

    while True:
        for name, p in (("guard", guard), ("mediamtx", mtx)):
            rc = p.poll()
            if rc is not None:
                print(f"[START] {name} saiu code={rc} — derrubando o outro", flush=True)
                _shutdown()
        time.sleep(0.5)


if __name__ == "__main__":
    main()
