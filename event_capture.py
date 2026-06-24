import traceback
from pathlib import Path

from capture import (
    capture_snapshot_jpeg_file,
    cleanup_work_dir,
    evento_work_dir,
    finalizar_captura,
    record_clip_mp4_file,
    try_iniciar_captura,
)
from camera_state import is_camera_active
from config import CLIP_DURACAO_SEG
from storage import evento_clip_key, evento_snapshot_key
from upload_queue import submit_upload
from urls import rtsp_url
from xano_client import create_evento, post_evento_clip, put_evento


def processar_deteccao(camera, confianca: float):
    camera_id = camera.get("id")
    if not is_camera_active(camera_id):
        print(f"[CAPTURA] camera={camera_id} inativa, evento ignorado")
        return

    id_franqueado = camera.get("id_franqueado")
    rtsp = rtsp_url(camera_id)
    evento_id = None
    evento = None
    work_dir: Path | None = None
    snapshot_future = None
    clip_future = None

    if not try_iniciar_captura(camera_id):
        print(f"[CAPTURA] camera={camera_id} captura ja em andamento, evento ignorado")
        return

    try:
        evento = create_evento(camera, confianca, status="capturando")
        evento_id = _evento_id(evento)
        if not evento_id:
            raise RuntimeError(f"Xano nao retornou id do evento: {evento}")

        work_dir = evento_work_dir(evento_id)
        snapshot_path = work_dir / "snapshot.jpg"
        clip_path = work_dir / "clip_001.mp4"

        print(
            f"[CAPTURA] evento id={evento_id} camera={camera_id} rtsp={rtsp} "
            f"duracao={CLIP_DURACAO_SEG}s dir={work_dir}"
        )

        capture_snapshot_jpeg_file(rtsp, snapshot_path)
        print(f"[CAPTURA] snapshot local ok evento={evento_id} bytes={snapshot_path.stat().st_size}")

        snapshot_key = evento_snapshot_key(id_franqueado, evento_id)
        snapshot_future = submit_upload(snapshot_path, snapshot_key, "image/jpeg")

        record_clip_mp4_file(rtsp, clip_path, CLIP_DURACAO_SEG)
        print(f"[CAPTURA] clip local ok evento={evento_id} bytes={clip_path.stat().st_size}")

        clip_key = evento_clip_key(id_franqueado, evento_id, 1)
        clip_future = submit_upload(clip_path, clip_key, "video/mp4")

    except Exception as exc:
        print(f"[ERRO] captura local evento camera={camera_id} evento={evento_id}: {exc}")
        traceback.print_exc()
        if snapshot_future:
            try:
                snapshot_future.result(timeout=120)
            except Exception:
                pass
        if evento_id:
            try:
                put_evento(evento_id, {"status": "erro", "clip_count": 0}, base=evento)
            except Exception as put_exc:
                print(f"[ERRO] nao foi possivel marcar evento={evento_id} como erro: {put_exc}")
        if work_dir:
            cleanup_work_dir(work_dir)
        return
    finally:
        finalizar_captura(camera_id)

    try:
        snapshot_url = snapshot_future.result()
        print(f"[CAPTURA] snapshot contabo ok evento={evento_id} url={snapshot_url}")

        video_url = clip_future.result()
        print(f"[CAPTURA] clip contabo ok evento={evento_id} url={video_url}")

        post_evento_clip(
            {
                "vis_evento_id": evento_id,
                "seq": 1,
                "video_url": video_url,
                "duracao_seg": CLIP_DURACAO_SEG,
                "snapshot_url": snapshot_url,
            }
        )

        put_evento(
            evento_id,
            {
                "snapshot_url": snapshot_url,
                "video_url": video_url,
                "clip_count": 1,
                "status": "pronto",
                "processado": False,
            },
            base=evento,
        )
        print(f"[CAPTURA] evento={evento_id} pronto snapshot+video")

    except Exception as exc:
        print(f"[ERRO] upload evento camera={camera_id} evento={evento_id}: {exc}")
        traceback.print_exc()
        if evento_id:
            try:
                put_evento(evento_id, {"status": "erro", "clip_count": 0}, base=evento)
            except Exception as put_exc:
                print(f"[ERRO] nao foi possivel marcar evento={evento_id} como erro: {put_exc}")
    finally:
        if work_dir:
            cleanup_work_dir(work_dir)


def _evento_id(evento) -> int | None:
    if not evento:
        return None
    if isinstance(evento, dict):
        if evento.get("id"):
            return int(evento["id"])
        dados = evento.get("dados")
        if isinstance(dados, dict) and dados.get("id"):
            return int(dados["id"])
    return None
