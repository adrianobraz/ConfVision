from config import MEDIAMTX_HLS_BASE, MEDIAMTX_RTMP_PUBLISH_BASE, MEDIAMTX_RTSP_BASE


def stream_path(camera_id) -> str:
    if camera_id:
        return f"live/{camera_id}"
    return "live/0"


def rtsp_url(camera_id) -> str:
    return f"{MEDIAMTX_RTSP_BASE}/{stream_path(camera_id)}"


def rtmp_publish_url(camera_id) -> str:
    return f"{MEDIAMTX_RTMP_PUBLISH_BASE}/{stream_path(camera_id)}"


def hls_url(camera_id) -> str:
    return f"{MEDIAMTX_HLS_BASE}/{stream_path(camera_id)}/index.m3u8"
