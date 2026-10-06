"""Headers para API vis (ConfCam / proxy Xano)."""

from __future__ import annotations

from typing import Any, Optional

from config import VIS_WORKER_API_KEY


def vis_api_headers(extra: Optional[dict[str, str]] = None) -> dict[str, str]:
    out: dict[str, Any] = dict(extra or {})
    key = (VIS_WORKER_API_KEY or "").strip()
    if key:
        out["X-Vis-Worker-Key"] = key
    return out
