#!/usr/bin/env bash
# Fase 5 — resumo /health + /metrics Rust A e B (CT111, sem cargo)
set -euo pipefail

PILOT_A="${PILOT_A:-https://foxpro-rust-pilot.rkr351.easypanel.host}"
PILOT_B="${PILOT_B:-https://foxpro-rust-pilot-b.rkr351.easypanel.host}"

if ! command -v jq >/dev/null 2>&1; then
  echo "jq obrigatorio" >&2
  exit 1
fi

show_one() {
  local base="$1"
  local label="$2"
  base="${base%/}"
  echo ""
  echo "========== ${label} — ${base} =========="
  local health metrics
  health="$(curl -fsS --max-time 25 "${base}/health")" || {
    echo "FAIL /health"
    return 1
  }
  metrics="$(curl -fsS --max-time 25 "${base}/metrics" 2>/dev/null || echo '{}')"

  echo "$health" | jq -c '{
    status,
    processor_id,
    worker_id,
    cameras_online,
    cameras_total,
    capacity_state,
    load_advisory,
    fps_total,
    frames_received,
    max_cameras,
    decode_backend_effective,
    gpu_decode_state
  }'

  echo "$metrics" | jq -c '{
    frames_dropped: .metrics.frames_dropped,
    frames_received: .metrics.frames_received,
    errors: .metrics.errors,
    hw_fallback_to_cpu_count: .metrics.hw_fallback_to_cpu_count
  }' 2>/dev/null || echo "$metrics" | jq -c '.metrics // .' | head -c 500

  local on tot
  on="$(echo "$health" | jq -r '.cameras_online // 0')"
  tot="$(echo "$health" | jq -r '.cameras_total // 0')"
  echo "  → câmeras: online=${on} total=${tot} (fonte: /health do processor)"
}

echo "=== Fase 5 health+metrics $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="

fail=0
show_one "$PILOT_A" "Rust A" || fail=1
show_one "$PILOT_B" "Rust B" || fail=1

a_on="$(curl -fsS --max-time 15 "${PILOT_A%/}/health" | jq -r '.cameras_online // 0')"
b_on="$(curl -fsS --max-time 15 "${PILOT_B%/}/health" | jq -r '.cameras_online // 0')"
echo ""
echo "Total online A+B: $((a_on + b_on))"

if [[ "$fail" -ne 0 ]]; then
  echo "RESULT: FAIL"
  exit 1
fi
echo "RESULT: OK"
echo ""
echo "Nota: sync Go (cameras no sync) pode ser 0 mesmo com câmeras 'ativas' no painel se analitico_pausado ou worker_id/RTSP."
