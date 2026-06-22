import time

import cv2
from ultralytics import YOLO

from config import FRAME_SKIP, YOLO_MODEL


class PersonDetector:
    def __init__(self):
        self.model = YOLO(YOLO_MODEL)

    def process_camera(self, rtsp_url, conf_min, cooldown_sec, on_person):
        cap = cv2.VideoCapture(rtsp_url, cv2.CAP_FFMPEG)
        cap.set(cv2.CAP_PROP_BUFFERSIZE, 1)

        if not cap.isOpened():
            print(f"[ERRO] Nao abriu stream: {rtsp_url}")
            return

        print(f"[OK] Stream aberto: {rtsp_url}")
        frame_index = 0
        ultimo_evento = 0.0

        while True:
            ok, frame = cap.read()
            if not ok:
                print(f"[WARN] Frame perdido, reconectando: {rtsp_url}")
                cap.release()
                time.sleep(2)
                cap = cv2.VideoCapture(rtsp_url, cv2.CAP_FFMPEG)
                continue

            frame_index += 1
            if frame_index % FRAME_SKIP != 0:
                continue

            results = self.model(frame, verbose=False)[0]
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
