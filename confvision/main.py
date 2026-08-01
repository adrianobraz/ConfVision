import socket
import threading
import time
from pathlib import Path

from bootstrap_check import validate_config
from camera_state import is_camera_active, set_active_camera_ids
from capture_workers import start_capture_workers
from config import (
    CAPTURE_WORKERS,
    CLIP_DURACAO_SEG,
    EVENT_QUEUE_BACKEND,
    MAX_CAMERAS,
    SYNC_INTERVAL_SEC,
    WORKER_SHARD_INDEX,
    WORKER_SHARD_TOTAL,
)
from area_utils import areas_ativas, camera_deve_rodar_thread, camera_elegivel_analitico, normalize_modo_deteccao
from capture import cancel_camera_captures, write_detection_snapshot
from device_armed import is_dispositivo_armado, prefetch_armado
from detector import PersonDetector
from event_queue import EventJob, get_event_queue
from sharding import shard_label
from urls import rtsp_url_for_camera, stream_path
from xano_client import get_cameras_ativas, post_ping

_restart_ids: set[int] = set()
_thread_cfg: dict[int, str] = {}


def _camera_cfg_key(camera: dict) -> str:
    areas = camera.get("areas") or []
    area_ids = sorted(
        str(a.get("id"))
        for a in areas
        if a.get("ativo", True) and a.get("id") is not None
    )
    return "|".join(
        [
            normalize_modo_deteccao(camera.get("modo_deteccao")),
            str(camera.get("confianca_min") or ""),
            str(camera.get("cooldown_seg") or ""),
            str(bool(camera.get("somente_armado"))),
            ",".join(area_ids),
        ]
    )


def loop_camera(camera, detector: PersonDetector, event_queue):
    camera_id = camera.get("id")
    conf_min = float(camera.get("confianca_min") or 0.5)
    cooldown = int(camera.get("cooldown_seg") or 30)
    modo = normalize_modo_deteccao(camera.get("modo_deteccao"))
    url = rtsp_url_for_camera(camera)
    zonas = areas_ativas(camera.get("areas"))

    if modo != "ambos" and not zonas:
        print(f"[AREA] camera id={camera_id} sem area cadastrada — thread encerrada")
        return

    def thread_deve_continuar():
        if camera_id in _restart_ids:
            return False
        if not is_camera_active(camera_id):
            return False
        if camera.get("analitico_pausado"):
            return False
        if camera.get("somente_armado"):
            id_disp = str(camera.get("id_dispositivo") or "").strip()
            if not id_disp or not is_dispositivo_armado(id_disp):
                return False
        return True

    def on_person(conf, frame, area=None):
        if not thread_deve_continuar():
            return
        if camera.get("somente_armado"):
            # analitico_armado_*: usa dispositivo.Armado
            # - fabricante CAMERA: armado pela UI ConfVision
            # - outros: armado pela central
            id_disp = camera.get("id_dispositivo")
            if not is_dispositivo_armado(id_disp):
                print(
                    f"[ARMADO] camera={camera_id} dispositivo={id_disp} "
                    f"desarmado ou indisponivel — evento ignorado"
                )
                return
        snapshot_path = None
        try:
            snapshot_path = str(write_detection_snapshot(camera_id, frame))
            area_label = (area or {}).get("nome") or (area or {}).get("id") or modo
            print(
                f"[EVENTO] snapshot instantaneo camera={camera_id} "
                f"area={area_label} modo={modo} path={snapshot_path}"
            )
        except Exception as exc:
            print(f"[WARN] snapshot instantaneo camera={camera_id} falhou: {exc}")
        job = EventJob(
            camera=camera,
            confianca=conf,
            detected_at=time.time(),
            snapshot_path=snapshot_path,
        )
        if not event_queue.publish(job):
            print(f"[FILA] cheia — evento descartado camera={camera_id} conf={conf:.2f}")
            if snapshot_path:
                Path(snapshot_path).unlink(missing_ok=True)

    def should_continue():
        return thread_deve_continuar()

    while should_continue():
        try:
            detector.process_camera(
                url,
                conf_min,
                cooldown,
                on_person,
                zonas,
                should_continue,
                modo_deteccao=modo,
            )
        except Exception as exc:
            print(f"[ERRO] camera {camera_id}: {exc}")
            if not should_continue():
                break
            time.sleep(5)

    _restart_ids.discard(camera_id)
    print(f"[THREAD] camera id={camera_id} encerrada (desativada, config ou sem area)")


