"""Notifica o terminal (receptorWeb) de forma assincrona apos deteccao analitica.

Nao passa pelo Xano: o worker posta direto no receptorWeb.
O terminal exibe foto, ao vivo e video via vis_evento_id no campo img.
"""

from __future__ import annotations

import threading
import time
from concurrent.futures import Future, ThreadPoolExecutor
from typing import Any, Optional

import requests

from config import (
    RECEPTOR_WEB_SENHA,
    RECEPTOR_WEB_URL,
    TERMINAL_CONTACT_ID,
    TERMINAL_NOTIFY_ENABLED,
    TERMINAL_NOTIFY_RETRIES,
    TERMINAL_NOTIFY_WORKERS,
)
from xano_client import put_evento

_executor: ThreadPoolExecutor | None = None
_executor_lock = threading.Lock()


def _get_executor() -> ThreadPoolExecutor:
    global _executor
    with _executor_lock:
        if _executor is None:
            _executor = ThreadPoolExecutor(
                max_workers=max(1, TERMINAL_NOTIFY_WORKERS),
                thread_name_prefix="terminal-notify",
            )
        return _executor


def _pad(value: Any, size: int) -> str:
    digits = "".join(ch for ch in str(value or "") if ch.isdigit())
    if not digits:
        return "0" * size
    if len(digits) >= size:
        return digits[-size:]
    return digits.zfill(size)


def _zona(value: Any) -> str:
    raw = str(value or "").strip()
    if not raw:
        return "001"
    if raw.isdigit():
        return raw.zfill(3)
    return raw


def submit_terminal_notify(
    camera: dict,
    vis_evento_id: int,
    evento_base: Optional[dict] = None,
) -> Optional[Future]:
    """Enfileira notify sem bloquear captura/ffmpeg. Retorna Future ou None."""
    if not TERMINAL_NOTIFY_ENABLED:
        return None
    if not RECEPTOR_WEB_URL or not RECEPTOR_WEB_SENHA:
        print(
            "[TERMINAL] notify desabilitado: defina RECEPTOR_WEB_URL e RECEPTOR_WEB_SENHA"
        )
        return None
    if not vis_evento_id:
        return None

    id_franqueado = str(camera.get("id_franqueado") or "").strip()
    conta = _pad(camera.get("conta"), 4)
    particao = _pad(camera.get("particao"), 2)
    zonauser = _zona(camera.get("zonauser"))

    if not id_franqueado or conta in ("", "0000"):
        print(
            f"[TERMINAL] notify ignorado evento={vis_evento_id}: "
            f"franqueado/conta ausentes"
        )
        return None

    payload = {
        "idFranqueado": id_franqueado,
        "conta": conta,
        "particao": particao,
        "zonauser": zonauser,
        "vis_evento_id": int(vis_evento_id),
        "senha": RECEPTOR_WEB_SENHA,
        "codigo": TERMINAL_CONTACT_ID,
    }

    return _get_executor().submit(
        _notify_with_retry, payload, evento_base if isinstance(evento_base, dict) else {}
    )


def _notify_with_retry(payload: dict, evento_base: dict) -> dict | None:
    url = f"{RECEPTOR_WEB_URL}/recebe-evento-confvision"
    last_exc: Exception | None = None
    vis_id = payload.get("vis_evento_id")

    for attempt in range(1, TERMINAL_NOTIFY_RETRIES + 1):
        try:
            response = requests.post(url, json=payload, timeout=15)
            if response.status_code >= 400:
                raise RuntimeError(
                    f"HTTP {response.status_code}: {response.text[:300]}"
                )
            data = response.json() if response.content else {}
            dados = data.get("dados") if isinstance(data, dict) else None
            if not isinstance(dados, dict):
                dados = data if isinstance(data, dict) else {}

            id_evento = str(dados.get("idEvento") or "").strip()
            id_processo = str(dados.get("idProcesso") or "").strip()
            print(
                f"[TERMINAL] ok evento={vis_id} tentativa={attempt} "
                f"idEvento={id_evento} processo={id_processo} "
                f"conta={payload.get('conta')} part={payload.get('particao')} "
                f"zona={payload.get('zonauser')}"
            )

            if id_evento or id_processo:
                try:
                    campos = {}
                    if id_evento:
                        campos["id_evento"] = id_evento
                    if id_processo:
                        campos["id_processo"] = id_processo
                    put_evento(int(vis_id), campos, base=evento_base)
                except Exception as put_exc:
                    print(
                        f"[TERMINAL] aviso: nao atualizou vis_evento={vis_id} "
                        f"com ids do terminal: {put_exc}"
                    )

            return dados if isinstance(dados, dict) else {"ok": True}
        except Exception as exc:
            last_exc = exc
            print(
                f"[TERMINAL] falha evento={vis_id} tentativa={attempt}/"
                f"{TERMINAL_NOTIFY_RETRIES}: {exc}"
            )
            if attempt < TERMINAL_NOTIFY_RETRIES:
                time.sleep(min(2**attempt, 10))

    print(f"[TERMINAL] esgotou retries evento={vis_id}: {last_exc}")
    return None
