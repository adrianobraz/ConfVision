"""Teste E2E leve: wake API + HLS (nao loga RTSP)."""

from __future__ import annotations

import os
import sys
import time
from pathlib import Path

import requests

for candidate in (
    Path(__file__).resolve().parents[2]
    / "home"
    / "confmonit"
    / "v4.0"
    / "confcam"
    / ".env",
):
    if candidate.is_file():
        for line in candidate.read_text(encoding="utf-8", errors="replace").splitlines():
            line = line.strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            k, _, v = line.partition("=")
            k = k.strip()
            if k and k not in os.environ:
                os.environ[k] = v.strip().strip('"').strip("'")
        break

_root = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(_root))
from vis_api_auth import vis_api_headers  # noqa: E402

BASE = (os.getenv("XANO_BASE_URL") or "https://vision.confmonit2.com.br").rstrip("/")
HLS = "https://hls1.dnsid.com.br/cam/kbjmpxv7y4r6/index.m3u8"
CAMERA_ID = 34


def check_hls() -> tuple[int, str]:
    r = requests.get(HLS, timeout=20, allow_redirects=True)
    body = (r.text or "")[:200]
    return r.status_code, body


def main() -> None:
    print("=== 1) Lista RTSP pull (worker-vps1) ===")
    r = requests.get(
        f"{BASE}/vis_camera_query_rtsp_pull",
        params={"worker_id": "worker-vps1"},
        headers=vis_api_headers(),
        timeout=30,
    )
    print(f"HTTP {r.status_code}")
    if r.status_code != 200:
        print(r.text[:400])
        sys.exit(1)
    dados = r.json().get("dados") or []
    ids = [c.get("id") for c in dados if isinstance(c, dict)]
    print(f"cameras: {ids}")
    if CAMERA_ID not in ids:
        print("FAIL: camera 34 ausente")
        sys.exit(2)

    print("=== 2) HLS antes do wake ===")
    code, body = check_hls()
    print(f"HLS HTTP {code} | {body[:80]!r}")

    print("=== 3) POST vis_rtsp_relay/wake ===")
    wr = requests.post(
        f"{BASE}/vis_rtsp_relay/wake",
        json={"vis_camera_id": CAMERA_ID},
        headers=vis_api_headers({"Content-Type": "application/json"}),
        timeout=30,
    )
    print(f"wake HTTP {wr.status_code} | {(wr.text or '')[:200]}")

    print("=== 4) HLS apos wake (aguarda relay ~25s) ===")
    for i in range(5):
        time.sleep(5)
        code, body = check_hls()
        ok = code == 200 and "#EXTM3U" in body
        print(f"  t+{(i+1)*5}s HTTP {code} {'OK' if ok else '...'}")
        if ok:
            print("PASS: HLS ao vivo disponivel")
            sys.exit(0)

    print("FAIL: HLS ainda sem stream (relay/ffmpeg ou publish MTX)")
    sys.exit(3)


if __name__ == "__main__":
    main()
