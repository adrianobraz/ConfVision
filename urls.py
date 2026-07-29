from config import MEDIAMTX_HLS_BASE, MEDIAMTX_RTMP_PUBLISH_BASE, MEDIAMTX_RTSP_BASE
from rtmp_token import chave_rtmp, stream_path_from_chave


def stream_path(camera_id, id_franqueado=None) -> str:
    """Path MediaMTX: cam/{hash12+} (Hashids do id). id_franqueado ignorado."""
    if camera_id is None:
        return ""
    chave = chave_rtmp(camera_id, id_franqueado)
    if chave:
        return stream_path_from_chave(chave)
    return ""


def stream_path_for_camera(camera) -> str:
    if not camera:
        return stream_path(None)
    return stream_path(camera.get("id"), camera.get("id_franqueado"))


def rtsp_url(camera_id, id_franqueado=None) -> str:
    path = stream_path(camera_id, id_franqueado)
    if not path:
        return MEDIAMTX_RTSP_BASE
    return f"{MEDIAMTX_RTSP_BASE}/{path}"


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
    if not path:
        return MEDIAMTX_RTMP_PUBLISH_BASE
    url = f"{MEDIAMTX_RTMP_PUBLISH_BASE}/{path}"
    if dvr and not url.endswith("/"):
        url += "/"
    return url


def hls_url(camera_id, id_franqueado=None) -> str:
    path = stream_path(camera_id, id_franqueado)
    if not path:
        return f"{MEDIAMTX_HLS_BASE}/index.m3u8"
    return f"{MEDIAMTX_HLS_BASE}/{path}/index.m3u8"
