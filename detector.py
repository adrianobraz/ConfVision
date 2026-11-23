import os
import threading
import time
from typing import Callable, Optional

import cv2
import numpy as np
from ultralytics import YOLO

from area_utils import areas_ativas, find_area_for_box, normalize_modo_deteccao
from detection_overlay import overlay_boxes_and_bbox_json
from config import (
    FRAME_SKIP,
    YOLO_DEVICE,
    YOLO_INFER_DECOUPLED,
    YOLO_MODEL,
    YOLO_MOTION_FRAME_SKIP,
    YOLO_MOTION_MISS_FRAMES,
    YOLO_ONLY_ON_MOTION,
)
from detection_metrics import (
    mark_event,
    mark_motion,
    metrics_enabled,
    note_capture_closed,
    note_capture_opened,
    note_frame_size,
    record_discard,
    record_gate,
    record_infer_dropped,
    record_mog2,
    record_motion_stages,
    record_read_interval,
    record_rtsp_read,
    record_yolo_timing,
    register_queue,
)
from inference_scheduler import InferRequest, InferenceScheduler
from motion_detect import create_motion_detector, frame_has_motion
from rtsp_backoff import after_open_failure, after_open_success, wait_if_backoff

os.environ.setdefault("OPENCV_FFMPEG_CAPTURE_OPTIONS", "rtsp_transport;tcp")


def warmup_model(model, device: str, infer_lock: threading.Lock) -> float:
    """Uma inferencia em frame sintetico. Nao abre RTSP e nao gera evento."""
    frame = np.zeros((640, 640, 3), dtype=np.uint8)
    started = time.perf_counter()
    with infer_lock:
        model(frame, device=device, verbose=False)
    return time.perf_counter() - started


def safe_warmup(model, device: str, infer_lock: threading.Lock) -> bool:
    try:
        elapsed = warmup_model(model, device, infer_lock)
    except Exception as exc:
        print(
            f"[YOLO] aquecimento falhou — worker segue sem aquecer: {exc}",
            flush=True,
        )
        return False
    print(f"[YOLO] aquecimento ok duracao_ms={elapsed * 1000:.0f}", flush=True)
    return True


