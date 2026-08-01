import time
import traceback

from event_capture import processar_evento_sensor
from xano_client import get_camera_by_id, get_eventos_sensor_pendentes

POLL_INTERVAL_SEC = 3


def _camera_payload(raw):
    if not raw:
        return None
    if isinstance(raw, dict):
        if raw.get("id"):
            return raw
        dados = raw.get("dados")
        if isinstance(dados, dict) and dados.get("id"):
            return dados
    return None


def loop_sensor_capturas():
    print("[SENSOR] worker de captura sensor iniciado")
    while True:
        try:
            pendentes = get_eventos_sensor_pendentes(limit=10)
            if not pendentes:
                time.sleep(POLL_INTERVAL_SEC)
                continue

            for evento in pendentes:
                camera_id = evento.get("vis_camera_id")
                evento_id = evento.get("id")
                if not camera_id or not evento_id:
                    continue
                try:
                    camera_raw = get_camera_by_id(int(camera_id))
                    camera = _camera_payload(camera_raw)
                    if not camera:
                        print(f"[SENSOR] camera={camera_id} nao encontrada evento={evento_id}")
                        continue
                    print(f"[SENSOR] capturando evento={evento_id} camera={camera_id}")
                    processar_evento_sensor(camera, evento)
                except Exception as exc:
                    print(f"[ERRO] sensor evento={evento_id} camera={camera_id}: {exc}")
                    traceback.print_exc()
        except Exception as exc:
            print(f"[ERRO] loop sensor: {exc}")
            traceback.print_exc()
        time.sleep(POLL_INTERVAL_SEC)


if __name__ == "__main__":
    loop_sensor_capturas()
