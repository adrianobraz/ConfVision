from config import MEDIAMTX_HLS_BASE, MEDIAMTX_RTMP_PUBLISH_BASE, MEDIAMTX_RTSP_BASE
from rtmp_token import chave_rtmp


def stream_path(camera_id, id_franqueado=None) -> str:
    """Path MediaMTX: live/{chave24}. Requer id_franqueado + RTMP_PUBLISH_SECRET."""
    if camera_id is None:
        return "live/0"
    if id_franqueado:
        chave = chave_rtmp(camera_id, id_franqueado)
        if chave:
            return f"live/{chave}"
    # Sem franqueado/secret: fallback legado (só útil se guard estiver off)
    return f"live/{int(camera_id)}"


def stream_path_for_camera(camera) -> str:
    if not camera:
        return stream_path(None)
    return stream_path(camera.get("id"), camera.get("id_franqueado"))


def rtsp_url(camera_id, id_franqueado=None) -> str:
    return f"{MEDIAMTX_RTSP_BASE}/{stream_path(camera_id, id_franqueado)}"


def rtsp_url_for_camera(camera) -> str:
    if camera:
        sec = str(camera.get("rtsp_url_sec") or "").strip()
        if sec.lower().startswith(("rtsp://", "rtsps://")):
            return sec
        cam_id = camera.get("id")
        if cam_id is not None:
            return rtsp_url(cam_id, camera.get("id_franqueado"))
    return rtsp_url(None)


def rtmp_publish_url(camera_id, id_franqueado=None, dvr: bool = False) -> str:
    path = stream_path(camera_id, id_franqueado)
    url = f"{MEDIAMTX_RTMP_PUBLISH_BASE}/{path}"
    if dvr and not url.endswith("/"):
        url += "/"
    return url


def hls_url(camera_id, id_franqueado=None) -> str:
    return f"{MEDIAMTX_HLS_BASE}/{stream_path(camera_id, id_franqueado)}/index.m3u8"
