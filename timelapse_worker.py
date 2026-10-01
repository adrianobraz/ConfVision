"""
Timelapse Inteligente — worker por câmera.

Dois estados:
  TIMELAPSE  — sem movimento: captura 1 frame a cada TIMELAPSE_FRAME_INTERVALO_SEG segundos.
               Acumula TIMELAPSE_FRAMES_POR_SEGMENTO frames e monta mp4 a 1fps -> envia S3.
               Ex.: 720s x 60 frames = 12h reais -> 1 min de video.

  MOVIMENTO  — com movimento: grava vídeo normal via ffmpeg, envia chunk a cada
               MOTION_CLIP_MAX_SEC (padrão 5 min) ou quando o movimento para
               (+ post-roll de MOTION_POST_ROLL_SEC segundos). Ao parar, volta para TIMELAPSE
               iniciando um novo segmento.
"""

import os
import signal
import subprocess
import threading
import time
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Optional

import cv2

from config import (
    MOTION_ANALYSIS_WIDTH,
    MOTION_CLIP_MAX_SEC,
    MOTION_FRAME_SKIP,
    MOTION_MIN_AREA,
    MOTION_MOG2_HISTORY,
    MOTION_MOG2_THRESHOLD,
    MOTION_POST_ROLL_SEC,
    MOTION_RECONNECT_SEC,
    MOTION_RECORD_DIR,
    TIMELAPSE_FRAME_INTERVALO_SEG,
    TIMELAPSE_FRAMES_POR_SEGMENTO,
)
from dvr_segment import process_segment_file
from urls import rtsp_url_for_camera
from xano_client import ack_flush_pedido

os.environ.setdefault("OPENCV_FFMPEG_CAPTURE_OPTIONS", "rtsp_transport;tcp")

_STATE_TIMELAPSE = "timelapse"
_STATE_MOVIMENTO = "movimento"


