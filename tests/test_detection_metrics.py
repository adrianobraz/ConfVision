"""Metricas do motor, sem RTSP e sem YOLO."""

from __future__ import annotations

import os
import threading
import time

import detection_metrics
from config import WORKER_ID
from event_queue import EventJob, MemoryEventQueue
from scheduler_queue import MemorySchedulerQueue, YoloJob


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


def _camera_line(lines: list[str], camera_id: int) -> str:
    prefix = f"[METRICS] camera={camera_id} "
    found = [line for line in lines if line.startswith(prefix)]
    assert found, lines
    return found[0]


def test_flag_desligada_nao_conta_e_nao_sobe_coletor():
    detection_metrics.configure(False, 30)
    detection_metrics.record_rtsp_read(1, 0.01)
    detection_metrics.record_infer(1, 0.02)
    detection_metrics.note_event_dropped()
    assert detection_metrics.start_reporter() is False
    assert detection_metrics.reporter_running() is False
    detection_metrics.configure(True, 30)
    lines = detection_metrics.collect_lines(30)
    assert all(not line.startswith("[METRICS] camera=") for line in lines)
    assert "eventos_descartados=0" in lines[0]


def test_flag_ligada_identifica_processo_e_camera():
    detection_metrics.note_active_cameras(3)
    detection_metrics.record_rtsp_read(2, 0.02)
    detection_metrics.record_gate(2, "ocioso")
    detection_metrics.register_queue("eventos", lambda: 1)
    lines = detection_metrics.collect_lines(1)
    processo = lines[0]
    assert f"worker={WORKER_ID}" in processo
    assert f"pid={os.getpid()}" in processo
    assert "cameras_ativas=3" in processo
    assert "fila_eventos=1" in processo
    assert "fila_yolo" not in processo
    assert "jobs_yolo_descartados" not in processo
    camera = _camera_line(lines, 2)
    assert "inferencias=0" in camera
    assert "gate=ocioso" in camera
    assert "leitura_rtsp_ms_medio=20.0" in camera
    assert "decode_ms" not in camera


def test_emissao_periodica(capsys, monkeypatch):
    monkeypatch.setattr(detection_metrics, "_MIN_INTERVAL_SEC", 0.05)
    detection_metrics.configure(True, 0.05)
    detection_metrics.record_rtsp_read(9, 0.01)
    detection_metrics.record_gate(9, "ocioso")
    assert detection_metrics.start_reporter() is True
    text = ""
    for _ in range(40):
        time.sleep(0.05)
        text += capsys.readouterr().out
        if "[METRICS] processo" in text and "camera=9" in text and "ligado" in text:
            break
    detection_metrics.stop_reporter()
    assert "ligado" in text
    assert "[METRICS] processo" in text
    assert "camera=9" in text
    assert "gate=ocioso" in text


def test_contadores_em_varias_threads():
    def worker(camera_id: int):
        for _ in range(100):
            detection_metrics.record_rtsp_read(camera_id, 0.001)
            detection_metrics.record_mog2(camera_id, 0.001)

    threads = [
        threading.Thread(target=worker, args=(camera_id,))
        for camera_id in range(8)
    ]
    for thread in threads:
        thread.start()
    for thread in threads:
        thread.join()
    lines = detection_metrics.collect_lines(1)
    for camera_id in range(8):
        line = _camera_line(lines, camera_id)
        assert "fps_recebido=100.00" in line
        assert "fps_analisado=100.00" in line


def test_inferencia_media_e_p95():
    for seconds in (0.010, 0.020, 0.030, 0.040, 0.100):
        detection_metrics.record_infer(4, seconds)
    line = _camera_line(detection_metrics.collect_lines(10), 4)
    assert "inferencias=5" in line
    assert "yolo_ms_medio=40.0" in line
    assert "yolo_ms_p95=100.0" in line


def test_p95_usa_as_ultimas_256_amostras():
    for index in range(300):
        detection_metrics.record_infer(1, index)
    line = _camera_line(detection_metrics.collect_lines(30), 1)
    assert "inferencias=300" in line
    assert "yolo_ms_p95=287000.0" in line


