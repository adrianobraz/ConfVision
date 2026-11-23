"""Contadores do motor de deteccao. Nao altera gate, evento nem fila.

Ligado apenas com DETECTION_METRICS=1. Com a flag desligada, as funcoes
retornam na hora.

leitura_rtsp_ms_medio e o tempo de parede de cap.read(), que inclui espera
do RTSP e o decode feito pelo FFmpeg. Nao e decode puro.

yolo_ms e a duracao da chamada model() que incluiu o frame. No modo
distributed essa chamada e um lote: o mesmo tempo entra em cada camera
do lote. Somar as cameras conta o lote de novo.

cpu_pct e rss_mb vem de /proc/self. Fora do Linux ficam indisponivel.
Falha aqui nao propaga para captura, MOG2, YOLO ou evento.
"""

from __future__ import annotations

import math
import os
import threading
import time
from collections import deque
from typing import Callable, Optional

from config import DETECTION_METRICS, DETECTION_METRICS_INTERVAL_SEC, WORKER_ID

_SAMPLE_CAP = 256
_MIN_INTERVAL_SEC = 5.0
_FAULT_INTERVAL_SEC = 30.0

_enabled = bool(DETECTION_METRICS)
_interval = max(_MIN_INTERVAL_SEC, float(DETECTION_METRICS_INTERVAL_SEC))
_lock = threading.Lock()
_cameras: dict[int, dict] = {}
_batches = 0
_batch_s = 0.0
_jobs_dropped = 0
_events_dropped = 0
_events_processed = 0
_active_cameras = 0
_queue_readers: dict[str, Callable[[], int]] = {}
_proc_prev_ticks: Optional[int] = None
_proc_prev_mono: float = 0.0
_fault_mono: float = 0.0
_reporter: Optional[threading.Thread] = None
_reporter_stop = threading.Event()


def metrics_enabled() -> bool:
    return _enabled


def configure(enabled: bool, interval_sec: float = 30) -> None:
    """Usado por teste. Producao le a flag uma vez, na importacao."""
    global _enabled, _interval
    _enabled = bool(enabled)
    _interval = max(_MIN_INTERVAL_SEC, float(interval_sec))


def reset() -> None:
    """Zera a janela. Mantem os leitores de fila registrados."""
    global _batches, _batch_s, _jobs_dropped, _events_dropped, _events_processed
    global _active_cameras, _proc_prev_ticks
    with _lock:
        _cameras.clear()
        _batches = 0
        _batch_s = 0.0
        _jobs_dropped = 0
        _events_dropped = 0
        _events_processed = 0
        _active_cameras = 0
        _proc_prev_ticks = None


def clear_queues() -> None:
    with _lock:
        _queue_readers.clear()


def register_queue(name: str, reader: Callable[[], int]) -> None:
    try:
        with _lock:
            _queue_readers[name] = reader
    except Exception as exc:
        _report_fault(exc)


def _cam(camera_id: int) -> dict:
    cid = int(camera_id)
    row = _cameras.get(cid)
    if row is None:
        row = {
            "frames_ok": 0,
            "read_s": 0.0,
            "discarded": 0,
            "mog2_n": 0,
            "mog2_s": 0.0,
            "resize_s": 0.0,
            "apply_s": 0.0,
            "threshold_s": 0.0,
            "morph_s": 0.0,
            "contour_s": 0.0,
            "motion_total_s": 0.0,
            "motion_n": 0,
            "infer_n": 0,
            "infer_s": deque(maxlen=_SAMPLE_CAP),
            "motion_at": 0.0,
            "event_ms": deque(maxlen=_SAMPLE_CAP),
            "events": 0,
            "gate": "",
            "width": 0,
            "height": 0,
        }
        _cameras[cid] = row
    return row


def _report_fault(exc: BaseException) -> None:
    global _fault_mono
    try:
        now = time.monotonic()
        if now - _fault_mono < _FAULT_INTERVAL_SEC:
            return
        _fault_mono = now
        print(f"[METRICS] falha isolada: {exc}", flush=True)
    except Exception:
        return


def record_rtsp_read(camera_id: int, seconds: float) -> None:
    """Tempo de cap.read(). Nao e uma medida de decode puro."""
    if not _enabled:
        return
    try:
        with _lock:
            row = _cam(camera_id)
            row["frames_ok"] += 1
            row["read_s"] += float(seconds)
    except Exception as exc:
        _report_fault(exc)


def record_discard(camera_id: int, count: int = 1) -> None:
    if not _enabled or count <= 0:
        return
    try:
        with _lock:
            _cam(camera_id)["discarded"] += int(count)
    except Exception as exc:
        _report_fault(exc)


