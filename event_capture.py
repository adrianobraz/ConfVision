import traceback

from capture import capture_snapshot_jpeg, record_clip_mp4
from config import CLIP_DURACAO_SEG
from storage import evento_clip_key, evento_snapshot_key, upload_bytes
from urls import rtsp_url
from xano_client import create_evento, post_evento_clip, put_evento


def processar_deteccao(camera, confianca: float):
    camera_id = camera.get("id")
    id_franqueado = camera.get("id_franqueado")
    rtsp = rtsp_url(camera_id)
    evento_id = None
    evento = None

    try:
        evento = create_evento(camera, confianca, status="capturando")
        evento_id = _evento_id(evento)
        if not evento_id:
            raise RuntimeError(f"Xano nao retornou id do evento: {evento}")

        print(f"[CAPTURA] evento id={evento_id} camera={camera_id} duracao={CLIP_DURACAO_SEG}s")

        snapshot_bytes = capture_snapshot_jpeg(rtsp)
        snapshot_key = evento_snapshot_key(id_franqueado, evento_id)
        snapshot_url = upload_bytes(snapshot_key, snapshot_bytes, "image/jpeg")
        print(f"[CAPTURA] snapshot ok evento={evento_id}")

        clip_bytes = record_clip_mp4(rtsp, CLIP_DURACAO_SEG)
        clip_key = evento_clip_key(id_franqueado, evento_id, 1)
        video_url = upload_bytes(clip_key, clip_bytes, "video/mp4")
        print(f"[CAPTURA] clip ok evento={evento_id} bytes={len(clip_bytes)}")

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
        print(f"[ERRO] captura evento camera={camera_id} evento={evento_id}: {exc}")
        traceback.print_exc()
        if evento_id:
            try:
                put_evento(evento_id, {"status": "erro", "clip_count": 0}, base=evento)
            except Exception as put_exc:
                print(f"[ERRO] nao foi possivel marcar evento={evento_id} como erro: {put_exc}")


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
