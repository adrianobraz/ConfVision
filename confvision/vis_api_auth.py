"""Header de autenticação para API Go ConfVision (workerAuth)."""

from __future__ import annotations

import os


def vis_api_headers(extra: dict[str, str] | None = None) -> dict[str, str]:
    """Retorna headers com X-Vis-Worker-Key se VIS_WORKER_API_KEY estiver definido."""
    headers: dict[str, str] = {}
    key = (os.getenv("VIS_WORKER_API_KEY") or "").strip()
    if key:
        headers["X-Vis-Worker-Key"] = key
    if extra:
        headers.update(extra)
    return headers
