"""Smoke test: lista RTSP pull via API vis (usa .env do ConfCam se existir). Não imprime URLs RTSP."""

from __future__ import annotations

import os
import sys
from pathlib import Path

import requests

# ConfCam .env local (POSTGRES/VIS key) — só nomes de vars, sem logar valores
for candidate in (
    Path(__file__).resolve().parents[2]
    / "home"
    / "confmonit"
    / "v4.0"
    / "confcam"
    / ".env",
    Path.cwd() / ".env",
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
if str(_root) not in sys.path:
    sys.path.insert(0, str(_root))

from vis_api_auth import vis_api_headers  # noqa: E402

base = (os.getenv("XANO_BASE_URL") or "https://vision.confmonit2.com.br").rstrip("/")
worker = (os.getenv("WORKER_ID") or "worker-vps1").strip()
key_set = bool((os.getenv("VIS_WORKER_API_KEY") or "").strip())
url = f"{base}/vis_camera_query_rtsp_pull"
params = {"worker_id": worker} if worker else {}

print(f"GET {url} worker_id={worker!r} VIS_WORKER_API_KEY={'sim' if key_set else 'nao'}")
try:
    r = requests.get(url, params=params, headers=vis_api_headers(), timeout=30)
except Exception as exc:
    print(f"FAIL rede: {exc}")
    sys.exit(1)

print(f"HTTP {r.status_code}")
if r.status_code != 200:
    print(r.text[:500])
    sys.exit(1)

data = r.json()
dados = data.get("dados") if isinstance(data, dict) else None
if not isinstance(dados, list):
    print("FAIL payload sem lista dados")
    sys.exit(1)

ids = []
for c in dados:
    if not isinstance(c, dict):
        continue
    cid = int(c.get("id") or 0)
    if cid < 1:
        continue
    wake = c.get("rtsp_live_wake_until")
    has_rtsp = bool(str(c.get("rtsp_url_sec") or "").strip().lower().startswith("rtsp"))
    ids.append((cid, has_rtsp, wake is not None and str(wake).strip() not in ("", "<nil>")))

print(f"OK cameras_rtsp_pull={len(ids)}")
for cid, has_rtsp, wake in sorted(ids):
    print(f"  id={cid} rtsp_url_sec={'ok' if has_rtsp else 'MISSING'} wake_field={'set' if wake else '-'}")

if not any(cid == 34 for cid, _, _ in ids):
    print("WARN camera 34 nao esta na lista deste worker")
    sys.exit(2)

print("OK camera 34 listada para relay")
