import os

from dotenv import load_dotenv

load_dotenv()

XANO_BASE_URL = os.getenv("XANO_BASE_URL", "").rstrip("/")
MEDIAMTX_RTSP_BASE = os.getenv(
    "MEDIAMTX_RTSP_BASE", "rtsp://foxpro_confvision:8554"
).rstrip("/")
MEDIAMTX_RTMP_PUBLISH_BASE = os.getenv(
    "MEDIAMTX_RTMP_PUBLISH_BASE", "rtmp://31.97.173.119:1935"
).rstrip("/")
MEDIAMTX_HLS_BASE = os.getenv(
    "MEDIAMTX_HLS_BASE", "http://31.97.173.119:8888"
).rstrip("/")
WORKER_ID = os.getenv("WORKER_ID", "worker-01")
WORKER_VERSION = os.getenv("WORKER_VERSION", "0.1.0")
SYNC_INTERVAL_SEC = int(os.getenv("SYNC_INTERVAL_SEC", "30"))
FRAME_SKIP = int(os.getenv("FRAME_SKIP", "5"))
YOLO_MODEL = os.getenv("YOLO_MODEL", "yolov8n.pt")
YOLO_CONF_DEFAULT = float(os.getenv("YOLO_CONF_DEFAULT", "0.5"))
