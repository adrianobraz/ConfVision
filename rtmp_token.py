"""Token RTMP: HMAC-SHA256 do id da câmera (compatível com Go)."""

from __future__ import annotations

import base64
import hashlib
import hmac
import os


def publish_secret() -> str:
    return (os.getenv("RTMP_PUBLISH_SECRET") or "").strip()


def token_for_camera_id(camera_id: int | str, secret: str | None = None) -> str:
    sec = (secret if secret is not None else publish_secret()).encode("utf-8")
    msg = str(int(camera_id)).encode("utf-8")
    dig = hmac.new(sec, msg, hashlib.sha256).digest()
    return base64.urlsafe_b64encode(dig).decode("ascii").rstrip("=")


def token_valido(camera_id: int | str, token: str, secret: str | None = None) -> bool:
    esperado = token_for_camera_id(camera_id, secret=secret)
    recebido = (token or "").strip()
    if not esperado or not recebido:
        return False
    return hmac.compare_digest(esperado, recebido)
