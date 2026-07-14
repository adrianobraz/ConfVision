import traceback
from pathlib import Path

from capture import (
    CaptureCancelled,
    cancel_camera_captures,
    capture_snapshot_jpeg_file,
    cleanup_work_dir,
    evento_work_dir,
    finalizar_captura,
    install_detection_snapshot,
    start_clip_capture,
    start_parallel_capture,
    try_iniciar_captura,
)
from camera_state import is_camera_active
from config import CLIP_DURACAO_SEG
from storage import evento_clip_key, evento_snapshot_key
from upload_queue import submit_upload
from urls import rtsp_url
from xano_client import create_evento, post_evento_clip, put_evento
from terminal_notify import submit_terminal_notify


def _discard_snapshot(snapshot_source: str | None):
    if snapshot_source:
        Path(snapshot_source).unlink(missing_ok=True)


def plan_grava_foto(camera) -> bool:
    if camera.get("evento_grava_foto") is not None:
        return bool(camera.get("evento_grava_foto"))
    return bool(camera.get("captura_analitico") or camera.get("captura_sensor"))


def plan_grava_video(camera) -> bool:
    if camera.get("evento_grava_video") is not None:
        return bool(camera.get("evento_grava_video"))
    return bool(camera.get("captura_analitico") or camera.get("captura_sensor"))


def _finalizar_somente_evento(evento_id, evento):
    put_evento(
        evento_id,
        {
            "status": "pronto",
            "clip_count": 0,
            "processado": False,
        },
        base=evento,
    )
    print(f"[CAPTURA] evento={evento_id} pronto (somente registro, sem midia)")


def processar_evento_sensor(camera, evento_existente):
    """Captura midia para evento sensor ja criado no Xano pelo receptor."""
    evento_id = _evento_id(evento_existente)
    if not evento_id:
        raise RuntimeError(f"Evento sensor sem id: {evento_existente}")

    evento = evento_existente if isinstance(evento_existente, dict) else {}
    if evento.get("dados"):
        evento = evento["dados"]

    processar_deteccao(
        camera,
        float(evento.get("confianca") or 1),
        snapshot_source=None,
        evento_base=evento,
        evento_id=evento_id,
        for_sensor=True,
    )


