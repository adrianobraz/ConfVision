import os
import time
from typing import Callable, Optional

import cv2
from ultralytics import YOLO

from config import FRAME_SKIP, YOLO_DEVICE, YOLO_MODEL

os.environ.setdefault("OPENCV_FFMPEG_CAPTURE_OPTIONS", "rtsp_transport;tcp")


class PersonDetector:
    def __init__(self):
        self.model = YOLO(YOLO_MODEL)
        self.device = self._resolve_device()
        print(f"[YOLO] model={YOLO_MODEL} device={self.device}")

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

    def _open_capture(self, rtsp_url):
        if rtsp_url.startswith("rtmp://"):
            raise ValueError(
                f"Use RTSP para leitura, nao RTMP: {rtsp_url}. "
                f"Ex.: rtsp://foxpro_confvision:8554/live/1"
            )
        cap = cv2.VideoCapture(rtsp_url, cv2.CAP_FFMPEG)
        cap.set(cv2.CAP_PROP_BUFFERSIZE, 1)
        return cap

    def process_camera(
        self,
        rtsp_url,
        conf_min,
        cooldown_sec,
        on_person,
        should_continue: Optional[Callable[[], bool]] = None,
    ):
        if should_continue is None:
            should_continue = lambda: True

        cap = self._open_capture(rtsp_url)

        if not cap.isOpened():
            print(f"[ERRO] Nao abriu stream RTSP: {rtsp_url}")
            time.sleep(10)
            return

        print(f"[OK] Stream aberto: {rtsp_url}")
        frame_index = 0
        ultimo_evento = 0.0
        falhas = 0

        while should_continue():
            ok, frame = cap.read()
            if not should_continue():
                break
            if not ok:
                falhas += 1
                print(f"[WARN] Frame perdido ({falhas}x), reconectando: {rtsp_url}")
                cap.release()
                if not should_continue():
                    break
                time.sleep(min(30, 2 * falhas))
                cap = self._open_capture(rtsp_url)
                if not cap.isOpened():
                    print(f"[ERRO] Reconexao falhou: {rtsp_url}")
                    time.sleep(10)
                    return
                continue

            falhas = 0

            frame_index += 1
            if frame_index % FRAME_SKIP != 0:
                continue

            results = self.model(frame, device=self.device, verbose=False)[0]
            best_conf = 0.0

            for box in results.boxes:
                if int(box.cls[0]) != 0:
                    continue
                conf = float(box.conf[0])
                if conf > best_conf:
                    best_conf = conf

            if best_conf >= conf_min:
                agora = time.time()
                if agora - ultimo_evento >= cooldown_sec:
                    ultimo_evento = agora
                    on_person(best_conf)
                    print(f"[EVENTO] pessoa conf={best_conf:.2f} url={rtsp_url}")

        cap.release()
