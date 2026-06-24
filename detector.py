import os
import time
from typing import Callable, Optional

import cv2
from ultralytics import YOLO

from area_utils import areas_ativas, bbox_foot_pct, find_area_for_point
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
        areas,
        should_continue: Optional[Callable[[], bool]] = None,
    ):
        if should_continue is None:
            should_continue = lambda: True

        zonas = areas_ativas(areas)
        if not zonas:
            print(f"[AREA] sem area ativa — deteccao ignorada: {rtsp_url}")
            return

        cap = self._open_capture(rtsp_url)

        if not cap.isOpened():
            print(f"[ERRO] Nao abriu stream RTSP: {rtsp_url}")
            time.sleep(10)
            return

        print(f"[OK] Stream aberto: {rtsp_url} areas={len(zonas)}")
        frame_index = 0
        ultimo_evento = 0.0
        falhas = 0
        estava_dentro = False

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

            h, w = frame.shape[:2]
            results = self.model(frame, device=self.device, verbose=False)[0]
            best_conf = 0.0
            best_area = None

            for box in results.boxes:
                if int(box.cls[0]) != 0:
                    continue
                conf = float(box.conf[0])
                if conf < conf_min:
                    continue
                cx, cy = bbox_foot_pct(box, w, h)
                area = find_area_for_point(cx, cy, zonas)
                if area and conf > best_conf:
                    best_conf = conf
                    best_area = area

            dentro_agora = best_conf >= conf_min and best_area is not None
            if dentro_agora and not estava_dentro:
                agora = time.time()
                if agora - ultimo_evento >= cooldown_sec:
                    ultimo_evento = agora
                    on_person(best_conf, frame.copy(), best_area)
                    area_nome = best_area.get("nome") or best_area.get("id")
                    print(
                        f"[EVENTO] entrou na area={area_nome} conf={best_conf:.2f} url={rtsp_url}"
                    )

            estava_dentro = dentro_agora

        cap.release()