def processar_deteccao(
    camera,
    confianca: float,
    snapshot_source: str | None = None,
    evento_base=None,
    evento_id=None,
    for_sensor: bool = False,
):
    camera_id = camera.get("id")
    grava_foto = plan_grava_foto(camera)
    grava_video = plan_grava_video(camera)

    if not for_sensor and not is_camera_active(camera_id):
        print(f"[CAPTURA] camera={camera_id} inativa, evento ignorado")
        _discard_snapshot(snapshot_source)
        return

    if not grava_foto and not grava_video:
        if not try_iniciar_captura(camera_id):
            print(f"[CAPTURA] camera={camera_id} captura ja em andamento, evento ignorado")
            _discard_snapshot(snapshot_source)
            return
        try:
            evento = create_evento(camera, confianca, status="capturando")
            evento_id = _evento_id(evento)
            if not evento_id:
                raise RuntimeError(f"Xano nao retornou id do evento: {evento}")
            if not for_sensor:
                submit_terminal_notify(camera, evento_id, evento)
            _finalizar_somente_evento(evento_id, evento)
        except Exception as exc:
            print(f"[ERRO] evento sem midia camera={camera_id}: {exc}")
            traceback.print_exc()
        finally:
            _discard_snapshot(snapshot_source)
            finalizar_captura(camera_id)
        return

    id_franqueado = camera.get("id_franqueado")
    rtsp = rtsp_url(camera_id)
    evento = None
    work_dir: Path | None = None
    snapshot_future = None
    t_clip = None
    clip_errors: dict[str, Exception] = {}
    evento_id_local = evento_id
    terminal_notificado = False

    if not try_iniciar_captura(camera_id):
        print(f"[CAPTURA] camera={camera_id} captura ja em andamento, evento ignorado")
        _discard_snapshot(snapshot_source)
        return

    try:
        if not for_sensor and not is_camera_active(camera_id):
            print(f"[CAPTURA] camera={camera_id} desativada antes do evento, ignorado")
            _discard_snapshot(snapshot_source)
            return

        if evento_id_local:
            evento = evento_base or {"id": evento_id_local}
            evento_id = evento_id_local
        else:
            evento = create_evento(camera, confianca, status="capturando")
            evento_id = _evento_id(evento)
        if not evento_id:
            raise RuntimeError(f"Xano nao retornou id do evento: {evento}")

        # Sem foto na licenca: abre terminal ja; com foto, notifica apos snapshot.
        if not for_sensor and not grava_foto:
            submit_terminal_notify(camera, evento_id, evento)
            terminal_notificado = True

        work_dir = evento_work_dir(evento_id)
        snapshot_path = work_dir / "snapshot.jpg"
        clip_path = work_dir / "clip_001.mp4"

        if grava_foto:
            used_detection_frame = install_detection_snapshot(snapshot_path, snapshot_source)
            if used_detection_frame:
                print(
                    f"[CAPTURA] evento id={evento_id} camera={camera_id} "
                    f"snapshot=frame_yolo bytes={snapshot_path.stat().st_size}"
                )
                snapshot_key = evento_snapshot_key(id_franqueado, evento_id)
                snapshot_future = submit_upload(snapshot_path, snapshot_key, "image/jpeg")
                if grava_video:
                    t_clip, clip_errors = start_clip_capture(
                        rtsp, clip_path, CLIP_DURACAO_SEG, camera_id=camera_id
                    )
                    print(
                        f"[CAPTURA] evento id={evento_id} gravando clip {CLIP_DURACAO_SEG}s "
                        f"(snapshot ja subindo)"
                    )
            elif grava_video:
                print(
                    f"[CAPTURA] evento id={evento_id} camera={camera_id} rtsp={rtsp} "
                    f"fallback snapshot+clip paralelo ffmpeg"
                )
                t_snapshot, t_clip, capture_errors = start_parallel_capture(
                    rtsp, snapshot_path, clip_path, CLIP_DURACAO_SEG, camera_id=camera_id
                )
                t_snapshot.join()
                if "snapshot" in capture_errors:
                    raise capture_errors["snapshot"]
                print(
                    f"[CAPTURA] snapshot local ok evento={evento_id} "
                    f"bytes={snapshot_path.stat().st_size}"
                )
                snapshot_key = evento_snapshot_key(id_franqueado, evento_id)
                snapshot_future = submit_upload(snapshot_path, snapshot_key, "image/jpeg")
                clip_errors = capture_errors
            else:
                print(
                    f"[CAPTURA] evento id={evento_id} camera={camera_id} rtsp={rtsp} "
                    f"snapshot ffmpeg (sem video)"
                )
                capture_snapshot_jpeg_file(rtsp, snapshot_path)
                snapshot_key = evento_snapshot_key(id_franqueado, evento_id)
                snapshot_future = submit_upload(snapshot_path, snapshot_key, "image/jpeg")
        elif grava_video:
            print(
                f"[CAPTURA] evento id={evento_id} camera={camera_id} rtsp={rtsp} "
                f"clip ffmpeg (sem foto)"
            )
            t_clip, clip_errors = start_clip_capture(
                rtsp, clip_path, CLIP_DURACAO_SEG, camera_id=camera_id
            )

    except Exception as exc:
        print(f"[ERRO] captura local evento camera={camera_id} evento={evento_id}: {exc}")
        traceback.print_exc()
        if t_clip is not None and t_clip.is_alive():
            t_clip.join()
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
        if not for_sensor and not is_camera_active(camera_id):
            cancel_camera_captures(camera_id)
            if t_clip is not None and t_clip.is_alive():
                t_clip.join(timeout=10)
            if evento_id:
                put_evento(evento_id, {"status": "erro", "clip_count": 0}, base=evento)
            print(f"[CAPTURA] camera={camera_id} desativada — evento={evento_id} cancelado")
            if work_dir:
                cleanup_work_dir(work_dir)
            return

        snapshot_url = None
        if grava_foto and snapshot_future:
            snapshot_url = snapshot_future.result()
            print(f"[CAPTURA] snapshot contabo ok evento={evento_id} url={snapshot_url}")
            put_evento(
                evento_id,
                {
                    "snapshot_url": snapshot_url,
                    "status": "capturando" if grava_video else "pronto",
                    "clip_count": 0,
                },
                base=evento,
            )
            if not for_sensor:
                submit_terminal_notify(camera, evento_id, evento)
                terminal_notificado = True
            if grava_video:
                print(f"[CAPTURA] evento={evento_id} snapshot visivel (capturando video)")

        if not grava_video:
            if evento_id:
                payload = {"status": "pronto", "clip_count": 0, "processado": False}
                if snapshot_url:
                    payload["snapshot_url"] = snapshot_url
                put_evento(evento_id, payload, base=evento)
                print(f"[CAPTURA] evento={evento_id} pronto snapshot (sem video)")
                if not for_sensor and not terminal_notificado:
                    submit_terminal_notify(camera, evento_id, evento)
            if work_dir:
                cleanup_work_dir(work_dir)
            return

        if not for_sensor and not is_camera_active(camera_id):
            cancel_camera_captures(camera_id)

        if t_clip is not None:
            t_clip.join()
        if "clip" in clip_errors:
            err = clip_errors["clip"]
            if isinstance(err, CaptureCancelled) or (
                not for_sensor and not is_camera_active(camera_id)
            ):
                put_evento(evento_id, {"status": "erro", "clip_count": 0}, base=evento)
                print(f"[CAPTURA] camera={camera_id} desativada — clip evento={evento_id} cancelado")
                if work_dir:
                    cleanup_work_dir(work_dir)
                return
            raise err
        if not for_sensor and not is_camera_active(camera_id):
            put_evento(evento_id, {"status": "erro", "clip_count": 0}, base=evento)
            print(f"[CAPTURA] camera={camera_id} desativada apos clip — evento={evento_id} erro")
            if work_dir:
                cleanup_work_dir(work_dir)
            return

        print(f"[CAPTURA] clip local ok evento={evento_id} bytes={clip_path.stat().st_size}")

        clip_key = evento_clip_key(id_franqueado, evento_id, 1)
        clip_future = submit_upload(clip_path, clip_key, "video/mp4")
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
        if not for_sensor and not terminal_notificado:
            submit_terminal_notify(camera, evento_id, evento)

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
