"""Chave RTMP ConfVision: Hashids do id da câmera (sem /live/).

URL: rtmp://host:1935/{hash12+}
Salt = RTMP_PUBLISH_SECRET, alfabeto 0-9a-z, min_length=12.
"""

from __future__ import annotations

import base64
import hashlib
import hmac
import os
import re

from hashids import Hashids

HASH_ALPHABET = "0123456789abcdefghijklmnopqrstuvwxyz"
HASH_MIN_LENGTH = 12
RE_HASH_PATH = re.compile(rf"^([{HASH_ALPHABET}]{{{HASH_MIN_LENGTH},}})/?$")


def publish_secret() -> str:
    return (os.getenv("RTMP_PUBLISH_SECRET") or "").strip()


def _hashids(secret: str | None = None) -> Hashids | None:
    sec = (secret if secret is not None else publish_secret()).strip()
    if not sec:
        return None
    return Hashids(salt=sec, min_length=HASH_MIN_LENGTH, alphabet=HASH_ALPHABET)


def chave_rtmp(
    camera_id: int | str,
    id_franqueado: str | int | None = None,
    secret: str | None = None,
) -> str:
    """Gera path/chave Hashids só com o id da câmera. id_franqueado é ignorado."""
    _ = id_franqueado
    h = _hashids(secret)
    try:
        cid = int(camera_id)
    except (TypeError, ValueError):
        return ""
    if not h or cid < 1:
        return ""
    return h.encode(cid)


def parse_chave_rtmp(chave: str, secret: str | None = None) -> int | None:
    """Decodifica Hashids → camera_id, ou None."""
    s = (chave or "").strip().rstrip("/")
    if s.startswith("live/"):
        s = s[5:]
    if not re.fullmatch(rf"[{HASH_ALPHABET}]{{{HASH_MIN_LENGTH},}}", s):
        return None
    h = _hashids(secret)
    if not h:
        return None
    nums = h.decode(s)
    if len(nums) != 1 or nums[0] < 1:
        return None
    return int(nums[0])


def chave_valida(chave: str, secret: str | None = None) -> bool:
    return parse_chave_rtmp(chave, secret=secret) is not None


# --- legado (query ?pass= HMAC url-safe) — mantido só se algum cliente ainda usar ---

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
