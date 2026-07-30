import os
import threading
import time
from typing import Callable, Optional

import cv2
from ultralytics import YOLO

from area_utils import areas_ativas, find_area_for_box, normalize_modo_deteccao
from config import FRAME_SKIP, YOLO_DEVICE, YOLO_MODEL

os.environ.setdefault("OPENCV_FFMPEG_CAPTURE_OPTIONS", "rtsp_transport;tcp")


class PersonDetector:
    def __init__(self):
        self.model = YOLO(YOLO_MODEL)
        self.device = self._resolve_device()
        self._infer_lock = threading.Lock()
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
                f"Ex.: rtsp://foxpro_confvision:8554/cam/{{hash12}}"
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
        modo_deteccao: str = "dentro",
    ):
        if should_continue is None:
            should_continue = lambda: True

        modo = normalize_modo_deteccao(modo_deteccao)
        zonas = areas_ativas(areas)
        if modo != "ambos" and not zonas:
            print(f"[AREA] sem area ativa — deteccao ignorada: {rtsp_url}")
            return

        cap = self._open_capture(rtsp_url)

        if not cap.isOpened():
            print(f"[ERRO] Nao abriu stream RTSP: {rtsp_url}")
            time.sleep(10)
            return

        print(
            f"[OK] Stream aberto: {rtsp_url} areas={len(zonas)} modo={modo}"
        )
        frame_index = 0
        ultimo_evento = 0.0
        ultimo_diag = 0.0
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

            h, w = frame.shape[:2]
            with self._infer_lock:
                results = self.model(frame, device=self.device, verbose=False)[0]
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

            dentro_agora = best_conf >= conf_min and pessoas_match > 0
            agora = time.time()

            if agora - ultimo_diag >= 15:
                ultimo_diag = agora
                if dentro_agora:
                    area_nome = (best_area or {}).get("nome") or (best_area or {}).get("id") or "-"
                    print(
                        f"[DETECT] match modo={modo} area={area_nome} conf={best_conf:.2f} "
                        f"pessoas={pessoas} url={rtsp_url}"
                    )
                elif pessoas:
                    print(
                        f"[DETECT] pessoa sem match modo={modo} pessoas={pessoas} "
                        f"conf_min={conf_min} url={rtsp_url}"
                    )
                else:
                    print(
                        f"[DETECT] nenhuma pessoa conf>={conf_min} modo={modo} url={rtsp_url}"
                    )

            if dentro_agora and (agora - ultimo_evento >= cooldown_sec):
                ultimo_evento = agora
                on_person(best_conf, frame.copy(), best_area)
                area_nome = (best_area or {}).get("nome") or (best_area or {}).get("id") or modo
                print(
                    f"[EVENTO] pessoa modo={modo} area={area_nome} "
                    f"conf={best_conf:.2f} url={rtsp_url}"
                )

        cap.release()
