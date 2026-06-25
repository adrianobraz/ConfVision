import threading
import time
from typing import Optional

import requests

from config import ARMADO_CACHE_TTL_SEC, CONFMONIT_API_URL

_lock = threading.Lock()
_cache: dict[str, tuple[Optional[bool], float]] = {}


def _fetch_armado(id_dispositivo: str) -> Optional[bool]:
    if not CONFMONIT_API_URL or not id_dispositivo:
        return None

    url = f"{CONFMONIT_API_URL.rstrip('/')}/v4/dispositivo/getDadosById"
    try:
        response = requests.post(
            url,
            json={"idDispositivo": str(id_dispositivo).strip()},
            timeout=10,
        )
        response.raise_for_status()
        data = response.json()
        disp = data.get("dados") if isinstance(data, dict) else None
        if not disp and isinstance(data, dict):
            disp = data
        if not isinstance(disp, dict):
            return None
        armado = str(disp.get("armado") or "").strip().upper()
        if armado == "S":
            return True
        if armado == "N":
            return False
        return None
    except Exception as exc:
        print(f"[ARMADO] consulta dispositivo={id_dispositivo} falhou: {exc}")
        return None


def is_dispositivo_armado(id_dispositivo: str) -> bool:
    """Retorna True se armado==S. Fail-closed se API indisponivel."""
    device_id = str(id_dispositivo or "").strip()
    if not device_id:
        return False

    now = time.time()
    with _lock:
        cached = _cache.get(device_id)
        if cached and (now - cached[1]) < ARMADO_CACHE_TTL_SEC:
            return cached[0] is True

    armed = _fetch_armado(device_id)
    with _lock:
        _cache[device_id] = (armed, now)

    return armed is True


def prefetch_armado(ids) -> None:
    """Atualiza cache no sync para cameras com somente_armado."""
    for device_id in ids:
        device_id = str(device_id or "").strip()
        if not device_id:
            continue
        armed = _fetch_armado(device_id)
        with _lock:
            _cache[device_id] = (armed, time.time())
