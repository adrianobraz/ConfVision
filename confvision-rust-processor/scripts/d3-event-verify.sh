#!/usr/bin/env bash
# D3 — health YOLO/captura + fila Redis (requer jq)
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "Uso: $0 <health-url> [health-url-b...]"
  exit 1
fi

fail=0
for base in "$@"; do
  base="${base%/}"
  h="$(curl -fsS "${base}/health")"
  echo "=== ${base} ==="
  for key in yolo_enabled yolo_backend yolo_device capture_enabled capture_workers \
    queue_backend event_queue_redis_ok events_published events_captured; do
    val="$(echo "$h" | jq -r ".${key} // empty")"
    echo "  ${key}=${val}"
  done
  yolo="$(echo "$h" | jq -r '.yolo_enabled // false')"
  cap="$(echo "$h" | jq -r '.capture_enabled // false')"
  redis_ok="$(echo "$h" | jq -r '.event_queue_redis_ok // false')"
  if [[ "$yolo" == "true" && "$(echo "$h" | jq -r '.yolo_backend // "off"')" == "off" ]]; then
    echo "FAIL yolo_enabled mas yolo_backend=off"
    fail=1
  fi
  if [[ "$cap" == "true" && "$redis_ok" != "true" ]]; then
    echo "WARN capture_enabled com event_queue_redis_ok=false (captura precisa redis ou memory queue)"
  fi
done

if [[ "$fail" -ne 0 ]]; then
  exit 1
fi
echo "OK d3-event-verify"
