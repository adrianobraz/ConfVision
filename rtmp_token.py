"""Chave RTMP ConfVision: 24 dígitos no path (compatível com DVR/WIFI sem query).

Formato: [9 MAC][6 últimos id_franqueado][9 vis_camera.id]
URL: rtmp://host:1935/live/{chave24}
"""

from __future__ import annotations

import hashlib
import hmac
import os
import re


def publish_secret() -> str:
    return (os.getenv("RTMP_PUBLISH_SECRET") or "").strip()


def franqueado_sufixo(id_franqueado: str | int | None) -> str:
    digits = re.sub(r"\D", "", str(id_franqueado or ""))
    if not digits:
        return ""
    if len(digits) < 6:
        return digits.zfill(6)
    return digits[-6:]


def camera_id_9(camera_id: int | str) -> str:
    return f"{int(camera_id):09d}"


def mac9(
    camera_id: int | str,
    id_franqueado: str | int | None,
    secret: str | None = None,
) -> str:
    """HMAC-SHA256 → 9 dígitos decimais. Mensagem: '{camera_id}|{franqueado6}'."""
    sec = (secret if secret is not None else publish_secret()).encode("utf-8")
    fra6 = franqueado_sufixo(id_franqueado)
    if not sec or not fra6:
        return ""
    msg = f"{int(camera_id)}|{fra6}".encode("utf-8")
    dig = hmac.new(sec, msg, hashlib.sha256).digest()
    n = int.from_bytes(dig[:8], "big") % 1_000_000_000
    return f"{n:09d}"


def chave_rtmp(
    camera_id: int | str,
    id_franqueado: str | int | None,
    secret: str | None = None,
) -> str:
    fra6 = franqueado_sufixo(id_franqueado)
    m = mac9(camera_id, fra6, secret=secret)
    if not m or not fra6:
        return ""
    return f"{m}{fra6}{camera_id_9(camera_id)}"


def parse_chave_rtmp(chave: str) -> tuple[str, str, int] | None:
    """Retorna (mac9, franqueado6, camera_id) ou None."""
    s = (chave or "").strip()
    if len(s) != 24 or not s.isdigit():
        return None
    mac = s[:9]
    fra6 = s[9:15]
    camera_id = int(s[15:24])
    if camera_id < 1:
        return None
    return mac, fra6, camera_id


def chave_valida(chave: str, secret: str | None = None) -> bool:
    parsed = parse_chave_rtmp(chave)
    if not parsed:
        return False
    mac, fra6, camera_id = parsed
    esperado = mac9(camera_id, fra6, secret=secret)
    if not esperado:
        return False
    return hmac.compare_digest(esperado, mac)


# --- legado (query ?pass= HMAC url-safe) — mantido só se algum cliente ainda usar ---

def token_for_camera_id(camera_id: int | str, secret: str | None = None) -> str:
    import base64

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
