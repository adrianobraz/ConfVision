#!/usr/bin/env python3
"""Sidecar HTTP YOLO (D3) — health, fila limitada, workers paralelos."""
from __future__ import annotations

import base64
import json
import os
import sys
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from socketserver import ThreadingMixIn

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
MAX_QUEUE = int(os.getenv("YOLO_MAX_QUEUE", "8"))
INFER_SLOTS = max(1, int(os.getenv("YOLO_INFER_SLOTS", "1")))
_infer_lock = threading.Semaphore(INFER_SLOTS)
_active = 0
_active_lock = threading.Lock()
model = YOLO(MODEL)


def _cpu_count() -> int:
    try:
        return len(os.sched_getaffinity(0))
    except (AttributeError, NotImplementedError):
        return os.cpu_count() or 1


def _resolve_workers() -> int:
    raw = os.getenv("YOLO_WORKERS", "auto").strip().lower()
    cap = int(os.getenv("YOLO_WORKERS_MAX", "4"))
    if raw in ("auto", ""):
        n = _cpu_count()
    else:
        n = int(raw)
    return max(1, min(n, cap))


class Handler(BaseHTTPRequestHandler):
    def log_message(self, fmt, *args):
        print(f"[sidecar] {self.address_string()} {fmt % args}")

    def do_GET(self):
        path = self.path.rstrip("/") or "/"
        if path == "/health":
            with _active_lock:
                busy = _active >= MAX_QUEUE
                act = _active
            body = json.dumps(
                {
                    "status": "ok",
                    "active_requests": act,
                    "max_queue": MAX_QUEUE,
                    "busy": busy,
                    "infer_slots": INFER_SLOTS,
                }
            ).encode("utf-8")
            self.send_response(200 if not busy else 503)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
            return
        self.send_error(501, "use POST /v1/detect or GET /health")

    def do_POST(self):
        if self.path.rstrip("/") != "/v1/detect":
            self.send_error(404)
            return

        with _active_lock:
            if _active >= MAX_QUEUE:
                err = json.dumps({"error": "sidecar busy"}).encode("utf-8")
                self.send_response(503)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(err)))
                self.end_headers()
                self.wfile.write(err)
                return
            _active += 1

        try:
            length = int(self.headers.get("Content-Length", "0"))
            body = self.rfile.read(length)
            with _infer_lock:
                payload = self._infer(body)
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(payload)))
            self.end_headers()
            self.wfile.write(payload)
        except ValueError as exc:
            err = json.dumps({"error": str(exc)}).encode("utf-8")
            self.send_response(400)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(err)))
            self.end_headers()
            self.wfile.write(err)
        except Exception as exc:
            err = json.dumps({"error": str(exc)}).encode("utf-8")
            self.send_response(500)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(err)))
            self.end_headers()
            self.wfile.write(err)
        finally:
            with _active_lock:
                _active -= 1

    def _infer(self, body: bytes) -> bytes:
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
        return json.dumps(out).encode("utf-8")


class ThreadingServer(ThreadingMixIn, ThreadingHTTPServer):
    daemon_threads = True


def main():
    workers = _resolve_workers()
    print(
        f"[sidecar] model={MODEL} device={DEVICE} port={PORT} "
        f"workers={workers} cpus_visible={_cpu_count()} max_queue={MAX_QUEUE}"
    )
    ThreadingServer(("0.0.0.0", PORT), Handler).serve_forever()


if __name__ == "__main__":
    main()
