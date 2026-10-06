import os

os.environ.setdefault("RTMP_PUBLISH_SECRET", "teste-salt-confvision")

from rtsp_live_relay import (
    _parse_wake_until,
    _rtmp_dest,
    _sanitize_ffmpeg_stderr,
    _wake_active,
)


def test_parse_wake_until_future():
    ts = _parse_wake_until("2030-01-01T12:00:00Z")
    assert ts is not None


def test_wake_active():
    import time

    cam = {"rtsp_live_wake_until": "2030-01-01T12:00:00+00:00"}
    assert _wake_active(cam, time.time()) is True


def test_sanitize_ffmpeg_stderr_masks_urls():
    raw = "fail rtsp://admin:secret@host:554/path rtmp://x:1935/cam/abc/"
    out = _sanitize_ffmpeg_stderr(raw)
    assert "secret" not in out
    assert "rtsp://***@" in out
    assert "rtmp://***" in out


def test_rtmp_dest_dvr_trailing_slash():
    cam = {
        "id": 5,
        "protocolo": "dvr",
        "mediamtx_rtmp_public": "rtmp://example:1935",
    }
    dest = _rtmp_dest(cam)
    assert dest.endswith("/")
    assert "/cam/" in dest
