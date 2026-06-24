import shutil
import subprocess

from config import (
    CLIP_DURACAO_SEG,
    CONTABO_S3_ACCESS_KEY,
    CONTABO_S3_BUCKET,
    MEDIAMTX_RTSP_BASE,
    XANO_BASE_URL,
)


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
    if ok:
        print(
            f"[CONFIG] OK | xano={XANO_BASE_URL} | rtsp={MEDIAMTX_RTSP_BASE} "
            f"| clip={CLIP_DURACAO_SEG}s | contabo=sim"
        )
    return ok
