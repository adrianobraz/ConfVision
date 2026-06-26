import socket
import time
from typing import Any

from config import (
    DVR_RECORD_DIR,
    DVR_SYNC_INTERVAL_SEC,
    DVR_WORKER_VERSION,
    MEDIAMTX_API_BASE,
    WORKER_ID,
    XANO_BASE_URL,
)
from dvr_watcher import DvrWatcher
from mediamtx_client import sync_record_paths
from xano_client import get_cameras_gravacao_ativas, post_ping


def _validate_dvr_config() -> bool:
    ok = True
    if not XANO_BASE_URL:
        print("[DVR] ERRO: XANO_BASE_URL nao definido")
        ok = False
    if not MEDIAMTX_API_BASE:
        print("[DVR] ERRO: MEDIAMTX_API_BASE nao definido")
        ok = False
    from pathlib import Path

    try:
        Path(DVR_RECORD_DIR).mkdir(parents=True, exist_ok=True)
    except Exception as exc:
        print(f"[DVR] AVISO: nao foi possivel criar DVR_RECORD_DIR={DVR_RECORD_DIR}: {exc}")

    if ok:
        print(
            f"[DVR] OK | xano={XANO_BASE_URL} | mtx_api={MEDIAMTX_API_BASE} "
            f"| record_dir={DVR_RECORD_DIR} | sync={DVR_SYNC_INTERVAL_SEC}s"
        )
    return ok


def main():
    print(
        f"[DVR] START ConfVision DVR worker | host={socket.gethostname()} "
        f"| v={DVR_WORKER_VERSION} | worker_id={WORKER_ID}"
    )
    if not _validate_dvr_config():
        raise SystemExit(1)

    cameras_cache: list[dict[str, Any]] = []
    active_ids: set[int] = set()

    def get_cameras():
        return cameras_cache

    watcher = DvrWatcher(get_cameras)
    watcher.start()

    while True:
        try:
            cameras = get_cameras_gravacao_ativas()
            cameras_cache = cameras
            active_ids = sync_record_paths(cameras, previous_ids=active_ids)
            post_ping(
                len(cameras),
                extra={
                    "worker_tipo": "dvr",
                    "dvr_cameras": len(cameras),
                    "versao": DVR_WORKER_VERSION,
                },
            )
            ids = [c["id"] for c in cameras]
            print(f"[DVR] SYNC {len(cameras)} camera(s) gravacao ids={ids}")
        except Exception as exc:
            print(f"[DVR] SYNC ERRO: {exc}")

        time.sleep(DVR_SYNC_INTERVAL_SEC)


if __name__ == "__main__":
    main()
