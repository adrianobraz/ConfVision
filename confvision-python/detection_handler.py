"""Pos-processamento YOLO: area, evento, metricas MOTION_TO_EVENT_MS."""

from __future__ import annotations

import time
from pathlib import Path
from typing import Any, Callable, Optional

from area_utils import areas_ativas, find_area_for_box, normalize_modo_deteccao
from capture import write_detection_snapshot
from device_armed import is_dispositivo_armado
from event_queue import EventJob


def evaluate_detections(
    results,
    conf_min: float,
    zonas: list[dict],
    modo: str,
) -> tuple[float, Optional[dict], int, int]:
    """Retorna (best_conf, best_area, pessoas_total, pessoas_match)."""
    h, w = results.orig_shape
    best_conf = 0.0
    best_area = None
    pessoas = 0
    pessoas_match = 0

    for box in results.boxes:
        if int(box.cls[0]) != 0:
            continue
        conf = float(box.conf[0])
        if conf < conf_min:
            continue
        pessoas += 1
        area = find_area_for_box(box, w, h, zonas) if zonas else None

        bate = False
        if modo == "ambos":
            bate = True
        elif modo == "fora":
            bate = area is None
        else:
            bate = area is not None

        if not bate:
            continue
        pessoas_match += 1
        if conf > best_conf:
            best_conf = conf
            best_area = area

    return best_conf, best_area, pessoas, pessoas_match


class DetectionHandler:
    def __init__(self, event_queue, frame_store, scheduler_queue, get_camera: Callable[[int], Optional[dict]]):
        self.event_queue = event_queue
        self.frame_store = frame_store
        self.scheduler_queue = scheduler_queue
        self.get_camera = get_camera
        self._cooldowns: dict[int, float] = {}

    def _should_emit_event(self, camera: dict, camera_id: int, cooldown_sec: int) -> bool:
        if camera.get("somente_armado"):
            id_disp = str(camera.get("id_dispositivo") or "").strip()
            if not id_disp or not is_dispositivo_armado(id_disp):
                return False
        agora = time.time()
        ultimo = self._cooldowns.get(camera_id, 0.0)
        if agora - ultimo < cooldown_sec:
            return False
        self._cooldowns[camera_id] = agora
        return True

    def handle_match(
        self,
        camera: dict,
        conf: float,
        frame,
        area: Optional[dict],
        motion_at: float,
        slot: str,
        session_id: int,
    ) -> None:
        camera_id = int(camera.get("id") or 0)
        cooldown = int(camera.get("cooldown_seg") or 30)
        modo = normalize_modo_deteccao(camera.get("modo_deteccao"))

        if slot.startswith("ev_"):
            self.frame_store.mark_person_confirmed(camera_id, session_id)
            self.scheduler_queue.cancel_camera_session(camera_id, session_id)

        if not self._should_emit_event(camera, camera_id, cooldown):
            return

        motion_to_event_ms = (time.time() - motion_at) * 1000 if motion_at > 0 else 0.0
        snapshot_path = None
        try:
            snapshot_path = str(write_detection_snapshot(camera_id, frame))
        except Exception as exc:
            print(f"[WARN] snapshot camera={camera_id} falhou: {exc}")

        job = EventJob(
            camera=camera,
            confianca=conf,
            detected_at=time.time(),
            snapshot_path=snapshot_path,
        )
        if not self.event_queue.publish(job):
            print(f"[FILA] cheia — evento descartado camera={camera_id} conf={conf:.2f}")
            if snapshot_path:
                Path(snapshot_path).unlink(missing_ok=True)
            return

        area_label = (area or {}).get("nome") or (area or {}).get("id") or modo
        print(
            f"[EVENTO] camera={camera_id} area={area_label} conf={conf:.2f} "
            f"slot={slot} MOTION_TO_EVENT_MS={motion_to_event_ms:.0f} modo={modo}"
        )

    def process_result(
        self,
        job,
        results,
        frame,
    ) -> None:
        camera = self.get_camera(job.camera_id)
        if not camera:
            return

        conf_min = float(camera.get("confianca_min") or 0.5)
        modo = normalize_modo_deteccao(camera.get("modo_deteccao"))
        zonas = areas_ativas(camera.get("areas"))
        best_conf, best_area, _pessoas, pessoas_match = evaluate_detections(
            results, conf_min, zonas, modo
        )

        if job.slot.startswith("ev_"):
            self.frame_store.mark_evidence_processed(job.camera_id, int(job.slot.split("_")[1]), job.session_id)

        if best_conf >= conf_min and pessoas_match > 0:
            self.handle_match(
                camera,
                best_conf,
                frame,
                best_area,
                job.motion_at,
                job.slot,
                job.session_id,
            )
