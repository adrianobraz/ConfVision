"""Fase 3: metricas, aquecimento e inferencia desacoplada. Sem RTSP e sem pesos."""

from __future__ import annotations

import sys
import threading
import time
import types
from types import SimpleNamespace

import numpy as np

ultralytics = types.ModuleType("ultralytics")
ultralytics.YOLO = object
sys.modules.setdefault("ultralytics", ultralytics)

import detection_metrics
import detector as detector_mod
import rtsp_backoff as rb
from inference_scheduler import InferRequest, InferenceScheduler


def setup_function():
    detection_metrics.stop_reporter()
    detection_metrics.configure(True, 30)
    detection_metrics.reset()
    detection_metrics.clear_queues()


def teardown_function():
    detection_metrics.stop_reporter()
    detection_metrics.configure(False, 30)
    detection_metrics.reset()
    detection_metrics.clear_queues()
    detection_metrics._fault_mono = 0.0


def _camera_line(lines, camera_id):
    prefix = f"[METRICS] camera={camera_id} "
    found = [line for line in lines if line.startswith(prefix)]
    assert found, lines
    return found[0]


def _field(line, name):
    token = name + "="
    for part in line.split():
        if part.startswith(token):
            return part[len(token):]
    raise AssertionError(f"{name} ausente em {line}")


class Box:
    def __init__(self, cls_id, conf, xyxy):
        self.cls = np.array([cls_id])
        self.conf = np.array([conf], dtype=float)
        self.xyxy = [np.array(xyxy, dtype=float)]


class Results:
    def __init__(self, boxes):
        self.boxes = boxes
        self.orig_shape = (100, 100)


def _session(generation=0, gate=None):
    return {
        "generation": generation,
        "ultimo_evento": 0.0,
        "ultimo_diag": 0.0,
        "gate": gate
        or {
            "yolo_ligado": True,
            "motion_pending_verify": False,
            "miss_streak": 0,
        },
        "gate_lock": threading.Lock(),
    }


def _person():
    person = detector_mod.PersonDetector.__new__(detector_mod.PersonDetector)
    person.device = "cpu"
    person._infer_lock = threading.Lock()
    person._scheduler = None
    person.model = lambda frame, device=None, verbose=False: [Results([])]
    return person


def _job(person, session, frame, generation=None, captured_at=0.0, conf_min=0.5, cooldown=30, modo="ambos", zonas=None, on_person=None):
    events = []

    def default_on_person(conf, snap, area=None, bbox_json=None):
        events.append((conf, snap, area, bbox_json))

    job = InferRequest(
        camera_id=7,
        generation=session["generation"] if generation is None else generation,
        frame=frame,
        session=session,
        captured_at=captured_at,
        rtsp_url="rtsp://camera.local/stream",
        conf_min=conf_min,
        cooldown_sec=cooldown,
        on_person=on_person or default_on_person,
        zonas=zonas or [],
        modo=modo,
        should_continue=lambda: True,
    )
    return job, events


def test_metricas_separam_espera_inferencia_e_intervalo():
    detection_metrics.record_yolo_timing(4, 0.020, 0.100)
    detection_metrics.record_yolo_timing(4, 0.040, 0.200)
    detection_metrics.record_read_interval(4, 0.066)
    detection_metrics.record_read_interval(4, 0.070)
    detection_metrics.record_rtsp_read(4, 0.064)
    detection_metrics.note_capture_opened(4)
    detection_metrics.note_capture_opened(8)
    detection_metrics.register_queue("infer_wait", lambda: 2)
    lines = detection_metrics.collect_lines(10)
    camera = _camera_line(lines, 4)
    assert "inferencias=2" in camera
    assert _field(camera, "yolo_ms_medio") == "150.0"
    assert _field(camera, "yolo_inference_ms_medio") == "150.0"
    assert _field(camera, "yolo_lock_wait_ms_medio") == "30.0"
    assert _field(camera, "yolo_total_ms_medio") == "180.0"
    assert _field(camera, "yolo_inference_ms_max") == "200.0"
    assert _field(camera, "cap_read_interval_max_ms") == "70.0"
    assert "fila_yolo" not in lines[0]
    assert _field(lines[0], "cameras_abertas") == "2"
    assert _field(lines[0], "inferencias_aguardando") == "2"
    detection_metrics.note_capture_closed(4)
    detection_metrics.note_capture_closed(8)


def test_falha_do_coletor_nao_propaga(monkeypatch):
    def boom(*_args, **_kwargs):
        raise RuntimeError("coletor")

    monkeypatch.setattr(detection_metrics, "_cam", boom)
    detection_metrics.record_yolo_timing(1, 0.1, 0.2)
    detection_metrics.record_read_interval(1, 0.05)
    detection_metrics.record_infer_dropped(1)
    detection_metrics.note_capture_opened(1)


