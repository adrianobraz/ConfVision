import re
import time
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import Any, Optional

from config import DVR_SEGMENTO_MINUTOS_DEFAULT, DVR_UPLOAD_RETRIES
from gravacao_storage import get_storage_config, upload_segment_file
from xano_client import post_gravacao_segmento

_FILENAME_RE = re.compile(
    r"^(?P<date>\d{4}-\d{2}-\d{2})_(?P<time>\d{2}-\d{2}-\d{2})"
)


def _parse_inicio(filename: str, segmento_minutos: int) -> datetime:
    match = _FILENAME_RE.match(Path(filename).stem)
    if match:
        dt = datetime.strptime(
            f"{match.group('date')} {match.group('time').replace('-', ':')}",
            "%Y-%m-%d %H:%M:%S",
        )
        return dt.replace(tzinfo=timezone.utc)
    fim = datetime.fromtimestamp(
        Path(filename).stat().st_mtime, tz=timezone.utc
    )
    return fim - timedelta(minutes=segmento_minutos)


def process_segment_file(
    file_path: Path,
    camera: dict[str, Any],
    segmento_minutos: int = DVR_SEGMENTO_MINUTOS_DEFAULT,
    tipo: str = "continua",
    inicio_em: Optional[datetime] = None,
    fim_em: Optional[datetime] = None,
) -> bool:
    camera_id = int(camera["id"])
    id_franqueado = str(camera.get("id_franqueado") or "").strip()
    if not id_franqueado:
        print(f"[DVR] camera={camera_id} sem id_franqueado — segmento ignorado")
        return False

    if file_path.suffix.lower() in (".part", ".tmp", ".uploaded"):
        return False

    size = file_path.stat().st_size
    if size == 0:
        print(f"[DVR] arquivo vazio ignorado: {file_path}")
        return False

    segmento = int(segmento_minutos or DVR_SEGMENTO_MINUTOS_DEFAULT)
    if inicio_em is not None:
        inicio = inicio_em if inicio_em.tzinfo else inicio_em.replace(tzinfo=timezone.utc)
    else:
        inicio = _parse_inicio(file_path.name, segmento)

    if fim_em is not None:
        fim = fim_em if fim_em.tzinfo else fim_em.replace(tzinfo=timezone.utc)
        duracao_seg = max(1, int((fim - inicio).total_seconds()))
    else:
        fim = inicio + timedelta(minutes=segmento)
        duracao_seg = segmento * 60

    storage = get_storage_config(id_franqueado)
    last_exc: Optional[Exception] = None
    s3_key = ""
    s3_url = ""

    for attempt in range(1, DVR_UPLOAD_RETRIES + 1):
        try:
            s3_key, s3_url = upload_segment_file(
                file_path, id_franqueado, camera_id, storage=storage
            )
            break
        except Exception as exc:
            last_exc = exc
            print(
                f"[DVR] upload tentativa {attempt}/{DVR_UPLOAD_RETRIES} "
                f"camera={camera_id} falhou: {exc}"
            )
            time.sleep(min(attempt * 2, 10))

    if not s3_key:
        print(f"[DVR] upload esgotado camera={camera_id} file={file_path.name}: {last_exc}")
        return False

    payload = {
        "vis_camera_id": camera_id,
        "vis_gravacao_storage_id": camera.get("vis_gravacao_storage_id"),
        "inicio_em": inicio.isoformat(),
        "fim_em": fim.isoformat(),
        "duracao_seg": duracao_seg,
        "s3_key": s3_key,
        "s3_url": s3_url,
        "tamanho_bytes": size,
        "status": "disponivel",
        "tipo": tipo or "continua",
        "uploaded_em": datetime.now(timezone.utc).isoformat(),
    }

    try:
        result = post_gravacao_segmento(payload)
        print(
            f"[DVR] segmento registrado camera={camera_id} "
            f"id={result.get('id')} key={s3_key} size={size}"
        )
    except Exception as exc:
        err = str(exc)
        if "duplicate" in err.lower() or "unique" in err.lower():
            print(f"[DVR] segmento duplicado camera={camera_id} inicio={inicio.isoformat()}")
        else:
            print(f"[DVR] POST segmento falhou camera={camera_id}: {exc}")
            return False

    uploaded = file_path.with_suffix(file_path.suffix + ".uploaded")
    try:
        file_path.rename(uploaded)
    except OSError:
        pass
    return True