def record_mog2(camera_id: int, seconds: float) -> None:
    if not _enabled:
        return
    try:
        with _lock:
            row = _cam(camera_id)
            row["mog2_n"] += 1
            row["mog2_s"] += float(seconds)
    except Exception as exc:
        _report_fault(exc)


def record_motion_stages(camera_id: int, stages: dict) -> None:
    """Tempos ja medidos dentro de frame_has_motion. Nao altera a decisao."""
    if not _enabled or not stages:
        return
    try:
        with _lock:
            row = _cam(camera_id)
            row["resize_s"] += float(stages.get("resize_s") or 0.0)
            row["apply_s"] += float(stages.get("apply_s") or 0.0)
            row["threshold_s"] += float(stages.get("threshold_s") or 0.0)
            row["morph_s"] += float(stages.get("morph_s") or 0.0)
            row["contour_s"] += float(stages.get("contour_s") or 0.0)
            row["motion_total_s"] += float(stages.get("total_s") or 0.0)
            row["motion_n"] += 1
    except Exception as exc:
        _report_fault(exc)


def record_infer(camera_id: int, seconds: float) -> None:
    if not _enabled:
        return
    try:
        with _lock:
            row = _cam(camera_id)
            row["infer_n"] += 1
            row["infer_s"].append(float(seconds))
    except Exception as exc:
        _report_fault(exc)


def record_gate(camera_id: int, state: str) -> None:
    if not _enabled:
        return
    try:
        with _lock:
            _cam(camera_id)["gate"] = str(state or "")
    except Exception as exc:
        _report_fault(exc)


def record_batch(seconds: float, size: int) -> None:
    if not _enabled or size <= 0:
        return
    global _batches, _batch_s
    try:
        with _lock:
            _batches += 1
            _batch_s += float(seconds)
    except Exception as exc:
        _report_fault(exc)


def record_jobs_dropped(count: int) -> None:
    """Jobs descartados pelo scheduler YOLO. O legacy nao chama isto."""
    if not _enabled or count <= 0:
        return
    global _jobs_dropped
    try:
        with _lock:
            _jobs_dropped += int(count)
    except Exception as exc:
        _report_fault(exc)


def note_event_dropped() -> None:
    """Fila de eventos cheia, ou job largado porque a camera ficou inativa."""
    if not _enabled:
        return
    global _events_dropped
    try:
        with _lock:
            _events_dropped += 1
    except Exception as exc:
        _report_fault(exc)


def note_event_processed() -> None:
    if not _enabled:
        return
    global _events_processed
    try:
        with _lock:
            _events_processed += 1
    except Exception as exc:
        _report_fault(exc)


def note_active_cameras(count: int) -> None:
    if not _enabled:
        return
    global _active_cameras
    try:
        with _lock:
            _active_cameras = max(0, int(count))
    except Exception as exc:
        _report_fault(exc)


def note_frame_size(camera_id: int, frame) -> bool:
    """Grava largura e altura do frame ja lido. Nao copia o frame nem abre RTSP."""
    if not _enabled:
        return False
    try:
        shape = getattr(frame, "shape", None)
        if shape is None or len(shape) < 2:
            return False
        height = int(shape[0])
        width = int(shape[1])
        if width <= 0 or height <= 0:
            return False
        with _lock:
            row = _cam(camera_id)
            row["width"] = width
            row["height"] = height
        return True
    except Exception as exc:
        _report_fault(exc)
        return False


def mark_motion(camera_id: int, at: float) -> None:
    if not _enabled:
        return
    try:
        with _lock:
            _cam(camera_id)["motion_at"] = float(at)
    except Exception as exc:
        _report_fault(exc)


def mark_event(camera_id: int, at: float) -> None:
    """Primeiro evento apos mark_motion. O marco e consumido."""
    if not _enabled:
        return
    try:
        with _lock:
            row = _cam(camera_id)
            started = float(row["motion_at"] or 0.0)
            row["events"] += 1
            if started > 0:
                row["event_ms"].append((float(at) - started) * 1000.0)
                row["motion_at"] = 0.0
    except Exception as exc:
        _report_fault(exc)


def percentile_95(samples: list[float]) -> Optional[float]:
    if not samples:
        return None
    ordered = sorted(samples)
    index = max(0, math.ceil(0.95 * len(ordered)) - 1)
    return ordered[index]


def _fmt(value: Optional[float], digits: int) -> str:
    if value is None:
        return "indisponivel"
    return f"{value:.{digits}f}"


def _read_cpu_ticks() -> Optional[int]:
    path = "/proc/self/stat"
    if not os.path.exists(path):
        return None
    try:
        raw = open(path, encoding="ascii").read()
        rest = raw.rsplit(")", 1)[1].split()
        return int(rest[11]) + int(rest[12])
    except (OSError, IndexError, ValueError):
        return None


