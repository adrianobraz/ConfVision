import socket
import threading
import time

from bootstrap_check import validate_config
from camera_state import is_camera_active, set_active_camera_ids
from config import CLIP_DURACAO_SEG, SYNC_INTERVAL_SEC
from detector import PersonDetector
from event_capture import processar_deteccao
from urls import rtsp_url, stream_path
from xano_client import get_cameras_ativas, post_ping


def loop_camera(camera, detector: PersonDetector):
    camera_id = camera.get("id")
    conf_min = float(camera.get("confianca_min") or 0.5)
    cooldown = int(camera.get("cooldown_seg") or 30)
    url = rtsp_url(camera_id)

    def on_person(conf):
        if not is_camera_active(camera_id):
            return
        threading.Thread(
            target=processar_deteccao,
            args=(camera, conf),
            daemon=True,
            name=f"capture-{camera_id}",
        ).start()

    def should_continue():
        return is_camera_active(camera_id)

    while should_continue():
        try:
            detector.process_camera(url, conf_min, cooldown, on_person, should_continue)
        except Exception as exc:
            print(f"[ERRO] camera {camera_id}: {exc}")
            if not should_continue():
                break
            time.sleep(5)

    print(f"[THREAD] camera id={camera_id} encerrada (desativada ou sem deteccao)")


def main():
    print(
        f"[START] ConfVision worker | host={socket.gethostname()} "
        f"| clip={CLIP_DURACAO_SEG}s"
    )
    validate_config()
    detector = PersonDetector()
    threads: dict[int, threading.Thread] = {}

    while True:
        try:
            cameras = get_cameras_ativas()
            active_ids = [c["id"] for c in cameras]
            set_active_camera_ids(active_ids)
            post_ping(len(cameras))
            print(f"[SYNC] {len(cameras)} camera(s) ativa(s) ids={active_ids}")

            active_set = set(active_ids)
            for camera_id, thread in list(threads.items()):
                if camera_id not in active_set and thread.is_alive():
                    print(
                        f"[SYNC] camera id={camera_id} desativada — "
                        f"aguardando thread encerrar"
                    )
                elif not thread.is_alive():
                    del threads[camera_id]

            for camera in cameras:
                camera_id = camera["id"]
                thread = threads.get(camera_id)
                if thread is None or not thread.is_alive():
                    thread = threading.Thread(
                        target=loop_camera,
                        args=(camera, detector),
                        daemon=True,
                        name=f"camera-{camera_id}",
                    )
                    threads[camera_id] = thread
                    thread.start()
                    print(
                        f"[THREAD] camera id={camera_id} "
                        f"path={stream_path(camera_id)} "
                        f"nome={camera.get('nome')} url={rtsp_url(camera_id)}"
                    )
        except Exception as exc:
            print(f"[ERRO] sync: {exc}")

        time.sleep(SYNC_INTERVAL_SEC)


if __name__ == "__main__":
    main()