def test_backoff_entra_no_gauge(monkeypatch):
    monkeypatch.setattr(rb, "RTSP_BACKOFF_ENABLED", True)
    monkeypatch.setattr(rb, "RTSP_OPEN_FAIL_MAX", 1)
    monkeypatch.setattr(rb, "RTSP_OPEN_BACKOFF_SEC", 60)
    monkeypatch.setattr(rb, "RTSP_OPEN_RETRY_SEC", 1)
    cid = 424242
    rb.after_open_failure(cid, "rtsp://test/cam/x")
    assert rb.count_in_backoff() == 1
    detection_metrics.record_rtsp_read(1, 0.01)
    line = detection_metrics.collect_lines(1)[0]
    assert _field(line, "cameras_em_backoff") == "1"
    rb.after_open_success(cid)
    assert rb.count_in_backoff() == 0


def test_aquecimento_falho_nao_deixa_excecao():
    class Bad:
        def __call__(self, *_args, **_kwargs):
            raise RuntimeError("init")

    assert detector_mod.safe_warmup(Bad(), "cpu", threading.Lock()) is False


def test_segunda_chamada_nao_repete_o_custo_inicial():
    class SlowFirst:
        def __init__(self):
            self.times = []

        def __call__(self, *_args, **_kwargs):
            started = time.perf_counter()
            time.sleep(0.06 if not self.times else 0.005)
            self.times.append(time.perf_counter() - started)
            return [Results([])]

    model = SlowFirst()
    lock = threading.Lock()
    assert detector_mod.safe_warmup(model, "cpu", lock) is True
    started = time.perf_counter()
    with lock:
        model(None, device="cpu", verbose=False)
    second = time.perf_counter() - started
    assert model.times[0] > model.times[1] * 3
    assert second < model.times[0] / 2


def test_classe_confianca_cooldown_e_area(monkeypatch):
    monkeypatch.setattr(
        detector_mod,
        "overlay_boxes_and_bbox_json",
        lambda frame, results, conf, zonas, modo: (frame, "{}"),
    )
    person = _person()
    frame = np.zeros((100, 100, 3), np.uint8)
    session = _session()
    job, events = _job(person, session, frame, captured_at=1000)
    person._handle_detection(
        job,
        Results([Box(2, 0.99, (10, 10, 40, 40)), Box(0, 0.40, (10, 10, 40, 40))]),
        0.0,
        0.01,
    )
    assert events == []

    job, events = _job(person, session, frame, captured_at=1000, cooldown=30)
    person._handle_detection(job, Results([Box(0, 0.80, (40, 40, 60, 80))]), 0.0, 0.01)
    assert len(events) == 1
    assert events[0][0] == 0.80
    assert events[0][1] is frame

    job, events = _job(person, session, frame, captured_at=1010, cooldown=30)
    person._handle_detection(job, Results([Box(0, 0.91, (40, 40, 60, 80))]), 0.0, 0.01)
    assert events == []

    polygon = {
        "ativo": True,
        "nome": "Area principal",
        "poligono_json": {
            "pontos": [
                {"x": 70, "y": 70},
                {"x": 100, "y": 70},
                {"x": 100, "y": 100},
                {"x": 70, "y": 100},
            ]
        },
    }
    dentro = _session()
    job, events = _job(
        person, dentro, frame, captured_at=2000, modo="dentro", zonas=[polygon]
    )
    person._handle_detection(job, Results([Box(0, 0.88, (0, 0, 5, 5))]), 0.0, 0.01)
    assert events == []
    job, events = _job(
        person, dentro, frame, captured_at=2000, modo="dentro", zonas=[polygon]
    )
    person._handle_detection(job, Results([Box(0, 0.88, (80, 80, 95, 95))]), 0.0, 0.01)
    assert len(events) == 1
    assert events[0][2]["nome"] == "Area principal"

    fora = _session()
    job, events = _job(person, fora, frame, captured_at=3000, modo="fora", zonas=[polygon])
    person._handle_detection(job, Results([Box(0, 0.88, (80, 80, 95, 95))]), 0.0, 0.01)
    assert events == []
    job, events = _job(person, fora, frame, captured_at=3000, modo="fora", zonas=[polygon])
    person._handle_detection(job, Results([Box(0, 0.77, (0, 0, 5, 5))]), 0.0, 0.01)
    assert len(events) == 1


