#!/usr/bin/env python3
"""Sidecar HTTP mínimo para YOLO (D3 YOLO_BACKEND=http)."""
from __future__ import annotations

import base64
import json
import os
import sys
from http.server import BaseHTTPRequestHandler, HTTPServer

try:
    import cv2
    import numpy as np
    from ultralytics import YOLO
except ImportError:
    print("Instale: pip install ultralytics opencv-python-headless", file=sys.stderr)
    raise

MODEL = os.getenv("YOLO_MODEL", "yolov8n.pt")
DEVICE = os.getenv("YOLO_DEVICE", "cpu")
PORT = int(os.getenv("YOLO_HTTP_PORT", "8091"))
model = YOLO(MODEL)


class Handler(BaseHTTPRequestHandler):
    def log_message(self, fmt, *args):
        print(f"[sidecar] {self.address_string()} {fmt % args}")

    def do_POST(self):
        if self.path.rstrip("/") != "/v1/detect":
            self.send_error(404)
            return
        length = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(length)
        try:
            data = json.loads(body.decode("utf-8"))
            raw = base64.b64decode(data["jpeg_base64"])
            arr = np.frombuffer(raw, dtype=np.uint8)
            frame = cv2.imdecode(arr, cv2.IMREAD_COLOR)
            if frame is None:
                raise ValueError("jpeg inválido")
            results = model(frame, device=DEVICE, verbose=False)[0]
            dets = []
            for box in results.boxes:
                if int(box.cls[0]) != 0:
                    continue
                x1, y1, x2, y2 = box.xyxy[0].tolist()
                dets.append(
                    {
                        "class_id": 0,
                        "confidence": float(box.conf[0]),
                        "xyxy": [x1, y1, x2, y2],
                    }
                )
            out = {"width": frame.shape[1], "height": frame.shape[0], "detections": dets}
            payload = json.dumps(out).encode("utf-8")
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(payload)))
            self.end_headers()
            self.wfile.write(payload)
        except Exception as exc:
            err = json.dumps({"error": str(exc)}).encode("utf-8")
            self.send_response(400)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(err)))
            self.end_headers()
            self.wfile.write(err)


def main():
    print(f"[sidecar] model={MODEL} device={DEVICE} port={PORT}")
    HTTPServer(("0.0.0.0", PORT), Handler).serve_forever()


if __name__ == "__main__":
    main()
