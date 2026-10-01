import threading
import time
from concurrent.futures import Future, ThreadPoolExecutor
from pathlib import Path

from config import UPLOAD_RETRIES, UPLOAD_WORKERS
from storage import upload_file

_executor: ThreadPoolExecutor | None = None
_executor_lock = threading.Lock()


def _get_executor() -> ThreadPoolExecutor:
    global _executor
    with _executor_lock:
        if _executor is None:
            _executor = ThreadPoolExecutor(
                max_workers=UPLOAD_WORKERS,
                thread_name_prefix="upload",
            )
        return _executor


def submit_upload(file_path: Path, key: str, content_type: str) -> Future:
    path = Path(file_path)
    return _get_executor().submit(_upload_with_retry, path, key, content_type)


def _upload_with_retry(file_path: Path, key: str, content_type: str) -> str:
    last_exc: Exception | None = None
    for attempt in range(1, UPLOAD_RETRIES + 1):
        try:
            url = upload_file(file_path, key, content_type)
            print(f"[UPLOAD] ok key={key} tentativa={attempt}")
            return url
        except Exception as exc:
            last_exc = exc
            print(f"[UPLOAD] falha key={key} tentativa={attempt}/{UPLOAD_RETRIES}: {exc}")
            if attempt < UPLOAD_RETRIES:
                time.sleep(min(2**attempt, 10))
    raise RuntimeError(f"upload Contabo esgotou tentativas ({key}): {last_exc}") from last_exc