def _read_rss_mb() -> Optional[float]:
    path = "/proc/self/status"
    if not os.path.exists(path):
        return None
    try:
        with open(path, encoding="ascii") as handle:
            for line in handle:
                if line.startswith("VmRSS:"):
                    return int(line.split()[1]) / 1024.0
    except (OSError, IndexError, ValueError):
        return None
    return None


def _sample_proc() -> tuple[Optional[float], Optional[float]]:
    global _proc_prev_ticks, _proc_prev_mono
    ticks = _read_cpu_ticks()
    rss = _read_rss_mb()
    now = time.monotonic()
    cpu = None
    if ticks is not None and _proc_prev_ticks is not None:
        elapsed = now - _proc_prev_mono
        if elapsed > 0:
            try:
                hz = os.sysconf("SC_CLK_TCK")
            except (OSError, ValueError, AttributeError):
                hz = 100
            cpu = ((ticks - _proc_prev_ticks) / float(hz)) / elapsed * 100.0
    if ticks is not None:
        _proc_prev_ticks = ticks
        _proc_prev_mono = now
    return cpu, rss


def _queue_value(reader: Callable[[], int]) -> str:
    try:
        return str(int(reader()))
    except Exception:
        return "indisponivel"


def _snapshot_globals() -> dict:
    return {
        "batches": _batches,
        "batch_s": _batch_s,
        "jobs_dropped": _jobs_dropped,
        "events_dropped": _events_dropped,
        "events_processed": _events_processed,
        "active_cameras": _active_cameras,
        "readers": dict(_queue_readers),
    }


def _reset_window_locked() -> None:
    global _batches, _batch_s, _jobs_dropped, _events_dropped, _events_processed
    kept = []
    for cid, row in _cameras.items():
        width = int(row.get("width") or 0)
        height = int(row.get("height") or 0)
        motion_at = float(row["motion_at"] or 0.0)
        if motion_at or (width > 0 and height > 0):
            kept.append((cid, motion_at, width, height))
    _cameras.clear()
    for cid, motion_at, width, height in kept:
        row = _cam(cid)
        if motion_at:
            row["motion_at"] = motion_at
        if width > 0 and height > 0:
            row["width"] = width
            row["height"] = height
    _batches = 0
    _batch_s = 0.0
    _jobs_dropped = 0
    _events_dropped = 0
    _events_processed = 0


def collect_lines(elapsed_sec: float) -> list[str]:
    """Fecha a janela, devolve as linhas e zera os contadores da janela."""
    elapsed = elapsed_sec if elapsed_sec > 0 else 0.0
    cpu, rss = _sample_proc()
    with _lock:
        cameras = {cid: dict(row) for cid, row in _cameras.items()}
        for row in cameras.values():
            row["infer_s"] = list(row["infer_s"])
            row["event_ms"] = list(row["event_ms"])
        snap = _snapshot_globals()
        _reset_window_locked()
    linhas = [_format_process(cpu, rss, snap)]
    if elapsed <= 0:
        return linhas
    for cid in sorted(cameras):
        line = _format_camera(cid, cameras[cid], elapsed)
        if line:
            linhas.append(line)
    return linhas


def _format_process(cpu, rss, snap: dict) -> str:
    readers = snap["readers"]
    eventos = _queue_value(readers["eventos"]) if "eventos" in readers else "-"
    parts = [
        "[METRICS] processo",
        f"worker={WORKER_ID}",
        f"pid={os.getpid()}",
        f"cpu_pct={_fmt(cpu, 1)}",
        f"rss_mb={_fmt(rss, 1)}",
        f"cameras_ativas={int(snap['active_cameras'])}",
        f"fila_eventos={eventos}",
        f"eventos_descartados={int(snap['events_dropped'])}",
        f"eventos_processados={int(snap['events_processed'])}",
    ]
    if "yolo" in readers or snap["batches"] or snap["jobs_dropped"]:
        yolo = _queue_value(readers["yolo"]) if "yolo" in readers else "-"
        batches = int(snap["batches"])
        lote_ms = (snap["batch_s"] / batches) * 1000.0 if batches else None
        lote_txt = _fmt(lote_ms, 1) if lote_ms is not None else "-"
        parts.append(f"fila_yolo={yolo}")
        parts.append(f"lotes={batches}")
        parts.append(f"lote_ms_medio={lote_txt}")
        parts.append(f"jobs_yolo_descartados={int(snap['jobs_dropped'])}")
    return " ".join(parts)


