import socket
import threading
import time
from urllib.parse import urlparse

from config import MEDIAMTX_RTSP_BASE, SYNC_INTERVAL_SEC
from detector import PersonDetector
from xano_client import get_cameras_ativas, post_evento, post_ping


def _path_from_stream_url(url: str) -> str:
    parsed = urlparse(url)
    return parsed.path.lstrip("/")


def _rewrite_localhost_rtsp(url: str, stream_path: str) -> str:
    parsed = urlparse(url)
    host = (parsed.hostname or "").lower()
    if host not in ("127.0.0.1", "localhost"):
        return url
    path = _path_from_stream_url(url) or stream_path
    rtsp = f"{MEDIAMTX_RTSP_BASE}/{path}"
    print(f"[WARN] localhost no banco, usando MediaMTX interno: {rtsp}")
    return rtsp


def stream_path_for(camera) -> str:
    camera_id = camera.get("id")
    path = (camera.get("stream_path") or "").strip().lstrip("/")
    if path:
        return path
    if camera_id:
        return f"live/{camera_id}"
    return "live/0"


def resolve_rtsp_url(camera):
    stream_path = stream_path_for(camera)
    rtsp_default = f"{MEDIAMTX_RTSP_BASE}/{stream_path}"

    url = (camera.get("rtmp_url") or camera.get("rtsp_url") or "").strip()
    if not url:
        return rtsp_default
    if url.startswith("rtsp://"):
        return _rewrite_localhost_rtsp(url, stream_path)
    if url.startswith("rtmp://"):
        path = _path_from_stream_url(url) or stream_path
        rtsp = f"{MEDIAMTX_RTSP_BASE}/{path}"
        print(f"[WARN] camera {camera.get('id')}: RTMP no banco, usando RTSP: {rtsp}")
        return rtsp
    if "://" not in url:
        return f"{MEDIAMTX_RTSP_BASE}/{url.lstrip('/')}"
    print(f"[WARN] camera {camera.get('id')}: URL ignorada, usando path padrao: {rtsp_default}")
    return rtsp_default


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
                        f"path={stream_path_for(camera)} "
                        f"nome={camera.get('nome')} url={resolve_rtsp_url(camera)}"
                    )
        except Exception as exc:
            print(f"[ERRO] sync: {exc}")

        time.sleep(SYNC_INTERVAL_SEC)


if __name__ == "__main__":
    main()