class TimelapseCameraWorker:
    def __init__(self, camera: dict[str, Any]):
        self.camera = camera
        self.camera_id = int(camera["id"])
        self._stop = threading.Event()
        self._thread: Optional[threading.Thread] = None
        self._flush_pedido = threading.Event()

    def request_flush(self):
        """Sinaliza para o worker fazer flush imediato do segmento atual."""
        self._flush_pedido.set()
        print(f"[TIMELAPSE] camera={self.camera_id} flush pedido")

    def start(self):
        if self._thread and self._thread.is_alive():
            return
        self._stop.clear()
        self._thread = threading.Thread(
            target=self._run,
            daemon=True,
            name=f"timelapse-{self.camera_id}",
        )
        self._thread.start()

    def stop(self, timeout: float = 20):
        self._stop.set()
        if self._thread:
            self._thread.join(timeout=timeout)

    def is_alive(self) -> bool:
        return bool(self._thread and self._thread.is_alive())

    def _run(self):
        while not self._stop.is_set():
            try:
                self._process_stream()
            except Exception as exc:
                print(f"[TIMELAPSE] camera={self.camera_id} ERRO loop: {exc}")
            if not self._stop.is_set():
                self._stop.wait(MOTION_RECONNECT_SEC)

    def _open_capture(self, url: str):
        cap = cv2.VideoCapture(url, cv2.CAP_FFMPEG)
        cap.set(cv2.CAP_PROP_BUFFERSIZE, 1)
        return cap

    def _detect_motion(self, frame, fgbg, kernel) -> bool:
        height, width = frame.shape[:2]
        target_w = min(MOTION_ANALYSIS_WIDTH, width)
        target_h = max(1, int(height * target_w / width))
        small = cv2.resize(frame, (target_w, target_h))
        fgmask = fgbg.apply(small)
        _, fgmask = cv2.threshold(fgmask, 200, 255, cv2.THRESH_BINARY)
        fgmask = cv2.morphologyEx(fgmask, cv2.MORPH_OPEN, kernel)
        contours, _ = cv2.findContours(
            fgmask, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE
        )
        return any(cv2.contourArea(c) > MOTION_MIN_AREA for c in contours)

    # ------------------------------------------------------------------ #
    # Timelapse helpers                                                    #
    # ------------------------------------------------------------------ #

    def _timelapse_dir(self) -> Path:
        d = Path(MOTION_RECORD_DIR) / "timelapse" / str(self.camera_id)
        d.mkdir(parents=True, exist_ok=True)
        return d

    def _assemble_timelapse(
        self,
        frames: list[tuple[datetime, Path]],
        inicio_em: datetime,
        fim_em: datetime,
    ):
        """
        Monta mp4 a partir dos frames JPEG acumulados usando ffmpeg -framerate 2.
        Cada frame representa TIMELAPSE_FRAME_INTERVALO_SEG segundos de tempo real.
        O vídeo resultante é muito pequeno (poucos frames).
        """
        if not frames:
            return

        # lista temporária de caminhos
        paths = [str(p) for _, p in frames]
        ts = inicio_em.strftime("%Y-%m-%d_%H-%M-%S")
        out_path = self._timelapse_dir() / f"{ts}_timelapse.mp4"

        # concat file para ffmpeg — 1 segundo por frame → 1fps → 120 frames = 2 min de vídeo
        concat_file = self._timelapse_dir() / f"{ts}_concat.txt"
        with open(concat_file, "w") as f:
            for p in paths:
                f.write(f"file '{p}'\n")
                f.write("duration 1\n")

        cmd = [
            "ffmpeg",
            "-hide_banner", "-loglevel", "error",
            "-f", "concat", "-safe", "0",
            "-i", str(concat_file),
            "-c:v", "libx264",
            "-preset", "veryfast",
            "-crf", "28",
            "-pix_fmt", "yuv420p",
            "-movflags", "+faststart",
            "-y", str(out_path),
        ]

        def _task():
            try:
                subprocess.run(cmd, timeout=120, check=True)
                concat_file.unlink(missing_ok=True)
                for _, p in frames:
                    p.unlink(missing_ok=True)
                ok = process_segment_file(
                    out_path,
                    self.camera,
                    tipo="timelapse",
                    inicio_em=inicio_em,
                    fim_em=fim_em,
                )
                if ok:
                    out_path.unlink(missing_ok=True)
            except Exception as exc:
                print(
                    f"[TIMELAPSE] camera={self.camera_id} "
                    f"erro montar/enviar timelapse: {exc}"
                )
                concat_file.unlink(missing_ok=True)

        threading.Thread(
            target=_task,
            daemon=True,
            name=f"timelapse-upload-{self.camera_id}",
        ).start()

    # ------------------------------------------------------------------ #
    # Movimento helpers                                                    #
    # ------------------------------------------------------------------ #

    def _start_ffmpeg(self, rtsp_url: str):
        out_dir = Path(MOTION_RECORD_DIR) / "movimento" / str(self.camera_id)
        out_dir.mkdir(parents=True, exist_ok=True)
        clip_start = datetime.now(timezone.utc)
        ts = clip_start.strftime("%Y-%m-%d_%H-%M-%S")
        clip_path = out_dir / f"{ts}.mp4"
        cmd = [
            "ffmpeg",
            "-hide_banner", "-loglevel", "error",
            "-rtsp_transport", "tcp",
            "-i", rtsp_url,
            "-c:v", "libx264",
            "-preset", "veryfast",
            "-crf", "28",
            "-an",
            "-movflags", "+faststart",
            "-y", str(clip_path),
        ]
        proc = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        return clip_start, clip_path, proc

    def _stop_ffmpeg(self, proc: subprocess.Popen, clip_path: Path, timeout: int = 45):
        if proc.poll() is not None:
            return
        try:
            proc.send_signal(signal.SIGINT)
            proc.wait(timeout=timeout)
        except subprocess.TimeoutExpired:
            proc.kill()
            proc.communicate(timeout=5)
        except Exception as exc:
            print(f"[TIMELAPSE] camera={self.camera_id} erro parar ffmpeg: {exc}")
            try:
                proc.kill()
            except Exception:
                pass
        if not clip_path.exists() or clip_path.stat().st_size == 0:
            clip_path.unlink(missing_ok=True)

    def _upload_movimento(self, clip_path: Path, clip_start: datetime, clip_end: datetime):
        def _task():
            try:
                ok = process_segment_file(
                    clip_path,
                    self.camera,
                    tipo="movimento",
                    inicio_em=clip_start,
                    fim_em=clip_end,
                )
                if ok:
                    clip_path.unlink(missing_ok=True)
            except Exception as exc:
                print(
                    f"[TIMELAPSE] camera={self.camera_id} "
                    f"upload movimento falhou file={clip_path.name}: {exc}"
                )

        threading.Thread(
            target=_task,
            daemon=True,
            name=f"timelapse-mov-upload-{self.camera_id}",
        ).start()

    # ------------------------------------------------------------------ #
    # Loop principal                                                       #
    # ------------------------------------------------------------------ #

    def _process_stream(self):
        url = rtsp_url_for_camera(self.camera)
        cap = self._open_capture(url)
        if not cap.isOpened():
            print(
                f"[TIMELAPSE] camera={self.camera_id} falhou abrir RTSP: {url}"
            )
            time.sleep(MOTION_RECONNECT_SEC)
            return

        print(f"[TIMELAPSE] camera={self.camera_id} stream OK url={url}")

        fgbg = cv2.createBackgroundSubtractorMOG2(
            history=MOTION_MOG2_HISTORY,
            varThreshold=MOTION_MOG2_THRESHOLD,
            detectShadows=True,
        )
        kernel = cv2.getStructuringElement(cv2.MORPH_ELLIPSE, (3, 3))

        state = _STATE_TIMELAPSE

        # --- timelapse state ---
        tl_frames: list[tuple[datetime, Path]] = []
        tl_segmento_inicio: Optional[datetime] = datetime.now(timezone.utc)
        tl_ultimo_frame: float = 0.0  # timestamp do último frame capturado

        # --- movimento state ---
        mv_recording = False
        mv_clip_start: Optional[datetime] = None
        mv_clip_path: Optional[Path] = None
        mv_ffmpeg_proc: Optional[subprocess.Popen] = None
        mv_last_motion: float = 0.0

        frame_idx = 0
        falhas = 0

        def flush_timelapse(fim_em: Optional[datetime] = None):
            nonlocal tl_frames, tl_segmento_inicio
            if not tl_frames or tl_segmento_inicio is None:
                tl_frames = []
                tl_segmento_inicio = datetime.now(timezone.utc)
                return
            fim = fim_em or datetime.now(timezone.utc)
            frames_snap = tl_frames[:]
            inicio_snap = tl_segmento_inicio
            tl_frames = []
            tl_segmento_inicio = datetime.now(timezone.utc)
            self._assemble_timelapse(frames_snap, inicio_snap, fim)

        def stop_movimento(force_motion: bool = False):
            nonlocal mv_recording, mv_clip_start, mv_clip_path, mv_ffmpeg_proc, mv_last_motion
            if not mv_recording or mv_ffmpeg_proc is None:
                return
            proc = mv_ffmpeg_proc
            path = mv_clip_path
            inicio = mv_clip_start
            mv_ffmpeg_proc = None
            mv_clip_path = None
            mv_clip_start = None
            mv_recording = False
            clip_end = datetime.now(timezone.utc)
            print(
                f"[TIMELAPSE] REC STOP camera={self.camera_id} "
                f"file={path.name} dur={(clip_end - inicio).total_seconds():.0f}s"
            )
            self._stop_ffmpeg(proc, path)
            self._upload_movimento(path, inicio, clip_end)
            if force_motion:
                mv_last_motion = time.time()

        def start_movimento():
            nonlocal mv_recording, mv_clip_start, mv_clip_path, mv_ffmpeg_proc
            if mv_recording:
                return
            mv_clip_start, mv_clip_path, mv_ffmpeg_proc = self._start_ffmpeg(url)
            mv_recording = True
            print(f"[TIMELAPSE] REC START camera={self.camera_id} file={mv_clip_path.name}")

        while not self._stop.is_set():
            ok, frame = cap.read()
            if not ok:
                falhas += 1
                print(f"[TIMELAPSE] camera={self.camera_id} frame perdido ({falhas}x), reconectando")
                if mv_recording:
                    stop_movimento()
                flush_timelapse()
                cap.release()
                time.sleep(MOTION_RECONNECT_SEC)
                return

            falhas = 0
            frame_idx += 1
            now = time.time()

            motion = False
            if frame_idx % MOTION_FRAME_SKIP == 0:
                motion = self._detect_motion(frame, fgbg, kernel)
                if motion:
                    mv_last_motion = now

            # -------------------------------------------------------- #
            # Máquina de estados                                        #
            # -------------------------------------------------------- #

            # -------------------------------------------------------- #
            # Flush manual pedido via API (botão "Parar e enviar")    #
            # -------------------------------------------------------- #
            if self._flush_pedido.is_set():
                self._flush_pedido.clear()
                print(f"[TIMELAPSE] camera={self.camera_id} FLUSH MANUAL pedido")
                if state == _STATE_TIMELAPSE and tl_frames:
                    flush_timelapse(datetime.now(timezone.utc))
                    print(f"[TIMELAPSE] camera={self.camera_id} flush timelapse executado")
                elif state == _STATE_MOVIMENTO and mv_recording:
                    stop_movimento()
                    state = _STATE_TIMELAPSE
                    tl_segmento_inicio = datetime.now(timezone.utc)
                    print(f"[TIMELAPSE] camera={self.camera_id} flush movimento executado → TIMELAPSE")
                ack_flush_pedido(self.camera_id)

            if state == _STATE_TIMELAPSE:
                if motion:
                    # Transição → MOVIMENTO
                    flush_timelapse(datetime.now(timezone.utc))
                    state = _STATE_MOVIMENTO
                    print(f"[TIMELAPSE] camera={self.camera_id} → MOVIMENTO")
                    start_movimento()
                else:
                    # Captura frame timelapse se passou TIMELAPSE_FRAME_INTERVALO_SEG (12 min)
                    if (now - tl_ultimo_frame) >= TIMELAPSE_FRAME_INTERVALO_SEG:
                        ts = datetime.now(timezone.utc)
                        frame_path = (
                            self._timelapse_dir()
                            / f"{ts.strftime('%Y-%m-%d_%H-%M-%S')}_{len(tl_frames):04d}.jpg"
                        )
                        cv2.imwrite(str(frame_path), frame, [cv2.IMWRITE_JPEG_QUALITY, 75])
                        tl_frames.append((ts, frame_path))
                        tl_ultimo_frame = now
                        print(
                            f"[TIMELAPSE] camera={self.camera_id} "
                            f"frame {len(tl_frames)}/{TIMELAPSE_FRAMES_POR_SEGMENTO} capturado"
                        )

                    # Fecha segmento quando acumula TIMELAPSE_FRAMES_POR_SEGMENTO frames
                    if len(tl_frames) >= TIMELAPSE_FRAMES_POR_SEGMENTO:
                        n = len(tl_frames)
                        flush_timelapse()
                        real_min = (n * TIMELAPSE_FRAME_INTERVALO_SEG) // 60
                        print(
                            f"[TIMELAPSE] camera={self.camera_id} "
                            f"segmento timelapse fechado ({n} frames, ~{real_min} min real -> {n}s video)"
                        )

            elif state == _STATE_MOVIMENTO:
                if motion:
                    mv_last_motion = now

                # Chunk de 5 min atingido → envia e abre novo se ainda há movimento
                if mv_recording and mv_clip_start is not None:
                    elapsed_mv = now - mv_clip_start.timestamp()
                    if elapsed_mv >= MOTION_CLIP_MAX_SEC:
                        still_moving = motion or (now - mv_last_motion) < MOTION_POST_ROLL_SEC
                        stop_movimento(force_motion=still_moving)
                        if still_moving and not self._stop.is_set():
                            start_movimento()
                        elif not still_moving:
                            state = _STATE_TIMELAPSE
                            tl_segmento_inicio = datetime.now(timezone.utc)
                            print(f"[TIMELAPSE] camera={self.camera_id} → TIMELAPSE")
                    elif not motion and (now - mv_last_motion) >= MOTION_POST_ROLL_SEC:
                        # Movimento parou → post-roll expirou
                        stop_movimento()
                        state = _STATE_TIMELAPSE
                        tl_segmento_inicio = datetime.now(timezone.utc)
                        print(f"[TIMELAPSE] camera={self.camera_id} → TIMELAPSE")

        cap.release()
        if mv_recording:
            stop_movimento()
        flush_timelapse()


class TimelapseWorkerManager:
    def __init__(self):
        self._workers: dict[int, TimelapseCameraWorker] = {}
        self._lock = threading.Lock()

    def sync(self, cameras: list[dict[str, Any]]):
        active_ids = {int(c["id"]) for c in cameras}

        with self._lock:
            for camera_id, worker in list(self._workers.items()):
                if camera_id not in active_ids:
                    print(f"[TIMELAPSE] camera={camera_id} removida — parando worker")
                    worker.stop()
                    del self._workers[camera_id]

            for camera in cameras:
                camera_id = int(camera["id"])
                worker = self._workers.get(camera_id)
                if worker is None or not worker.is_alive():
                    if worker is not None:
                        worker.stop(timeout=5)
                    worker = TimelapseCameraWorker(camera)
                    self._workers[camera_id] = worker
                    worker.start()
                    print(
                        f"[TIMELAPSE] worker iniciado camera={camera_id} "
                        f"nome={camera.get('nome')}"
                    )
                elif camera.get("gravacao_flush_pedido"):
                    worker.request_flush()

    def stop_all(self):
        with self._lock:
            for worker in self._workers.values():
                worker.stop()
            self._workers.clear()

    def count(self) -> int:
        with self._lock:
            return len(self._workers)
