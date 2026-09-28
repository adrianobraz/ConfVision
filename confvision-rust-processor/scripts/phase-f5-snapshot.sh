#!/usr/bin/env bash
# Fase 5 — uma linha CSV por processor (health + capacity-report + metrics)
# Uso:
#   phase-f5-snapshot.sh                    # A + B (env PILOT_A, PILOT_B)
#   phase-f5-snapshot.sh URL                # um processor
#   F5_STAGE=10 phase-f5-snapshot.sh >> ramp.csv
set -euo pipefail

PILOT_A="${PILOT_A:-https://foxpro-rust-pilot.rkr351.easypanel.host}"
PILOT_B="${PILOT_B:-https://foxpro-rust-pilot-b.rkr351.easypanel.host}"
STAGE="${F5_STAGE:-}"
NOTE="${F5_NOTE:-}"

if ! command -v jq >/dev/null 2>&1; then
  echo "jq obrigatorio" >&2
  exit 1
fi

snapshot_one() {
  local base="$1"
  base="${base%/}"
  local ts health report metrics
  ts="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  health="$(curl -fsS --max-time 20 "${base}/health" 2>/dev/null || echo '{}')"
  report="$(curl -fsS --max-time 20 "${base}/capacity-report" 2>/dev/null || echo '{}')"
  metrics="$(curl -fsS --max-time 20 "${base}/metrics" 2>/dev/null || echo '{}')"

  local proc wid on tot cap adv fps fr drop cpu mem gpu vram est avail lim r404 decode
  proc="$(echo "$health" | jq -r '.processor_id // "?"')"
  wid="$(echo "$health" | jq -r '.worker_id // "?"')"
  on="$(echo "$health" | jq -r '.cameras_online // 0')"
  tot="$(echo "$health" | jq -r '.cameras_total // 0')"
  cap="$(echo "$health" | jq -r '.capacity_state // "unknown"')"
  adv="$(echo "$health" | jq -r '.load_advisory // "unknown"')"
  fps="$(echo "$health" | jq -r '.fps_total // 0')"
  fr="$(echo "$health" | jq -r '.frames_received // 0')"
  drop="$(echo "$metrics" | jq -r '.metrics.frames_dropped // 0')"
  cpu="$(echo "$report" | jq -r '.capacity.cpu.percent // empty')"
  mem="$(echo "$report" | jq -r '.capacity.memory.percent // empty')"
  gpu="$(echo "$report" | jq -r '.capacity.gpu.percent // empty')"
  vram="$(echo "$report" | jq -r '.capacity.vram.percent // empty')"
  est="$(echo "$report" | jq -r '.capacity.estimated_capacity_cameras // empty')"
  avail="$(echo "$report" | jq -r '.capacity.estimated_available_cameras // empty')"
  lim="$(echo "$report" | jq -r '.capacity.limiting_resource // empty')"
  r404="$(echo "$report" | jq -r '.summary.rtsp_404_count // 0')"
  decode="$(echo "$health" | jq -r '.decode_backend_effective // "cpu"')"

  echo "${ts},${STAGE},${NOTE},${base},${proc},${wid},${on},${tot},${cap},${adv},${fps},${fr},${drop},${cpu:-},${mem:-},${gpu:-},${vram:-},${est:-},${avail:-},${lim:-},${r404},${decode}"
}

print_header() {
  echo "timestamp_utc,f5_stage,note,base_url,processor_id,worker_id,cameras_online,cameras_total,capacity_state,load_advisory,fps_total,frames_received,frames_dropped,cpu_pct,mem_pct,gpu_pct,vram_pct,est_capacity_cameras,est_available_cameras,limiting_resource,rtsp_404_count,decode_backend"
}

if [[ "${F5_HEADER:-}" == "1" ]]; then
  print_header
  exit 0
fi

if [[ -n "${1:-}" ]]; then
  snapshot_one "$1"
else
  snapshot_one "$PILOT_A"
  snapshot_one "$PILOT_B"
fi
