import os

os.environ.setdefault("RTMP_PUBLISH_SECRET", "teste-salt-confvision")

from mtx_config import rtsp_on_demand_path_conf, _encode_path_name
from rtsp_mtx_path_sync import _path_name


def test_rtsp_on_demand_conf():
    conf = rtsp_on_demand_path_conf("rtsp://127.0.0.1/stream")
    assert conf["sourceOnDemand"] is True
    assert conf["disablePublisherOverride"] is True
    assert conf["source"].startswith("rtsp://")


def test_encode_path_name():
    assert _encode_path_name("cam/abc123456789") == "cam/abc123456789"


def test_path_name_from_camera():
    cam = {"id": 34, "protocolo": "dvr"}
    p = _path_name(cam)
    assert p.startswith("cam/")
    assert len(p.split("/", 1)[1]) >= 12
