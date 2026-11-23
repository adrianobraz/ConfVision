"""Uma inferencia por vez, no maximo um frame pendente por camera.

A thread de leitura nao espera o modelo. Se chegar outro frame elegivel
antes de a inferencia comecar, o anterior e descartado e fica o mais novo.
O horario de entrada na fila nao muda nessa troca, para uma camera nao
furar a frente das outras.

Frames intermediarios deixam de ser analisados enquanto houver contencao.
Sem contencao, o intervalo do FRAME_SKIP e maior que a inferencia e o
descarte nao acontece.

Nao ha fila ilimitada: o pendente cabe em um slot por camera. A instancia
do modelo nao e chamada em paralelo.
"""

from __future__ import annotations

import threading
import time
from typing import Callable, Optional


class InferRequest:
    def __init__(
        self,
        camera_id: int,
        generation: int,
        frame,
        session: dict,
        captured_at: float,
        rtsp_url: str,
        conf_min: float,
        cooldown_sec: float,
        on_person: Callable,
        zonas: list,
        modo: str,
        should_continue: Callable[[], bool],
    ):
        self.camera_id = int(camera_id)
        self.generation = int(generation)
        self.frame = frame
        self.session = session
        self.captured_at = float(captured_at)
        self.rtsp_url = rtsp_url
        self.conf_min = float(conf_min)
        self.cooldown_sec = float(cooldown_sec)
        self.on_person = on_person
        self.zonas = zonas
        self.modo = modo
        self.should_continue = should_continue
        self.gate = session["gate"]
        self.gate_lock = session["gate_lock"]
        self.submitted_at = time.perf_counter()
        self.queue_s = 0.0


class InferenceScheduler:
    def __init__(self, infer_fn: Callable, model_lock: threading.Lock, on_result: Callable):
        self._infer_fn = infer_fn
        self._lock = model_lock
        self._on_result = on_result
        self._pending: dict[int, InferRequest] = {}
        self._cv = threading.Condition()
        self._stop = False
        self._inflight = 0
        self._thread: Optional[threading.Thread] = None

    def start(self) -> None:
        if self._thread is not None and self._thread.is_alive():
            return
        self._stop = False
        self._thread = threading.Thread(
            target=self._loop,
            name="yolo-infer",
            daemon=True,
        )
        self._thread.start()

    def stop(self, timeout: float = 2.0) -> None:
        with self._cv:
            self._stop = True
            self._cv.notify_all()
        thread = self._thread
        if thread is not None and thread.is_alive() and thread is not threading.current_thread():
            thread.join(timeout)

    def pending_count(self) -> int:
        with self._cv:
            return len(self._pending)

    def inflight_count(self) -> int:
        with self._cv:
            return self._inflight

    def invalidate(self, camera_id: int) -> None:
        with self._cv:
            self._pending.pop(int(camera_id), None)

    def submit(self, job: InferRequest) -> bool:
        """Guarda o frame mais novo. Retorna True se substituiu outro ainda nao iniciado."""
        with self._cv:
            if self._stop:
                return False
            prev = self._pending.get(job.camera_id)
            replaced = prev is not None
            if replaced:
                job.submitted_at = prev.submitted_at
            self._pending[job.camera_id] = job
            self._cv.notify()
        return replaced

    def _take(self) -> Optional[InferRequest]:
        with self._cv:
            while not self._pending and not self._stop:
                self._cv.wait(timeout=0.5)
            if not self._pending:
                return None
            camera_id = min(self._pending, key=lambda cid: self._pending[cid].submitted_at)
            job = self._pending.pop(camera_id)
            self._inflight += 1
            return job

    def _finish_inflight(self) -> None:
        with self._cv:
            if self._inflight > 0:
                self._inflight -= 1

    def _loop(self) -> None:
        while True:
            job = self._take()
            if job is None:
                return
            try:
                self._run(job)
            except Exception as exc:
                print(
                    f"[YOLO] inferencia isolada camera={job.camera_id}: {exc}",
                    flush=True,
                )
            finally:
                self._finish_inflight()

    def _run(self, job: InferRequest) -> None:
        started = time.perf_counter()
        job.queue_s = max(0.0, started - job.submitted_at)
        with self._lock:
            waited = time.perf_counter() - started
            t_infer = time.perf_counter()
            results = self._infer_fn(job.frame)
            infer_s = time.perf_counter() - t_infer
        self._on_result(job, results, waited, infer_s)
