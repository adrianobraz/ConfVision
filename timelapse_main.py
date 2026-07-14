"""
Entry point do worker Timelapse Inteligente.

Sincroniza câmeras com modo_gravacao="timelapse" a cada MOTION_SYNC_INTERVAL_SEC segundos
e gerencia TimelapseCameraWorker para cada câmera ativa.
"""

import socket
import time
from pathlib import Path

from config import (
    MEDIAMTX_RTSP_BASE,
    MOTION_CLIP_MAX_SEC,
    MOTION_POST_ROLL_SEC,
    MOTION_RECORD_DIR,
    MOTION_SYNC_INTERVAL_SEC,
    MOTION_WORKER_VERSION,
    TIMELAPSE_FRAME_INTERVALO_SEG,
    TIMELAPSE_FRAMES_POR_SEGMENTO,
    WORKER_ID,
    XANO_BASE_URL,
)
from sharding import filter_gravacao_cameras, shard_label
from timelapse_worker import TimelapseWorkerManager
from xano_client import get_cameras_gravacao_ativas, post_ping


def _validate_config() -> bool:
    ok = True
    if not XANO_BASE_URL:
        print("[TIMELAPSE] ERRO: XANO_BASE_URL nao definido")
        ok = False
    if not MEDIAMTX_RTSP_BASE:
        print("[TIMELAPSE] ERRO: MEDIAMTX_RTSP_BASE nao definido")
        ok = False
    try:
        Path(MOTION_RECORD_DIR).mkdir(parents=True, exist_ok=True)
    except Exception as exc:
        print(
            f"[TIMELAPSE] AVISO: nao foi possivel criar "
            f"MOTION_RECORD_DIR={MOTION_RECORD_DIR}: {exc}"
        )
    if ok:
        print(
            f"[TIMELAPSE] OK | xano={XANO_BASE_URL} | rtsp={MEDIAMTX_RTSP_BASE} "
            f"| record_dir={MOTION_RECORD_DIR} | sync={MOTION_SYNC_INTERVAL_SEC}s "
            f"| frame_intervalo={TIMELAPSE_FRAME_INTERVALO_SEG}s "
            f"| frames_por_segmento={TIMELAPSE_FRAMES_POR_SEGMENTO} "
            f"| clip_max={MOTION_CLIP_MAX_SEC}s | post_roll={MOTION_POST_ROLL_SEC}s "
            f"| {shard_label()}"
        )
    return ok


def main():
    print(
        f"[TIMELAPSE] START ConfVision timelapse worker | host={socket.gethostname()} "
        f"| v={MOTION_WORKER_VERSION} | worker_id={WORKER_ID}"
    )
    if not _validate_config():
        raise SystemExit(1)

    manager = TimelapseWorkerManager()

    try:
        while True:
            try:
                cameras = filter_gravacao_cameras(
                    get_cameras_gravacao_ativas(), timelapse=True
                )
                manager.sync(cameras)
                post_ping(
                    manager.count(),
                    extra={
                        "worker_tipo": "timelapse",
                        "timelapse_cameras": manager.count(),
                        "versao": MOTION_WORKER_VERSION,
                    },
                )
                ids = [c["id"] for c in cameras]
                print(f"[TIMELAPSE] SYNC {len(cameras)} camera(s) timelapse ids={ids}")
            except Exception as exc:
                print(f"[TIMELAPSE] SYNC ERRO: {exc}")

            time.sleep(MOTION_SYNC_INTERVAL_SEC)
    finally:
        manager.stop_all()


if __name__ == "__main__":
    main()
