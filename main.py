import socket
import threading
import time
from urllib.parse import urlparse

from config import MEDIAMTX_RTSP_BASE, SYNC_INTERVAL_SEC
from detector import PersonDetector
from xano_client import get_cameras_ativas, post_evento, post_ping


def _path_from_stream_url(url: str) -> str:
    parsed = urlparse(url)
    path = parsed.path.lstrip("/")
    if path:
        return path
    return ""


def resolve_rtsp_url(camera):
    url = (camera.get("rtmp_url") or camera.get("rtsp_url") or "").strip()
    if url.startswith("rtsp://"):
        return url
    if url.startswith("rtmp://"):
        path = _path_from_stream_url(url) or f"live/cam_{camera['id']}"
        rtsp = f"{MEDIAMTX_RTSP_BASE}/{path}"
        print(
            f"[WARN] camera {camera.get('id')}: URL RTMP no banco convertida para RTSP: {rtsp}"
        )
        return rtsp
    if url:
        if "://" not in url:
            path = url.lstrip("/")
            return f"{MEDIAMTX_RTSP_BASE}/{path}"
        print(f"[WARN] camera {camera.get('id')}: URL desconhecida ignorada: {url}")
    return f"{MEDIAMTX_RTSP_BASE}/live/cam_{camera['id']}"


def loop_camera(camera, detector: PersonDetector):
    conf_min = float(camera.get("confianca_min") or 0.5)
    cooldown = int(camera.get("cooldown_seg") or 30)
    rtsp_url = resolve_rtsp_url(camera)

    def on_person(conf):
        try:
            post_evento(camera, conf)
        except Exception as exc:
            print(f"[ERRO] post_evento camera {camera.get('id')}: {exc}")

    while True:
        try:
            detector.process_camera(rtsp_url, conf_min, cooldown, on_person)
        except Exception as exc:
            print(f"[ERRO] camera {camera.get('id')}: {exc}")
            time.sleep(5)


def main():
    print(f"[START] ConfVision worker | host={socket.gethostname()}")
    detector = PersonDetector()
    threads = {}

    while True:
        try:
            cameras = get_cameras_ativas()
            post_ping(len(cameras))
            print(f"[SYNC] {len(cameras)} camera(s) ativa(s)")

            for camera in cameras:
                camera_id = camera["id"]
                thread = threads.get(camera_id)
                if thread is None or not thread.is_alive():
                    thread = threading.Thread(
                        target=loop_camera,
                        args=(camera, detector),
                        daemon=True,
                    )
                    threads[camera_id] = thread
                    thread.start()
                    print(
                        f"[THREAD] camera id={camera_id} "
                        f"nome={camera.get('nome')} url={resolve_rtsp_url(camera)}"
                    )
        except Exception as exc:
            print(f"[ERRO] sync: {exc}")

        time.sleep(SYNC_INTERVAL_SEC)


if __name__ == "__main__":
    main()
