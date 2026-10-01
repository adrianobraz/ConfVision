import boto3
from pathlib import Path
from botocore.config import Config
from botocore.exceptions import ClientError

from config import (
    CONTABO_S3_ACCESS_KEY,
    CONTABO_S3_BUCKET,
    CONTABO_S3_ENDPOINT,
    CONTABO_S3_REGION,
    CONTABO_S3_SECRET_KEY,
    CONTABO_S3_TENANT_ID,
)


def _enabled():
    return bool(CONTABO_S3_ACCESS_KEY and CONTABO_S3_SECRET_KEY)


def _client():
    if not _enabled():
        raise RuntimeError(
            "upload Contabo nao configurado no worker — defina CONTABO_S3_ACCESS_KEY e CONTABO_S3_SECRET_KEY no .env do FoxPro"
        )
    endpoint = CONTABO_S3_ENDPOINT or "https://usc1.contabostorage.com"
    return boto3.client(
        "s3",
        endpoint_url=endpoint,
        aws_access_key_id=CONTABO_S3_ACCESS_KEY,
        aws_secret_access_key=CONTABO_S3_SECRET_KEY,
        region_name=CONTABO_S3_REGION or "us-east-1",
        config=Config(s3={"addressing_style": "path"}),
    )


def _public_url(key: str) -> str:
    endpoint = (CONTABO_S3_ENDPOINT or "https://usc1.contabostorage.com").rstrip("/")
    bucket = CONTABO_S3_BUCKET or "confvision"
    if CONTABO_S3_TENANT_ID:
        bucket = f"{CONTABO_S3_TENANT_ID}:{bucket}"
    key = key.lstrip("/")
    return f"{endpoint}/{bucket}/{key}"


def evento_prefix(id_franqueado, vis_evento_id: int) -> str:
    franq = str(id_franqueado or "sem-franqueado").strip()
    return f"confvision/eventos/{franq}/{vis_evento_id}"


def evento_snapshot_key(id_franqueado, vis_evento_id: int) -> str:
    return f"{evento_prefix(id_franqueado, vis_evento_id)}/snapshot.jpg"


def evento_clip_key(id_franqueado, vis_evento_id: int, seq: int) -> str:
    return f"{evento_prefix(id_franqueado, vis_evento_id)}/clip_{seq:03d}.mp4"


def upload_bytes(key: str, data: bytes, content_type: str) -> str:
    cli = _client()
    bucket = CONTABO_S3_BUCKET or "confvision"
    params = {
        "Bucket": bucket,
        "Key": key,
        "Body": data,
        "ContentType": content_type,
    }
    try:
        cli.put_object(**params, ACL="public-read")
    except ClientError as exc:
        code = exc.response.get("Error", {}).get("Code", "")
        if code in ("AccessControlListNotSupported", "InvalidArgument"):
            cli.put_object(**params)
        else:
            raise RuntimeError(f"Contabo upload falhou ({code}): {exc}") from exc
    return _public_url(key)


def upload_file(file_path, key: str, content_type: str) -> str:
    path = Path(file_path) if not isinstance(file_path, Path) else file_path
    if not path.exists() or path.stat().st_size == 0:
        raise RuntimeError(f"arquivo vazio ou inexistente: {path}")

    cli = _client()
    bucket = CONTABO_S3_BUCKET or "confvision"
    params = {
        "Bucket": bucket,
        "Key": key,
        "ContentType": content_type,
    }
    with path.open("rb") as body:
        try:
            cli.put_object(**params, Body=body, ACL="public-read")
        except ClientError as exc:
            code = exc.response.get("Error", {}).get("Code", "")
            if code in ("AccessControlListNotSupported", "InvalidArgument"):
                body.seek(0)
                cli.put_object(**params, Body=body)
            else:
                raise RuntimeError(f"Contabo upload falhou ({code}): {exc}") from exc
    return _public_url(key)
