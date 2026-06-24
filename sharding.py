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