class PersonDetector:
    def __init__(self):
        self.model = YOLO(YOLO_MODEL)
        self.device = self._resolve_device()
        self._infer_lock = threading.Lock()
        self._scheduler = None
        gate = "sim" if YOLO_ONLY_ON_MOTION else "nao"
        print(
            f"[YOLO] model={YOLO_MODEL} device={self.device} "
            f"motion_gate={gate} latch=sim miss={YOLO_MOTION_MISS_FRAMES} "
            f"infer_desacoplada={'sim' if YOLO_INFER_DECOUPLED else 'nao'}"
        )
        if YOLO_INFER_DECOUPLED:
            self._scheduler = InferenceScheduler(
                self._call_model,
                self._infer_lock,
                self._handle_detection,
            )
            self._scheduler.start()
            register_queue("infer_wait", self._scheduler.pending_count)
        safe_warmup(self.model, self.device, self._infer_lock)

    def _resolve_device(self) -> str:
        if YOLO_DEVICE:
            return YOLO_DEVICE
        try:
            import torch

            if torch.cuda.is_available():
                name = torch.cuda.get_device_name(0)
                print(f"[YOLO] GPU detectada: {name}")
                return "cuda:0"
        except Exception:
            pass
        return "cpu"

    def _call_model(self, frame):
        return self.model(frame, device=self.device, verbose=False)[0]

    def _open_capture(self, rtsp_url):
        if rtsp_url.startswith("rtmp://"):
            raise ValueError(
                f"Use RTSP para leitura, nao RTMP: {rtsp_url}. "
                f"Ex.: rtsp://foxpro_confvision:8554/cam/{{hash12}}"
            )
        cap = cv2.VideoCapture(rtsp_url, cv2.CAP_FFMPEG)
        cap.set(cv2.CAP_PROP_BUFFERSIZE, 1)
        return cap

    def _reset_motion_gate(self):
        return {
            "yolo_ligado": False,
            "motion_pending_verify": False,
            "miss_streak": 0,
        }

    def _handle_detection(self, job: InferRequest, results, lock_wait_s: float, infer_s: float):
        camera_id = job.camera_id
        try:
            record_yolo_timing(
                camera_id,
                lock_wait_s,
                infer_s,
                getattr(job, "queue_s", 0.0),
            )
        except Exception:
            pass
        try:
            session = job.session
            if session["generation"] != job.generation:
                return
            if not job.should_continue():
                return
            frame = job.frame
            h, w = frame.shape[:2]
            best_conf = 0.0
            best_area = None
            pessoas = 0
            pessoas_match = 0
            boxes = getattr(results, "boxes", None) or []
            modo = job.modo
            zonas = job.zonas
            conf_min = job.conf_min

            for box in boxes:
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

            dentro_agora = best_conf >= conf_min and pessoas_match > 0
            with job.gate_lock:
                gate = job.gate
                if YOLO_ONLY_ON_MOTION:
                    if dentro_agora:
                        gate["motion_pending_verify"] = False
                        gate["miss_streak"] = 0
                    elif gate["motion_pending_verify"]:
                        gate["yolo_ligado"] = False
                        gate["motion_pending_verify"] = False
                        gate["miss_streak"] = 0
                    else:
                        gate["miss_streak"] += 1
                        if gate["miss_streak"] >= max(1, YOLO_MOTION_MISS_FRAMES):
                            gate["yolo_ligado"] = False
                            gate["miss_streak"] = 0
                gate_open = bool(gate["yolo_ligado"])

            agora = job.captured_at
            if agora - session["ultimo_diag"] >= 15:
                session["ultimo_diag"] = agora
                if dentro_agora:
                    area_nome = (best_area or {}).get("nome") or (best_area or {}).get("id") or "-"
                    print(
                        f"[DETECT] match modo={modo} area={area_nome} conf={best_conf:.2f} "
                        f"pessoas={pessoas} url={job.rtsp_url}"
                    )
                elif pessoas:
                    print(
                        f"[DETECT] pessoa sem match modo={modo} pessoas={pessoas} "
                        f"conf_min={conf_min} url={job.rtsp_url}"
                    )
                elif gate_open or not YOLO_ONLY_ON_MOTION:
                    print(
                        f"[DETECT] nenhuma pessoa conf>={conf_min} modo={modo} url={job.rtsp_url}"
                    )
                elif YOLO_ONLY_ON_MOTION:
                    print(
                        f"[DETECT] movimento sem pessoa — gate off modo={modo} url={job.rtsp_url}"
                    )

            if not dentro_agora or (agora - session["ultimo_evento"] < job.cooldown_sec):
                return
            if session["generation"] != job.generation or not job.should_continue():
                return
            session["ultimo_evento"] = agora
            frame_snapshot, bbox_json = overlay_boxes_and_bbox_json(
                frame, results, conf_min, zonas, modo
            )
            if metrics_enabled():
                mark_event(camera_id, time.time())
            job.on_person(best_conf, frame_snapshot, best_area, bbox_json)
            area_nome = (best_area or {}).get("nome") or (best_area or {}).get("id") or modo
            print(
                f"[EVENTO] pessoa modo={modo} area={area_nome} "
                f"conf={best_conf:.2f} url={job.rtsp_url}"
            )
        except Exception as exc:
            print(f"[YOLO] resultado isolado camera={camera_id}: {exc}", flush=True)

    def _dispatch_infer(
        self,
        frame,
        session: dict,
        camera_id: int,
        rtsp_url: str,
        conf_min: float,
        cooldown_sec: float,
        on_person,
        zonas,
        modo: str,
        should_continue: Callable[[], bool],
        captured_at: float,
    ):
        scheduler = getattr(self, "_scheduler", None)
        if scheduler is not None:
            job = InferRequest(
                camera_id=camera_id,
                generation=session["generation"],
                frame=frame.copy(),
                session=session,
                captured_at=captured_at,
                rtsp_url=rtsp_url,
                conf_min=conf_min,
                cooldown_sec=cooldown_sec,
                on_person=on_person,
                zonas=zonas,
                modo=modo,
                should_continue=should_continue,
            )
            if scheduler.submit(job):
                record_infer_dropped(camera_id)
            return
        started = time.perf_counter()
        with self._infer_lock:
            waited = time.perf_counter() - started
            t_infer = time.perf_counter()
            results = self._call_model(frame)
            infer_s = time.perf_counter() - t_infer
        job = InferRequest(
            camera_id=camera_id,
            generation=session["generation"],
            frame=frame,
            session=session,
            captured_at=captured_at,
            rtsp_url=rtsp_url,
            conf_min=conf_min,
            cooldown_sec=cooldown_sec,
            on_person=on_person,
            zonas=zonas,
            modo=modo,
            should_continue=should_continue,
        )
        self._handle_detection(job, results, waited, infer_s)

    def process_camera(
        self,
        rtsp_url,
        conf_min,
        cooldown_sec,
        on_person,
        areas,
        should_continue: Optional[Callable[[], bool]] = None,
        modo_deteccao: str = "dentro",
        camera_id: int = 0,
    ):
        if should_continue is None:
            should_continue = lambda: True

        modo = normalize_modo_deteccao(modo_deteccao)
        zonas = areas_ativas(areas)
        if modo != "ambos" and not zonas:
            print(f"[AREA] sem area ativa — deteccao ignorada: {rtsp_url}")
            return

        if wait_if_backoff(camera_id, rtsp_url, should_continue):
            return

        cap = self._open_capture(rtsp_url)
        session = None
        opened = False

        try:
            if not cap.isOpened():
                try:
                    cap.release()
                except Exception:
                    pass
                cap = None
                time.sleep(after_open_failure(camera_id, rtsp_url))
                return

            after_open_success(camera_id, rtsp_url)
            opened = True
            note_capture_opened(camera_id)

            motion_fgbg = None
            motion_kernel = None
            gate = self._reset_motion_gate()
            motion_buffers = {}
            if YOLO_ONLY_ON_MOTION:
                motion_fgbg, motion_kernel = create_motion_detector()
            session = {
                "generation": 0,
                "ultimo_evento": 0.0,
                "ultimo_diag": 0.0,
                "gate": gate,
                "gate_lock": threading.Lock(),
            }

            print(
                f"[OK] Stream aberto: {rtsp_url} areas={len(zonas)} modo={modo} "
                f"motion_gate={'sim' if YOLO_ONLY_ON_MOTION else 'nao'}"
            )
            frame_index = 0
            falhas = 0
            measure = metrics_enabled()
            size_noted = False
            last_read_mono = None

            def _observe_gate():
                if not measure:
                    return
                if not YOLO_ONLY_ON_MOTION:
                    record_gate(camera_id, "desligado")
                    return
                with session["gate_lock"]:
                    gate_now = session["gate"]
                    if not gate_now["yolo_ligado"]:
                        state = "ocioso"
                    elif gate_now["motion_pending_verify"]:
                        state = "verificando"
                    else:
                        state = "acompanhando"
                record_gate(camera_id, state)

            while should_continue():
                if measure:
                    t_read = time.perf_counter()
                ok, frame = cap.read()
                if measure and ok:
                    read_done = time.perf_counter()
                    record_rtsp_read(camera_id, read_done - t_read)
                    now_mono = time.monotonic()
                    if last_read_mono is not None:
                        record_read_interval(camera_id, now_mono - last_read_mono)
                    last_read_mono = now_mono
                    if not size_noted:
                        size_noted = note_frame_size(camera_id, frame)
                if not should_continue():
                    break
                if not ok:
                    falhas += 1
                    print(f"[WARN] Frame perdido ({falhas}x), reconectando: {rtsp_url}")
                    session["generation"] += 1
                    scheduler = getattr(self, "_scheduler", None)
                    if scheduler is not None:
                        scheduler.invalidate(camera_id)
                    try:
                        cap.release()
                    except Exception:
                        pass
                    if not should_continue():
                        cap = None
                        break
                    time.sleep(min(30, 2 * falhas))
                    cap = self._open_capture(rtsp_url)
                    if not cap.isOpened():
                        time.sleep(after_open_failure(camera_id, rtsp_url))
                        return
                    after_open_success(camera_id, rtsp_url)
                    size_noted = False
                    last_read_mono = None
                    with session["gate_lock"]:
                        session["gate"] = self._reset_motion_gate()
                    if YOLO_ONLY_ON_MOTION:
                        motion_fgbg, motion_kernel = create_motion_detector()
                        motion_buffers = {}
                    continue

                falhas = 0
                frame_index += 1
                agora = time.time()
                ran_mog2 = False

                if YOLO_ONLY_ON_MOTION:
                    if frame_index % max(1, YOLO_MOTION_FRAME_SKIP) == 0:
                        ran_mog2 = True
                        timings = {} if measure else None
                        if measure:
                            t_mog2 = time.perf_counter()
                        moved = frame_has_motion(
                            frame,
                            motion_fgbg,
                            motion_kernel,
                            buffers=motion_buffers,
                            timings=timings,
                        )
                        if measure:
                            record_mog2(camera_id, time.perf_counter() - t_mog2)
                            record_motion_stages(camera_id, timings)
                        if moved:
                            with session["gate_lock"]:
                                gate_now = session["gate"]
                                if not gate_now["yolo_ligado"]:
                                    gate_now["yolo_ligado"] = True
                                    gate_now["motion_pending_verify"] = True
                                    gate_now["miss_streak"] = 0
                                    if measure:
                                        mark_motion(camera_id, agora)
                    with session["gate_lock"]:
                        gate_open = bool(session["gate"]["yolo_ligado"])
                    if not gate_open:
                        if measure and not ran_mog2:
                            record_discard(camera_id)
                        _observe_gate()
                        continue

                if frame_index % FRAME_SKIP != 0:
                    if measure and not ran_mog2:
                        record_discard(camera_id)
                    _observe_gate()
                    continue

                self._dispatch_infer(
                    frame,
                    session,
                    int(camera_id),
                    rtsp_url,
                    conf_min,
                    cooldown_sec,
                    on_person,
                    zonas,
                    modo,
                    should_continue,
                    agora,
                )
                _observe_gate()
        finally:
            if session is not None:
                session["generation"] += 1
                scheduler = getattr(self, "_scheduler", None)
                if scheduler is not None:
                    scheduler.invalidate(camera_id)
            if opened:
                note_capture_closed(camera_id)
            if cap is not None:
                try:
                    cap.release()
                except Exception:
                    pass