def _format_camera(cid: int, row: dict, elapsed: float) -> Optional[str]:
    frames_ok = int(row["frames_ok"])
    mog2_n = int(row["mog2_n"])
    infer_n = int(row["infer_n"])
    discarded = int(row["discarded"])
    events = int(row["events"])
    if frames_ok == 0 and mog2_n == 0 and infer_n == 0 and discarded == 0 and events == 0:
        return None
    rate = elapsed if elapsed > 0 else 1.0
    samples = row["infer_s"]
    avg = (sum(samples) / len(samples) * 1000.0) if samples else None
    p95 = percentile_95(samples)
    p95_ms = p95 * 1000.0 if p95 is not None else None
    ev = row["event_ms"]
    ev_ms = (sum(ev) / len(ev)) if ev else None
    read_avg = (row["read_s"] / frames_ok * 1000.0) if frames_ok else None
    mog2_avg = (row["mog2_s"] / mog2_n * 1000.0) if mog2_n else None
    motion_n = int(row.get("motion_n") or 0)

    def _stage(key: str):
        if motion_n <= 0:
            return None
        return float(row.get(key) or 0.0) / motion_n * 1000.0

    gate = row.get("gate") or "-"
    width = int(row.get("width") or 0)
    height = int(row.get("height") or 0)
    if width > 0 and height > 0:
        resolucao = f"resolucao={width}x{height} largura={width} altura={height}"
    else:
        resolucao = "resolucao=desconhecida largura=desconhecida altura=desconhecida"
    return (
        f"[METRICS] camera={cid} "
        f"{resolucao} "
        f"fps_recebido={frames_ok / rate:.2f} "
        f"fps_analisado={mog2_n / rate:.2f} "
        f"inferencias={infer_n} "
        f"yolo_ms_medio={_fmt(avg, 1) if avg is not None else '-'} "
        f"yolo_ms_p95={_fmt(p95_ms, 1) if p95_ms is not None else '-'} "
        f"leitura_rtsp_ms_medio={_fmt(read_avg, 1) if read_avg is not None else '-'} "
        f"mog2_ms_medio={_fmt(mog2_avg, 1) if mog2_avg is not None else '-'} "
        f"resize_ms_medio={_fmt(_stage('resize_s'), 1) if motion_n else '-'} "
        f"mog2_apply_ms_medio={_fmt(_stage('apply_s'), 1) if motion_n else '-'} "
        f"threshold_ms_medio={_fmt(_stage('threshold_s'), 1) if motion_n else '-'} "
        f"morphology_ms_medio={_fmt(_stage('morph_s'), 1) if motion_n else '-'} "
        f"contours_ms_medio={_fmt(_stage('contour_s'), 1) if motion_n else '-'} "
        f"motion_total_ms_medio={_fmt(_stage('motion_total_s'), 1) if motion_n else '-'} "
        f"descartados={discarded} "
        f"gate={gate} "
        f"eventos={events} "
        f"movimento_ate_evento_ms={_fmt(ev_ms, 1) if ev_ms is not None else '-'}"
    )


def _emit(elapsed: float) -> None:
    try:
        for line in collect_lines(elapsed):
            print(line, flush=True)
    except Exception as exc:
        print(f"[METRICS] erro ao emitir: {exc}", flush=True)


def start_reporter() -> bool:
    global _reporter
    if not _enabled:
        return False
    if _reporter is not None and _reporter.is_alive():
        return True
    _reporter_stop.clear()
    _reporter = threading.Thread(
        target=_reporter_loop,
        name="detection-metrics",
        daemon=True,
    )
    _reporter.start()
    print(
        f"[METRICS] ligado intervalo={_interval:.0f}s worker={WORKER_ID} pid={os.getpid()}",
        flush=True,
    )
    return True


def stop_reporter(timeout: float = 2.0) -> None:
    global _reporter
    _reporter_stop.set()
    thread = _reporter
    if thread is not None and thread.is_alive() and thread is not threading.current_thread():
        thread.join(timeout)
    if thread is None or not thread.is_alive():
        _reporter = None


def reporter_running() -> bool:
    return _reporter is not None and _reporter.is_alive()


def start_if_enabled(event_queue=None, yolo_queue=None) -> bool:
    if not _enabled:
        return False
    if event_queue is not None and hasattr(event_queue, "size"):
        register_queue("eventos", event_queue.size)
    if yolo_queue is not None and hasattr(yolo_queue, "size"):
        register_queue("yolo", yolo_queue.size)
    return start_reporter()


def _emit_gauge() -> None:
    """Linha do processo sem fechar a janela das cameras."""
    try:
        cpu, rss = _sample_proc()
        with _lock:
            snap = _snapshot_globals()
        print(_format_process(cpu, rss, snap), flush=True)
    except Exception as exc:
        print(f"[METRICS] erro ao emitir: {exc}", flush=True)


def _reporter_loop() -> None:
    _emit_gauge()
    started = time.monotonic()
    while not _reporter_stop.wait(_interval):
        now = time.monotonic()
        elapsed = now - started
        started = now
        _emit(elapsed)