def main():
    print(
        f"[START] ConfVision worker | host={socket.gethostname()} "
        f"| clip={CLIP_DURACAO_SEG}s | {shard_label()}"
    )
    validate_config()

    event_queue = get_event_queue()
    start_capture_workers(CAPTURE_WORKERS, event_queue)

    detector = PersonDetector()
    threads: dict[int, threading.Thread] = {}

    while True:
        try:
            cameras = get_cameras_ativas()
            cameras_elegiveis = [c for c in cameras if camera_elegivel_analitico(c)]
            cameras_com_area = [c for c in cameras_elegiveis if camera_deve_rodar_thread(c)]
            ids_somente_armado = {
                str(c.get("id_dispositivo")).strip()
                for c in cameras_elegiveis
                if c.get("somente_armado") and c.get("id_dispositivo")
            }
            if ids_somente_armado:
                prefetch_armado(ids_somente_armado)
            active_ids = [c["id"] for c in cameras_com_area]
            set_active_camera_ids(active_ids)
            post_ping(
                len(cameras_com_area),
                extra={
                    "shard_index": WORKER_SHARD_INDEX if WORKER_SHARD_INDEX >= 0 else None,
                    "shard_total": WORKER_SHARD_TOTAL if WORKER_SHARD_TOTAL > 0 else None,
                    "max_cameras": MAX_CAMERAS,
                    "yolo_device": detector.device,
                    "queue_backend": EVENT_QUEUE_BACKEND,
                },
            )
            print(
                f"[SYNC] {len(cameras_com_area)} camera(s) analitico ids={active_ids} "
                f"({len(cameras_elegiveis) - len(cameras_com_area)} pausadas/desarmadas, "
                f"{len(cameras) - len(cameras_elegiveis)} ignoradas)"
            )

            active_set = set(active_ids)
            for camera_id, thread in list(threads.items()):
                if camera_id not in active_set:
                    cancel_camera_captures(camera_id)
                    _thread_cfg.pop(camera_id, None)
                    if thread.is_alive():
                        print(
                            f"[SYNC] camera id={camera_id} desativada — "
                            f"aguardando thread encerrar"
                        )
                elif not thread.is_alive():
                    del threads[camera_id]
                    _thread_cfg.pop(camera_id, None)

            for camera in cameras_com_area:
                camera_id = camera["id"]
                cfg_key = _camera_cfg_key(camera)
                thread = threads.get(camera_id)
                if thread is not None and thread.is_alive():
                    if _thread_cfg.get(camera_id) != cfg_key:
                        print(
                            f"[SYNC] camera id={camera_id} config alterada — "
                            f"reiniciando thread (modo={normalize_modo_deteccao(camera.get('modo_deteccao'))})"
                        )
                        _restart_ids.add(camera_id)
                        continue
                if thread is None or not thread.is_alive():
                    _restart_ids.discard(camera_id)
                    thread = threading.Thread(
                        target=loop_camera,
                        args=(camera, detector, event_queue),
                        daemon=True,
                        name=f"camera-{camera_id}",
                    )
                    threads[camera_id] = thread
                    _thread_cfg[camera_id] = cfg_key
                    thread.start()
                    print(
                        f"[THREAD] camera id={camera_id} "
                        f"path={stream_path(camera_id, camera.get('id_franqueado'))} "
                        f"nome={camera.get('nome')} modo={normalize_modo_deteccao(camera.get('modo_deteccao'))} "
                        f"url={rtsp_url_for_camera(camera)}"
                    )
        except Exception as exc:
            print(f"[ERRO] sync: {exc}")

        time.sleep(SYNC_INTERVAL_SEC)


if __name__ == "__main__":
    main()
