from config import (
    MAX_CAMERAS,
    SHARD_MODE,
    WORKER_ID,
    WORKER_SHARD_INDEX,
    WORKER_SHARD_TOTAL,
)


def shard_enabled() -> bool:
    if SHARD_MODE == "worker_id":
        return bool(WORKER_ID)
    if SHARD_MODE == "hash":
        return WORKER_SHARD_TOTAL > 0 and WORKER_SHARD_INDEX >= 0
    if SHARD_MODE == "auto":
        return bool(WORKER_ID) or (WORKER_SHARD_TOTAL > 0 and WORKER_SHARD_INDEX >= 0)
    return False


def camera_belongs_to_shard(camera_id) -> bool:
    if not shard_enabled():
        return True
    if SHARD_MODE == "worker_id":
        return True
    if WORKER_SHARD_TOTAL <= 0 or WORKER_SHARD_INDEX < 0:
        return True
    return int(camera_id) % WORKER_SHARD_TOTAL == WORKER_SHARD_INDEX


def filter_cameras(cameras: list) -> list:
    filtered = [
        camera
        for camera in cameras
        if camera.get("ativo")
        and camera.get("deteccao_humano")
        and camera_belongs_to_shard(camera.get("id"))
    ]
    if len(filtered) > MAX_CAMERAS:
        print(
            f"[SHARD] {len(filtered)} cameras no shard, limite MAX_CAMERAS={MAX_CAMERAS} — truncando"
        )
        filtered = filtered[:MAX_CAMERAS]
    return filtered


def shard_label() -> str:
    parts = [f"mode={SHARD_MODE}", f"worker_id={WORKER_ID}", f"max={MAX_CAMERAS}"]
    if WORKER_SHARD_TOTAL > 0 and WORKER_SHARD_INDEX >= 0:
        parts.append(f"shard={WORKER_SHARD_INDEX}/{WORKER_SHARD_TOTAL}")
    return " ".join(parts)


def query_params() -> dict:
    # API 2210 aceita só worker_id opcional; hash filtra no Python (filter_cameras)
    if SHARD_MODE == "worker_id" and WORKER_ID:
        return {"worker_id": WORKER_ID}
    return {}


def is_motion_camera(camera: dict) -> bool:
    if camera.get("grava_movimento"):
        return True
    return str(camera.get("modo_gravacao") or "").strip().lower() == "movimento"


def filter_gravacao_cameras(cameras: list, motion: bool = False) -> list:
    filtered = []
    for camera in cameras:
        camera_id = camera.get("id")
        if camera_id is None:
            continue
        if not camera_belongs_to_shard(camera_id):
            continue
        cam_motion = is_motion_camera(camera)
        if motion and not cam_motion:
            continue
        if not motion and cam_motion:
            continue
        filtered.append(camera)
    if len(filtered) > MAX_CAMERAS:
        modo = "movimento" if motion else "continua"
        print(
            f"[SHARD] gravacao {modo}: {len(filtered)} cameras, "
            f"limite MAX_CAMERAS={MAX_CAMERAS} — truncando"
        )
        filtered = filtered[:MAX_CAMERAS]
    return filtered