def test_falha_do_coletor_nao_interrompe_o_passo(monkeypatch, capsys):
    def boom(camera_id):
        raise RuntimeError("coletor")

    monkeypatch.setattr(detection_metrics, "_cam", boom)

    def passo_da_camera():
        detection_metrics.record_rtsp_read(1, 0.01)
        detection_metrics.record_mog2(1, 0.002)
        detection_metrics.record_infer(1, 0.03)
        detection_metrics.record_gate(1, "acompanhando")
        return "captura-segue"

    assert passo_da_camera() == "captura-segue"
    assert "falha isolada: coletor" in capsys.readouterr().out

    def inferencia_real():
        raise RuntimeError("yolo")

    try:
        detection_metrics.record_infer(1, 0.01)
        inferencia_real()
        raised = False
    except RuntimeError as exc:
        raised = str(exc) == "yolo"
    assert raised


def test_camera_parada_nao_infere():
    detection_metrics.record_rtsp_read(7, 0.01)
    detection_metrics.record_rtsp_read(7, 0.01)
    detection_metrics.record_mog2(7, 0.002)
    detection_metrics.record_discard(7)
    detection_metrics.record_gate(7, "ocioso")
    lines = detection_metrics.collect_lines(2)
    line = _camera_line(lines, 7)
    assert "fps_recebido=1.00" in line
    assert "fps_analisado=0.50" in line
    assert "inferencias=0" in line
    assert "yolo_ms_p95=-" in line
    assert "descartados=1" in line
    assert "leitura_rtsp_ms_medio=10.0" in line
    assert "mog2_ms_medio=2.0" in line
    assert "gate=ocioso" in line
    assert "decode_ms" not in line


def test_encerramento_do_coletor():
    assert detection_metrics.start_reporter() is True
    assert detection_metrics.reporter_running() is True
    detection_metrics.stop_reporter()
    assert detection_metrics.reporter_running() is False
    time.sleep(0.2)
    assert detection_metrics.reporter_running() is False


def test_movimento_ate_evento_consome_o_marco():
    detection_metrics.mark_motion(3, 1000.0)
    detection_metrics.mark_event(3, 1000.25)
    detection_metrics.mark_event(3, 1005.0)
    line = _camera_line(detection_metrics.collect_lines(30), 3)
    assert "eventos=2" in line
    assert "movimento_ate_evento_ms=250.0" in line


def test_marco_de_movimento_sobrevive_a_janela():
    detection_metrics.mark_motion(8, 50.0)
    detection_metrics.record_rtsp_read(8, 0.001)
    detection_metrics.collect_lines(30)
    detection_metrics.mark_event(8, 50.4)
    line = _camera_line(detection_metrics.collect_lines(30), 8)
    assert "movimento_ate_evento_ms=400.0" in line


def test_fila_de_eventos_do_legacy_sem_fila_yolo():
    eventos = MemoryEventQueue(10)
    assert eventos.size() == 0
    eventos.publish(EventJob(camera={"id": 1}, confianca=0.9, detected_at=1.0))
    assert eventos.size() == 1

    detection_metrics.register_queue("eventos", eventos.size)
    detection_metrics.note_event_dropped()
    detection_metrics.note_event_processed()
    line = detection_metrics.collect_lines(30)[0]
    assert "fila_eventos=1" in line
    assert "eventos_descartados=1" in line
    assert "eventos_processados=1" in line
    assert "fila_yolo" not in line

    jobs = MemorySchedulerQueue()
    assert jobs.size() == 0
    jobs.enqueue(YoloJob.create(1, "latest", 3, 1, 0.0, "frame.jpg"))
    assert jobs.size() == 1


def test_fila_yolo_so_aparece_quando_o_scheduler_existe():
    detection_metrics.register_queue("yolo", lambda: 9)
    detection_metrics.record_jobs_dropped(2)
    detection_metrics.record_batch(0.05, 4)
    line = detection_metrics.collect_lines(30)[0]
    assert "fila_yolo=9" in line
    assert "jobs_yolo_descartados=2" in line
    assert "lotes=1" in line
    assert "lote_ms_medio=50.0" in line


def test_sem_proc_marca_indisponivel(monkeypatch):
    monkeypatch.setattr(detection_metrics, "_read_cpu_ticks", lambda: None)
    monkeypatch.setattr(detection_metrics, "_read_rss_mb", lambda: None)
    line = detection_metrics.collect_lines(30)[0]
    assert "cpu_pct=indisponivel" in line
    assert "rss_mb=indisponivel" in line
