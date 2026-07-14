import threading
import time
from typing import Optional

import requests

from config import ARMADO_CACHE_TTL_SEC, CONFMONIT_API_URL

# Fabricante CAMERA no ConfMonit — armado controlado pela UI ConfVision.
# Demais fabricantes: armado vem da central (mesmo campo dispositivo.Armado).
FABRICANTE_CAMERA = "7"

_lock = threading.Lock()
_cache: dict[str, tuple[Optional[bool], float]] = {}
_fabricante_cache: dict[str, str] = {}


def _parse_dispositivo(data) -> Optional[dict]:
    if not isinstance(data, dict):
        return None
    disp = data.get("dados") if isinstance(data.get("dados"), dict) else data
    return disp if isinstance(disp, dict) else None


def _fetch_dispositivo(id_dispositivo: str) -> Optional[dict]:
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
        return _parse_dispositivo(response.json())
    except Exception as exc:
        print(f"[ARMADO] consulta dispositivo={id_dispositivo} falhou: {exc}")
        return None


def _armado_from_disp(disp: Optional[dict]) -> Optional[bool]:
    if not disp:
        return None
    armado = str(disp.get("armado") or "").strip().upper()
    if armado == "S":
        return True
    if armado == "N":
        return False
    return None


def _fabricante_from_disp(disp: Optional[dict]) -> str:
    if not disp:
        return ""
    return str(disp.get("idFabricante") or "").strip()


def _fetch_armado(id_dispositivo: str) -> Optional[bool]:
    disp = _fetch_dispositivo(id_dispositivo)
    if disp is not None:
        fab = _fabricante_from_disp(disp)
        with _lock:
            _fabricante_cache[str(id_dispositivo).strip()] = fab
    return _armado_from_disp(disp)


def fabricante_dispositivo(id_dispositivo: str) -> str:
    device_id = str(id_dispositivo or "").strip()
    if not device_id:
        return ""
    with _lock:
        return _fabricante_cache.get(device_id, "")


def is_fabricante_camera(id_dispositivo: str) -> bool:
    return fabricante_dispositivo(id_dispositivo) == FABRICANTE_CAMERA


def is_dispositivo_armado(id_dispositivo: str) -> bool:
    """Retorna True se armado==S. Fail-closed se API indisponivel.

    Mesmo campo dispositivo.Armado para:
    - fabricante CAMERA (armado pela UI ConfVision)
    - demais fabricantes (armado pela central)
    """
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
        fab = fabricante_dispositivo(device_id) or "?"
        origem = "CAMERA" if fab == FABRICANTE_CAMERA else f"central({fab})"
        print(
            f"[ARMADO] prefetch dispositivo={device_id} origem={origem} "
            f"armado={'S' if armed is True else 'N' if armed is False else '?'}"
        )
        with _lock:
            _cache[device_id] = (armed, time.time())
