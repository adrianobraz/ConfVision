#!/usr/bin/env python3
"""Sidecar HTTP YOLO (D3) — health, fila limitada, workers paralelos."""
from __future__ import annotations

import base64
import json
import os
import sys
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

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


def _write_json_response(handler: BaseHTTPRequestHandler, status: int, body: bytes) -> None:
    """Evita BrokenPipeError ruidoso quando o cliente Rust cancela por timeout."""
    try:
        handler.send_response(status)
        handler.send_header("Content-Type", "application/json")
        handler.send_header("Content-Length", str(len(body)))
        handler.end_headers()
        handler.wfile.write(body)
    except (BrokenPipeError, ConnectionResetError):
        print("[sidecar] cliente fechou conexao antes da resposta (timeout?)", file=sys.stderr)


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
            _write_json_response(self, 200 if not busy else 503, body)
            return
        self.send_error(501, "use POST /v1/detect or GET /health")

    def do_POST(self):
        global _active
        if self.path.rstrip("/") != "/v1/detect":
            self.send_error(404)
            return

        with _active_lock:
            if _active >= MAX_QUEUE:
                err = json.dumps({"error": "sidecar busy"}).encode("utf-8")
                _write_json_response(self, 503, err)
                return
            _active += 1

        try:
            length = int(self.headers.get("Content-Length", "0"))
            if length <= 0:
                raise ValueError("body vazio")
            body = self.rfile.read(length)
            if len(body) != length:
                raise ValueError("Content-Length incompleto")
            with _infer_lock:
                payload = self._infer(body)
            _write_json_response(self, 200, payload)
        except ValueError as exc:
            err = json.dumps({"error": str(exc)}).encode("utf-8")
            _write_json_response(self, 400, err)
        except Exception as exc:
            err = json.dumps({"error": str(exc)}).encode("utf-8")
            _write_json_response(self, 500, err)
        finally:
            with _active_lock:
                _active -= 1

    def _infer(self, body: bytes) -> bytes:
        data = json.loads(body.decode("utf-8"))
        b64 = data.get("jpeg_base64")
        if not b64 or not str(b64).strip():
            raise ValueError("jpeg_base64 ausente")
        raw = base64.b64decode(b64, validate=True)
        if not raw:
            raise ValueError("jpeg vazio apos base64")
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


class ThreadingServer(ThreadingHTTPServer):
    """ThreadingHTTPServer já inclui ThreadingMixIn (Py3 — não herdar os dois)."""

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
