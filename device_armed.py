import threading
import time
from typing import Optional

import requests

from config import ARMADO_CACHE_TTL_SEC, CONFMONIT_API_URL

# Fabricante CAMERA no ConfMonit — armado simulado pela UI ConfVision (setArmadoById).
# Demais fabricantes: armado via comando da central; ConfVision/worker leem dispositivo.Armado.
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
    """Consulta Armado via getArmadoById (dados = 'S'|'N') e atualiza fabricante."""
    if not CONFMONIT_API_URL or not id_dispositivo:
        return None

    device_id = str(id_dispositivo).strip()
    base = CONFMONIT_API_URL.rstrip("/")

    # 1) Status armado (leve)
    try:
        response = requests.post(
            f"{base}/v4/dispositivo/getArmadoById",
            json={"idDispositivo": device_id},
            timeout=10,
        )
        response.raise_for_status()
        payload = response.json()
        data = payload if isinstance(payload, dict) else {}
        # respApp.Dados → { status, dados: "S"|"N" } ou objeto
        dados = data.get("dados")
        if isinstance(dados, str):
            armado = dados.strip().upper()
        elif isinstance(dados, dict):
            armado = str(dados.get("armado") or "").strip().upper()
        else:
            armado = str(data.get("armado") or "").strip().upper()
    except Exception as exc:
        print(f"[ARMADO] getArmadoById dispositivo={device_id} falhou: {exc}")
        # fallback getDadosById
        disp = _fetch_dispositivo(device_id)
        if disp is not None:
            fab = _fabricante_from_disp(disp)
            with _lock:
                _fabricante_cache[device_id] = fab
            return _armado_from_disp(disp)
        return None

    # 2) Fabricante (cache) — opcional, não bloqueia o arme
    with _lock:
        tem_fab = device_id in _fabricante_cache
    if not tem_fab:
        disp = _fetch_dispositivo(device_id)
        if disp is not None:
            with _lock:
                _fabricante_cache[device_id] = _fabricante_from_disp(disp)

    if armado == "S":
        return True
    if armado == "N":
        return False
    return None


def fabricante_dispositivo(id_dispositivo: str) -> str:
    device_id = str(id_dispositivo or "").strip()
    if not device_id:
        return ""
    with _lock:
        cached = _fabricante_cache.get(device_id)
    if cached:
        return cached
    disp = _fetch_dispositivo(device_id)
    fab = _fabricante_from_disp(disp)
    if fab:
        with _lock:
            _fabricante_cache[device_id] = fab
    return fab


def is_fabricante_camera(id_dispositivo: str) -> bool:
    return fabricante_dispositivo(id_dispositivo) == FABRICANTE_CAMERA


def invalidate_armado(id_dispositivo: str = "") -> None:
    """Limpa cache de armado (após armar/desarmar na UI, se necessário)."""
    with _lock:
        if id_dispositivo:
            _cache.pop(str(id_dispositivo).strip(), None)
        else:
            _cache.clear()


def is_dispositivo_armado(id_dispositivo: str) -> bool:
    """Retorna True se armado==S. Fail-closed se API indisponivel.

    Cache: se estava desarmado, reconsulta em até 5s (para refletir arme local da UI).
    Se estava armado, respeita ARMADO_CACHE_TTL_SEC.
    """
    device_id = str(id_dispositivo or "").strip()
    if not device_id:
        return False

    now = time.time()
    with _lock:
        cached = _cache.get(device_id)
        if cached:
            armed_cached, ts = cached
            age = now - ts
            if armed_cached is True and age < ARMADO_CACHE_TTL_SEC:
                return True
            # Desarmado / desconhecido: TTL curto para pegar arme recente na UI
            if armed_cached is not True and age < 5:
                return False

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