def test_gate_fecha_abre_e_ignora_resultado_velho():
    person = _person()
    frame = np.zeros((48, 64, 3), np.uint8)
    gate = {"yolo_ligado": True, "motion_pending_verify": True, "miss_streak": 0}
    session = _session(gate=gate)
    job, _events = _job(person, session, frame, captured_at=10)
    person._handle_detection(job, Results([]), 0.0, 0.01)
    assert gate["yolo_ligado"] is False

    gate = {"yolo_ligado": True, "motion_pending_verify": False, "miss_streak": 0}
    session = _session(gate=gate)
    for _ in range(2):
        job, _events = _job(person, session, frame, captured_at=10)
        person._handle_detection(job, Results([]), 0.0, 0.01)
    assert gate["yolo_ligado"] is False

    gate = {"yolo_ligado": True, "motion_pending_verify": True, "miss_streak": 0}
    session = _session(gate=gate)
    job, events = _job(person, session, frame, captured_at=10)
    person._handle_detection(job, Results([Box(0, 0.9, (1, 1, 4, 4))]), 0.0, 0.01)
    assert gate["yolo_ligado"] is True
    assert gate["motion_pending_verify"] is False

    session = _session(generation=2)
    job, events = _job(person, session, frame, generation=1, captured_at=10)
    person._handle_detection(job, Results([Box(0, 0.9, (1, 1, 4, 4))]), 0.0, 0.01)
    assert events == []


def test_ultimo_frame_por_camera_nao_cresce_e_nao_mistura():
    seen = []

    def infer(frame):
        seen.append((int(frame[0, 0, 0]), int(frame[0, 0, 1])))
        return Results([])

    lock = threading.Lock()
    sched = InferenceScheduler(infer, lock, lambda *_args: None)
    session_a = _session()
    session_b = _session()

    def job_for(session, camera_id, pixel, channel):
        frame = np.zeros((8, 8, 3), np.uint8)
        frame[0, 0, channel] = pixel
        job = InferRequest(
            camera_id=camera_id,
            generation=0,
            frame=frame,
            session=session,
            captured_at=0,
            rtsp_url="rtsp://camera.local/stream",
            conf_min=0.5,
            cooldown_sec=30,
            on_person=lambda *_args: None,
            zonas=[],
            modo="ambos",
            should_continue=lambda: True,
        )
        return job

    replaced = 0
    for pixel in range(1, 51):
        if sched.submit(job_for(session_a, 1, pixel, 0)):
            replaced += 1
    assert sched.pending_count() == 1
    assert replaced == 49
    assert sched.submit(job_for(session_b, 2, 9, 1)) is False
    assert sched.pending_count() == 2
    sched.start()
    sched.stop()
    assert seen == [(50, 0), (0, 9)]
    assert sched.pending_count() == 0


def test_excecao_de_uma_camera_nao_para_a_outra():
    calls = {"n": 0}

    def infer(_frame):
        calls["n"] += 1
        if calls["n"] == 1:
            raise RuntimeError("camera ruim")
        return Results([])

    done = []
    sched = InferenceScheduler(
        infer,
        threading.Lock(),
        lambda *_args: done.append(1),
    )
    sched.start()
    for camera_id in (1, 2):
        frame = np.zeros((8, 8, 3), np.uint8)
        sched.submit(
            InferRequest(
                camera_id=camera_id,
                generation=0,
                frame=frame,
                session=_session(),
                captured_at=0,
                rtsp_url="rtsp://camera.local/stream",
                conf_min=0.5,
                cooldown_sec=30,
                on_person=lambda *_args: None,
                zonas=[],
                modo="ambos",
                should_continue=lambda: True,
            )
        )
    sched.stop()
    assert calls["n"] == 2
    assert done == [1]


def test_copia_do_frame_nao_segue_o_buffer_seguinte():
    seen = []

    def infer(frame):
        seen.append(int(frame[0, 0, 0]))
        time.sleep(0.05)
        return Results([])

    person = _person()
    person.model = lambda frame, device=None, verbose=False: [Results([])]
    sched = InferenceScheduler(person._call_model, person._infer_lock, lambda *_a: None)
    person._scheduler = sched
    person._call_model = infer
    sched._infer_fn = infer
    sched.start()
    session = _session()
    frame = np.zeros((8, 8, 3), np.uint8)
    frame[0, 0, 0] = 7
    person._dispatch_infer(
        frame,
        session,
        3,
        "rtsp://camera.local/stream",
        0.5,
        30,
        lambda *_args: None,
        [],
        "ambos",
        lambda: True,
        0.0,
    )
    frame[0, 0, 0] = 9
    sched.stop()
    assert seen == [7]


