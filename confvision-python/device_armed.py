import threading
import time
from typing import Any, Optional

import requests

from config import ARMADO_CACHE_TTL_SEC, CONFMONIT_API_TOKEN, CONFMONIT_API_URL

# Fabricante CAMERA no ConfMonit — armado simulado pela UI ConfVision (setArmadoById).
# Demais fabricantes: armado via comando da central; ConfVision/worker leem dispositivo.Armado.
FABRICANTE_CAMERA = "7"

_lock = threading.Lock()
_cache: dict[str, tuple[Optional[bool], float]] = {}
_fabricante_cache: dict[str, str] = {}


def _auth_headers() -> dict[str, str]:
    token = (CONFMONIT_API_TOKEN or "").strip()
    if not token:
        return {}
    if token.lower().startswith("bearer "):
        return {"Authorization": token}
    return {"Authorization": f"Bearer {token}"}


def _parse_armado_value(valor: Any) -> Optional[bool]:
    armado = str(valor or "").strip().upper()
    if armado == "S":
        return True
    if armado == "N":
        return False
    return None


def _fetch_armado(id_dispositivo: str) -> Optional[bool]:
    """Consulta Armado via workerGetArmadoById (rota nova, sem JWT)."""
    if not CONFMONIT_API_URL or not id_dispositivo:
        print("[ARMADO] CONFMONIT_API_URL vazio — nao e possivel consultar armado")
        return None

    device_id = str(id_dispositivo).strip()
    url = f"{CONFMONIT_API_URL.rstrip('/')}/v4/dispositivo/workerGetArmadoById"

    try:
        response = requests.post(
            url,
            json={"idDispositivo": device_id},
            headers=_auth_headers(),
            timeout=10,
        )
        if response.status_code >= 400:
            print(
                f"[ARMADO] workerGetArmadoById dispositivo={device_id} "
                f"HTTP {response.status_code} body={response.text[:180]!r}"
            )
            return None

        payload = response.json()
        if not isinstance(payload, dict):
            print(
                f"[ARMADO] workerGetArmadoById resposta invalida dispositivo={device_id} "
                f"body={response.text[:180]!r}"
            )
            return None

        dados = payload.get("dados")
        armado_raw = None
        fab = ""
        if isinstance(dados, dict):
            armado_raw = dados.get("armado")
            fab = str(dados.get("idFabricante") or "").strip()
        elif isinstance(dados, str):
            armado_raw = dados

        if fab:
            with _lock:
                _fabricante_cache[device_id] = fab

        armed = _parse_armado_value(armado_raw)
        if armed is None:
            print(
                f"[ARMADO] workerGetArmadoById sem S/N dispositivo={device_id} "
                f"body={response.text[:180]!r}"
            )
        return armed
    except Exception as exc:
        print(f"[ARMADO] workerGetArmadoById dispositivo={device_id} falhou: {exc}")
        return None


def fabricante_dispositivo(id_dispositivo: str) -> str:
    device_id = str(id_dispositivo or "").strip()
    if not device_id:
        return ""
    with _lock:
        return _fabricante_cache.get(device_id, "")


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
