import time
from pathlib import Path
from typing import Any, Optional

import boto3
from botocore.config import Config
from botocore.exceptions import ClientError

from config import DVR_STORAGE_CACHE_TTL_SEC
from xano_client import get_gravacao_storage_credenciais

_cache: dict[str, tuple[float, dict[str, Any]]] = {}


def _resolve_bucket(storage: dict[str, Any]) -> str:
    bucket = (storage.get("s3_bucket") or "gravacoes").strip()
    tenant = (storage.get("s3_tenant_id") or "").strip()
    if tenant:
        return f"{tenant}:{bucket}"
    return bucket


def _public_url(storage: dict[str, Any], key: str) -> str:
    endpoint = (storage.get("s3_endpoint") or "https://usc1.contabostorage.com").rstrip(
        "/"
    )
    bucket = _resolve_bucket(storage)
    return f"{endpoint}/{bucket}/{key.lstrip('/')}"


def get_storage_config(id_franqueado: str) -> dict[str, Any]:
    franq = str(id_franqueado or "").strip()
    if not franq:
        raise RuntimeError("id_franqueado obrigatorio para upload de gravacao")

    cached = _cache.get(franq)
    now = time.time()
    if cached and now - cached[0] < DVR_STORAGE_CACHE_TTL_SEC:
        return cached[1]

    storage = get_gravacao_storage_credenciais(franq)
    if not storage.get("s3_access_key") or not storage.get("s3_secret_key"):
        raise RuntimeError(f"credenciais S3 incompletas para franqueado={franq}")

    _cache[franq] = (now, storage)
    return storage


def segment_s3_key(id_franqueado: str, camera_id: int, filename: str) -> str:
    franq = str(id_franqueado or "sem-franqueado").strip()
    return f"{franq}/camera-{camera_id}/{filename}"


def _client(storage: dict[str, Any]):
    endpoint = (storage.get("s3_endpoint") or "https://usc1.contabostorage.com").rstrip(
        "/"
    )
    return boto3.client(
        "s3",
        endpoint_url=endpoint,
        aws_access_key_id=storage["s3_access_key"],
        aws_secret_access_key=storage["s3_secret_key"],
        region_name="us-east-1",
        config=Config(s3={"addressing_style": "path"}),
    )


def upload_segment_file(
    file_path: Path,
    id_franqueado: str,
    camera_id: int,
    storage: Optional[dict[str, Any]] = None,
) -> tuple[str, str]:
    storage = storage or get_storage_config(id_franqueado)
    if not file_path.exists() or file_path.stat().st_size == 0:
        raise RuntimeError(f"arquivo vazio ou inexistente: {file_path}")

    key = segment_s3_key(id_franqueado, camera_id, file_path.name)
    bucket = storage.get("s3_bucket") or "gravacoes"
    cli = _client(storage)
    params = {
        "Bucket": bucket,
        "Key": key,
        "ContentType": "video/mp4",
    }
    with file_path.open("rb") as body:
        try:
            cli.put_object(**params, Body=body, ACL="public-read")
        except ClientError as exc:
            code = exc.response.get("Error", {}).get("Code", "")
            if code in ("AccessControlListNotSupported", "InvalidArgument"):
                body.seek(0)
                cli.put_object(**params, Body=body)
            else:
                raise RuntimeError(f"upload gravacao falhou ({code}): {exc}") from exc

    return key, _public_url(storage, key)
