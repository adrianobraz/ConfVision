#!/usr/bin/env bash
# Relatório operacional de capacidade (Rust processor).
# Uso: capacity-report.sh [BASE_URL]
# Ex.: capacity-report.sh https://foxpro-rust-pilot.rkr351.easypanel.host

set -euo pipefail

BASE="${1:-http://127.0.0.1:8090}"
BASE="${BASE%/}"

report="$(curl -fsS "${BASE}/capacity-report")" || {
  echo "FAIL: /capacity-report unreachable at ${BASE}"
  exit 1
}

if command -v jq >/dev/null 2>&1; then
  echo "=== capacity-report @ ${BASE} ==="
  echo "$report" | jq '{
    generated_at,
    processor_id: .identity.processor_id,
    summary,
    load,
    capacity_headroom: {
      state: .capacity.state,
      estimated_capacity_cameras: .capacity.estimated_capacity_cameras,
      estimated_available_cameras: .capacity.estimated_available_cameras,
      limiting_resource: .capacity.limiting_resource,
      cpu_percent: .capacity.cpu.percent
    },
    recommended_actions,
    rtsp_404: [.cameras[]? | select(.issue == "rtsp_404") | {camera_id, status, last_error_excerpt}]
  }'
else
  echo "$report"
fi
