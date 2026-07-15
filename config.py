import os

from dotenv import load_dotenv

load_dotenv()

XANO_BASE_URL = os.getenv("XANO_BASE_URL", "").rstrip("/")
CONFMONIT_API_URL = os.getenv("CONFMONIT_API_URL", "").rstrip("/")
# Opcional: Bearer JWT se getArmadoById / getDadosById estiverem Seguro=true
CONFMONIT_API_TOKEN = os.getenv("CONFMONIT_API_TOKEN", "").strip()
ARMADO_CACHE_TTL_SEC = int(os.getenv("ARMADO_CACHE_TTL_SEC", "20"))
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
WORKER_VERSION = os.getenv("WORKER_VERSION", "0.3.1")
SYNC_INTERVAL_SEC = int(os.getenv("SYNC_INTERVAL_SEC", "30"))
FRAME_SKIP = int(os.getenv("FRAME_SKIP", "5"))
YOLO_MODEL = os.getenv("YOLO_MODEL", "yolov8n.pt")
YOLO_CONF_DEFAULT = float(os.getenv("YOLO_CONF_DEFAULT", "0.5"))
YOLO_DEVICE = os.getenv("YOLO_DEVICE", "").strip()

CLIP_DURACAO_SEG = int(os.getenv("CLIP_DURACAO_SEG", "20"))
SNAPSHOT_JPEG_QUALITY = int(os.getenv("SNAPSHOT_JPEG_QUALITY", "85"))
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

# Notifica terminal (receptorWeb) apos deteccao analitica — assincrono, sem Xano
RECEPTOR_WEB_URL = os.getenv("RECEPTOR_WEB_URL", "").rstrip("/")
RECEPTOR_WEB_SENHA = os.getenv("RECEPTOR_WEB_SENHA", "")
TERMINAL_NOTIFY_ENABLED = os.getenv("TERMINAL_NOTIFY_ENABLED", "true").strip().lower() in (
    "1",
    "true",
    "yes",
    "on",
)
TERMINAL_NOTIFY_WORKERS = int(os.getenv("TERMINAL_NOTIFY_WORKERS", "4"))
TERMINAL_NOTIFY_RETRIES = int(os.getenv("TERMINAL_NOTIFY_RETRIES", "3"))
TERMINAL_CONTACT_ID = os.getenv("TERMINAL_CONTACT_ID", "CV01").strip() or "CV01"

CONTABO_S3_ACCESS_KEY = os.getenv("CONTABO_S3_ACCESS_KEY", "")
CONTABO_S3_SECRET_KEY = os.getenv("CONTABO_S3_SECRET_KEY", "")
CONTABO_S3_ENDPOINT = os.getenv("CONTABO_S3_ENDPOINT", "https://usc1.contabostorage.com")
CONTABO_S3_BUCKET = os.getenv("CONTABO_S3_BUCKET", "confvision")
CONTABO_S3_REGION = os.getenv("CONTABO_S3_REGION", "us-east-1")
CONTABO_S3_TENANT_ID = os.getenv("CONTABO_S3_TENANT_ID", "").strip()

# DVR — gravacao continua (MediaMTX record + upload S3)
MEDIAMTX_API_BASE = os.getenv(
    "MEDIAMTX_API_BASE", "http://31.97.173.119:9997"
).rstrip("/")
MEDIAMTX_API_USER = os.getenv("MEDIAMTX_API_USER", "").strip()
MEDIAMTX_API_PASS = os.getenv("MEDIAMTX_API_PASS", "").strip()
DVR_RECORD_DIR = os.getenv("DVR_RECORD_DIR", "/recordings").rstrip("/")
DVR_SYNC_INTERVAL_SEC = int(os.getenv("DVR_SYNC_INTERVAL_SEC", "30"))
DVR_SEGMENTO_MINUTOS_DEFAULT = int(os.getenv("DVR_SEGMENTO_MINUTOS_DEFAULT", "5"))
DVR_STABLE_SEC = int(os.getenv("DVR_STABLE_SEC", "3"))
DVR_STORAGE_CACHE_TTL_SEC = int(os.getenv("DVR_STORAGE_CACHE_TTL_SEC", "300"))
DVR_UPLOAD_RETRIES = int(os.getenv("DVR_UPLOAD_RETRIES", "3"))
DVR_WORKER_VERSION = os.getenv("DVR_WORKER_VERSION", "0.1.1")
DVR_MTX_SYNC_API = os.getenv("DVR_MTX_SYNC_API", "true").strip().lower() in (
    "1",
    "true",
    "yes",
    "on",
)

# Gravacao por movimento (worker separado: python -u motion_main.py)
MOTION_SYNC_INTERVAL_SEC = int(os.getenv("MOTION_SYNC_INTERVAL_SEC", "30"))
MOTION_WORKER_VERSION = os.getenv("MOTION_WORKER_VERSION", "0.1.0")
MOTION_RECORD_DIR = os.getenv("MOTION_RECORD_DIR", "/tmp/confvision/motion").rstrip("/")
MOTION_CLIP_MAX_SEC = int(os.getenv("MOTION_CLIP_MAX_SEC", "300"))
MOTION_POST_ROLL_SEC = int(os.getenv("MOTION_POST_ROLL_SEC", "5"))
MOTION_FRAME_SKIP = int(os.getenv("MOTION_FRAME_SKIP", "3"))
MOTION_MIN_AREA = int(os.getenv("MOTION_MIN_AREA", "1500"))
MOTION_MOG2_HISTORY = int(os.getenv("MOTION_MOG2_HISTORY", "300"))
MOTION_MOG2_THRESHOLD = int(os.getenv("MOTION_MOG2_THRESHOLD", "25"))
MOTION_RECONNECT_SEC = int(os.getenv("MOTION_RECONNECT_SEC", "5"))
MOTION_ANALYSIS_WIDTH = int(os.getenv("MOTION_ANALYSIS_WIDTH", "640"))

# Timelapse Inteligente
# Intervalo entre capturas de frame no modo timelapse (SEGUNDOS — 720 = 12 minutos)
TIMELAPSE_FRAME_INTERVALO_SEG = int(os.getenv("TIMELAPSE_FRAME_INTERVALO_SEG", "720"))
# Frames por segmento antes de montar/enviar MP4 (60 frames x 720s = 12h reais -> 1 min video)
TIMELAPSE_FRAMES_POR_SEGMENTO = int(os.getenv("TIMELAPSE_FRAMES_POR_SEGMENTO", "60"))
