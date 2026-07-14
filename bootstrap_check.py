import shutil

from config import (
    CAPTURE_DIR,
    CAPTURE_WORKERS,
    CLIP_DURACAO_SEG,
    CONTABO_S3_ACCESS_KEY,
    CONTABO_S3_BUCKET,
    EVENT_QUEUE_BACKEND,
    MAX_CAMERAS,
    MEDIAMTX_RTSP_BASE,
    REDIS_URL,
    SHARD_MODE,
    UPLOAD_WORKERS,
    WORKER_SHARD_INDEX,
    WORKER_SHARD_TOTAL,
    XANO_BASE_URL,
)
from sharding import shard_label


def validate_config():
    ok = True
    if not XANO_BASE_URL:
        print("[CONFIG] ERRO: XANO_BASE_URL nao definido")
        ok = False
    if not MEDIAMTX_RTSP_BASE:
        print("[CONFIG] ERRO: MEDIAMTX_RTSP_BASE nao definido")
        ok = False
    if not shutil.which("ffmpeg"):
        print("[CONFIG] ERRO: ffmpeg nao encontrado no PATH")
        ok = False
    if not CONTABO_S3_ACCESS_KEY:
        print("[CONFIG] ERRO: CONTABO_S3_ACCESS_KEY nao definido — captura vai falhar no upload")
        ok = False
    if not CONTABO_S3_BUCKET:
        print("[CONFIG] AVISO: CONTABO_S3_BUCKET vazio, usando confvision")
    if EVENT_QUEUE_BACKEND == "redis" and not REDIS_URL:
        print("[CONFIG] AVISO: EVENT_QUEUE_BACKEND=redis sem REDIS_URL — fila cai para memory")
    if SHARD_MODE == "hash" and (WORKER_SHARD_TOTAL <= 0 or WORKER_SHARD_INDEX < 0):
        print("[CONFIG] AVISO: SHARD_MODE=hash requer WORKER_SHARD_INDEX e WORKER_SHARD_TOTAL")
    try:
        Path = __import__("pathlib").Path
        Path(CAPTURE_DIR).mkdir(parents=True, exist_ok=True)
    except Exception as exc:
        print(f"[CONFIG] AVISO: nao foi possivel criar CAPTURE_DIR={CAPTURE_DIR}: {exc}")
    if ok:
        print(
            f"[CONFIG] OK | xano={XANO_BASE_URL} | rtsp={MEDIAMTX_RTSP_BASE} "
            f"| clip={CLIP_DURACAO_SEG}s | contabo=sim | upload_workers={UPLOAD_WORKERS} "
            f"| capture_workers={CAPTURE_WORKERS} | {shard_label()} | capture_dir={CAPTURE_DIR}"
        )
    return ok
