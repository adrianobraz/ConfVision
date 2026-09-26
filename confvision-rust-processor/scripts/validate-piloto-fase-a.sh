#!/usr/bin/env bash
# Valida endpoints do piloto Fase A.
# Uso: validate-piloto-fase-a.sh [BASE_URL]
#      MAX_CAMERAS=10 validate-piloto-fase-a.sh https://...
# Ex.: validate-piloto-fase-a.sh https://foxpro-rust-pilot.rkr351.easypanel.host

set -euo pipefail

MAX_EXPECT="${MAX_CAMERAS:-10}"
BASE="${1:-http://127.0.0.1:8090}"
BASE="${BASE%/}"

echo "=== Fase A validation @ ${BASE} ==="

health="$(curl -fsS "${BASE}/health")" || {
  echo "FAIL: /health unreachable"
  exit 1
}

ready="$(curl -fsS "${BASE}/ready")" || {
  echo "FAIL: /ready unreachable"
  exit 1
}

metrics="$(curl -fsS "${BASE}/metrics")" || {
  echo "FAIL: /metrics unreachable"
  exit 1
}

if command -v jq >/dev/null 2>&1; then
  echo "--- /health ---"
  echo "$health" | jq '{status, processor_id, cameras_total, cameras_online, cameras_offline, frames_received, fps_total, capacity_state, load_advisory}'
  echo "--- /ready ---"
  echo "$ready" | jq .
  echo "--- /metrics (cameras) ---"
  echo "$metrics" | jq '{cameras_total, cameras_online, fps_total, load_advisory, cameras: [.cameras[]? | {camera_id, status, fps, rtsp_errors, last_error: (.last_error | if . then .[0:80] else null end)}]}'
  ct="$(echo "$health" | jq -r '.cameras_total // empty')"
  on="$(echo "$health" | jq -r '.cameras_online // empty')"
  if [[ -n "$ct" && "$ct" -gt "$MAX_EXPECT" ]]; then
    echo "WARN: cameras_total=$ct > MAX_CAMERAS=$MAX_EXPECT (truncagem no Rust ou ajuste env)"
    exit 2
  fi
  if [[ "$on" == "0" && "$ct" != "0" ]]; then
    echo "WARN: cameras_online=0 com cameras_total=$ct — RTSP/sync"
    exit 2
  fi
else
  echo "$health"
  echo "$ready"
  echo "(instale jq para checagens automáticas cameras_total=1)"
fi

echo "OK: endpoints respondem; revise WARN acima se houver."
