"""Sobe Guard (auth/ban) + MediaMTX no mesmo container.

Usado pelo Dockerfile.mediamtx. Auth via http://127.0.0.1:8100/auth
"""

from __future__ import annotations

import os
import signal
import subprocess
import sys
import time

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


def main() -> None:
    signal.signal(signal.SIGTERM, _shutdown)
    signal.signal(signal.SIGINT, _shutdown)

    cfg = (os.getenv("MEDIAMTX_CONFIG") or "/mediamtx.yml").strip()
    mtx_bin = (os.getenv("MEDIAMTX_BIN") or "/mediamtx").strip()

    print("[START] Guard RTMP na :8100 ...", flush=True)
    guard = subprocess.Popen([sys.executable, "-u", "rtmp_guard_main.py"])
    _procs.append(guard)
    time.sleep(1.0)
    if guard.poll() is not None:
        print(f"[START] Guard saiu cedo code={guard.returncode}", flush=True)
        sys.exit(guard.returncode or 1)

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
