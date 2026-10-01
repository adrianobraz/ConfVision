import threading

_lock = threading.Lock()
_active_ids: set[int] = set()


def set_active_camera_ids(ids) -> None:
    global _active_ids
    with _lock:
        _active_ids = {int(i) for i in ids if i is not None}


def is_camera_active(camera_id) -> bool:
    with _lock:
        if not _active_ids:
            return False
        return int(camera_id) in _active_ids
