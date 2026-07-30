import socket
import time
from pathlib import Path

from config import (
    MOTION_CLIP_MAX_SEC,
    MOTION_POST_ROLL_SEC,
    MOTION_RECORD_DIR,
    MOTION_SYNC_INTERVAL_SEC,
    MOTION_WORKER_VERSION,
    WORKER_ID,
    XANO_BASE_URL,
    MEDIAMTX_RTSP_BASE,
)
from motion_worker import MotionWorkerManager
from bootstrap_check import check_rtmp_publish_secret
from sharding import filter_gravacao_cameras, shard_label
from xano_client import get_cameras_gravacao_ativas, post_ping


def _validate_motion_config() -> bool:
    ok = True
    if not XANO_BASE_URL:
        print("[MOTION] ERRO: XANO_BASE_URL nao definido")
        ok = False
    if not MEDIAMTX_RTSP_BASE:
        print("[MOTION] ERRO: MEDIAMTX_RTSP_BASE nao definido")
        ok = False
    if not check_rtmp_publish_secret("MOTION"):
        ok = False
    try:
        Path(MOTION_RECORD_DIR).mkdir(parents=True, exist_ok=True)
    except Exception as exc:
        print(
            f"[MOTION] AVISO: nao foi possivel criar "
            f"MOTION_RECORD_DIR={MOTION_RECORD_DIR}: {exc}"
        )
    if ok:
        print(
            f"[MOTION] OK | xano={XANO_BASE_URL} | rtsp={MEDIAMTX_RTSP_BASE} "
            f"| record_dir={MOTION_RECORD_DIR} | sync={MOTION_SYNC_INTERVAL_SEC}s "
            f"| clip_max={MOTION_CLIP_MAX_SEC}s | post_roll={MOTION_POST_ROLL_SEC}s "
            f"| {shard_label()}"
        )
    return ok


def main():
    print(
        f"[MOTION] START ConfVision motion worker | host={socket.gethostname()} "
        f"| v={MOTION_WORKER_VERSION} | worker_id={WORKER_ID}"
    )
    if not _validate_motion_config():
        raise SystemExit(1)

    manager = MotionWorkerManager()

    try:
        while True:
            try:
                cameras = filter_gravacao_cameras(
                    get_cameras_gravacao_ativas(), motion=True
                )
                manager.sync(cameras)
                post_ping(
                    manager.count(),
                    extra={
                        "worker_tipo": "motion",
                        "motion_cameras": manager.count(),
                        "versao": MOTION_WORKER_VERSION,
                    },
                )
                ids = [c["id"] for c in cameras]
                print(f"[MOTION] SYNC {len(cameras)} camera(s) movimento ids={ids}")
            except Exception as exc:
                print(f"[MOTION] SYNC ERRO: {exc}")

            time.sleep(MOTION_SYNC_INTERVAL_SEC)
    finally:
        manager.stop_all()


if __name__ == "__main__":
    main()
