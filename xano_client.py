from datetime import datetime, timezone

import requests

from config import WORKER_ID, WORKER_VERSION, XANO_BASE_URL


def get_cameras_ativas():
    url = f"{XANO_BASE_URL}/vis_camera_query_ativas"
    response = requests.get(url, timeout=15)
    response.raise_for_status()
    data = response.json()
    cameras = data.get("dados", data) if isinstance(data, dict) else data
    return [
        camera
        for camera in cameras
        if camera.get("ativo") and camera.get("deteccao_humano")
    ]


def post_evento(camera, confianca, tipo="humano"):
    url = f"{XANO_BASE_URL}/vis_evento"
    payload = {
        "vis_camera_id": camera["id"],
        "id_franqueado": camera.get("id_franqueado"),
        "id_cliente": camera.get("id_cliente"),
        "id_dispositivo": camera.get("id_dispositivo"),
        "conta": camera.get("conta"),
        "particao": camera.get("particao"),
        "canal": camera.get("canal"),
        "tipo_deteccao": tipo,
        "confianca": confianca,
        "processado": False,
        "ignorado": False,
    }
    response = requests.post(url, json=payload, timeout=15)
    response.raise_for_status()
    return response.json()


def post_ping(cameras_ativas: int):
    url = f"{XANO_BASE_URL}/vis_worker_ping"
    payload = {
        "worker_id": WORKER_ID,
        "hostname": WORKER_ID,
        "versao": WORKER_VERSION,
        "cameras_ativas": cameras_ativas,
        "ultimo_ping_em": datetime.now(timezone.utc).isoformat(),
        "ativo": True,
    }
    response = requests.post(url, json=payload, timeout=15)
    response.raise_for_status()
    return response.json()