def test_frame_skip_permanece_no_ciclo(monkeypatch):
    monkeypatch.setattr(detector_mod, "YOLO_ONLY_ON_MOTION", False)
    monkeypatch.setattr(detector_mod, "FRAME_SKIP", 5)
    monkeypatch.setattr(detector_mod.time, "sleep", lambda *_a, **_k: None)
    monkeypatch.setattr(detector_mod, "wait_if_backoff", lambda *_a, **_k: False)
    monkeypatch.setattr(detector_mod, "after_open_success", lambda *_a, **_k: None)
    monkeypatch.setattr(detector_mod, "after_open_failure", lambda *_a, **_k: 0)
    person = _person()
    state = {"left": 11}

    class Cap:
        def read(self):
            if state["left"] <= 0:
                return False, None
            state["left"] -= 1
            return True, np.zeros((32, 32, 3), np.uint8)

        def isOpened(self):
            return True

        def release(self):
            return None

    monkeypatch.setattr(detector_mod.PersonDetector, "_open_capture", lambda _self, _url: Cap())
    calls = {"n": 0}

    def model(frame, device=None, verbose=False):
        calls["n"] += 1
        return [Results([])]

    person.model = model
    person.process_camera(
        "rtsp://camera.local/stream",
        0.5,
        30,
        lambda *_args: None,
        [],
        lambda: state["left"] > 0,
        "ambos",
        11,
    )
    assert calls["n"] == 2


class _PacedCap:
    def __init__(self, pace, deadline):
        self.pace = pace
        self.deadline = deadline
        self.reads = 0

    def read(self):
        remaining = self.deadline - time.monotonic()
        if remaining <= 0:
            return True, np.zeros((24, 24, 3), np.uint8)
        time.sleep(min(self.pace, remaining))
        self.reads += 1
        return True, np.zeros((24, 24, 3), np.uint8)

    def isOpened(self):
        return True

    def release(self):
        return None


def _run_pair(monkeypatch, person, pace, seconds):
    monkeypatch.setattr(detector_mod, "YOLO_ONLY_ON_MOTION", False)
    monkeypatch.setattr(detector_mod, "FRAME_SKIP", 1)
    monkeypatch.setattr(detector_mod, "wait_if_backoff", lambda *_a, **_k: False)
    monkeypatch.setattr(detector_mod, "after_open_success", lambda *_a, **_k: None)
    monkeypatch.setattr(detector_mod, "after_open_failure", lambda *_a, **_k: 0)
    deadline = time.monotonic() + seconds
    caps = [_PacedCap(pace, deadline), _PacedCap(pace, deadline)]

    def open_cap(_self, _url):
        return caps.pop(0)

    monkeypatch.setattr(detector_mod.PersonDetector, "_open_capture", open_cap)
    threads = []
    for camera_id in (21, 22):
        thread = threading.Thread(
            target=person.process_camera,
            args=(
                "rtsp://camera.local/stream",
                0.5,
                30,
                lambda *_args: None,
                [],
                lambda: time.monotonic() < deadline,
                "ambos",
                camera_id,
            ),
        )
        threads.append(thread)
        thread.start()
    for thread in threads:
        thread.join()
    lines = detection_metrics.collect_lines(seconds)
    gaps = []
    waits = []
    for camera_id in (21, 22):
        line = _camera_line(lines, camera_id)
        gaps.append(float(_field(line, "cap_read_interval_max_ms")))
        wait = _field(line, "yolo_lock_wait_ms_max")
        waits.append(0.0 if wait == "-" else float(wait))
    received = 0
    for camera_id in (21, 22):
        received += int(round(float(_field(_camera_line(lines, camera_id), "fps_recebido")) * seconds))
    return received, max(gaps), max(waits)


def test_leitura_nao_espera_o_lock_quando_duas_cameras_inferem(monkeypatch, capsys):
    pace = 0.01
    seconds = 0.45

    class Slow:
        def __call__(self, _frame, device=None, verbose=False):
            time.sleep(0.08)
            return [Results([])]

    inline = _person()
    inline.model = Slow()
    inline_reads, inline_gap, inline_wait = _run_pair(monkeypatch, inline, pace, seconds)

    detection_metrics.reset()
    decoupled = _person()
    decoupled.model = Slow()
    decoupled._scheduler = InferenceScheduler(
        decoupled._call_model,
        decoupled._infer_lock,
        decoupled._handle_detection,
    )
    decoupled._scheduler.start()
    try:
        dec_reads, dec_gap, dec_wait = _run_pair(monkeypatch, decoupled, pace, seconds)
    finally:
        decoupled._scheduler.stop()

    print(
        f"[BENCH] inline leituras={inline_reads} intervalo_max_ms={inline_gap:.1f} "
        f"espera_lock_max_ms={inline_wait:.1f}",
        flush=True,
    )
    print(
        f"[BENCH] desacoplado leituras={dec_reads} intervalo_max_ms={dec_gap:.1f} "
        f"espera_lock_max_ms={dec_wait:.1f}",
        flush=True,
    )
    assert inline_wait > 20
    assert dec_reads > inline_reads * 2
    assert dec_gap < inline_gap
