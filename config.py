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
WORKER_VERSION = os.getenv("WORKER_VERSION", "0.3.0")
SYNC_INTERVAL_SEC = int(os.getenv("SYNC_INTERVAL_SEC", "30"))
FRAME_SKIP = int(os.getenv("FRAME_SKIP", "5"))
YOLO_MODEL = os.getenv("YOLO_MODEL", "yolov8n.pt")
YOLO_CONF_DEFAULT = float(os.getenv("YOLO_CONF_DEFAULT", "0.5"))
YOLO_DEVICE = os.getenv("YOLO_DEVICE", "").strip()

CLIP_DURACAO_SEG = int(os.getenv("CLIP_DURACAO_SEG", "20"))
CAPTURE_DIR = os.getenv("CAPTURE_DIR", "/tmp/confvision").rstrip("/")
UPLOAD_WORKERS = int(os.getenv("UPLOAD_WORKERS", "4"))
UPLOAD_RETRIES = int(os.getenv("UPLOAD_RETRIES", "3"))

# Escala nacional — sharding
SHARD_MODE = os.getenv("SHARD_MODE", "auto").strip().lower()
WORKER_SHARD_INDEX = int(os.getenv("WORKER_SHARD_INDEX", "-1"))
WORKER_SHARD_TOTAL = int(os.getenv("WORKER_SHARD_TOTAL", "0"))
MAX_CAMERAS = int(os.getenv("MAX_CAMERAS", "50"))

# Fila de eventos (detecção → captura)
EVENT_QUEUE_BACKEND = os.getenv("EVENT_QUEUE_BACKEND", "memory").strip().lower()
REDIS_URL = os.getenv("REDIS_URL", "").strip()
EVENT_QUEUE_KEY = os.getenv("EVENT_QUEUE_KEY", "confvision:eventos").strip()
EVENT_QUEUE_MAX_SIZE = int(os.getenv("EVENT_QUEUE_MAX_SIZE", "1000"))
CAPTURE_WORKERS = int(os.getenv("CAPTURE_WORKERS", "8"))

CONTABO_S3_ACCESS_KEY = os.getenv("CONTABO_S3_ACCESS_KEY", "")
CONTABO_S3_SECRET_KEY = os.getenv("CONTABO_S3_SECRET_KEY", "")
CONTABO_S3_ENDPOINT = os.getenv("CONTABO_S3_ENDPOINT", "https://usc1.contabostorage.com")
CONTABO_S3_BUCKET = os.getenv("CONTABO_S3_BUCKET", "confvision")
CONTABO_S3_REGION = os.getenv("CONTABO_S3_REGION", "us-east-1")
CONTABO_S3_TENANT_ID = os.getenv("CONTABO_S3_TENANT_ID", "").strip()
