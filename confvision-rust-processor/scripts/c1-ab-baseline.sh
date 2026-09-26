#!/usr/bin/env bash
# C1 — amostras para comparativo Python vs Rust (preencher docs/c1-ab-results.template.md)
# Uso: c1-ab-baseline.sh [BASE_URL] [INTERVAL_SEC] [SAMPLES]
# Ex.: c1-ab-baseline.sh https://foxpro-rust-pilot.rkr351.easypanel.host 60 15
set -euo pipefail

BASE="${1:-https://foxpro-rust-pilot.rkr351.easypanel.host}"
INTERVAL="${2:-60}"
SAMPLES="${3:-10}"
BASE="${BASE%/}"

if ! command -v jq >/dev/null 2>&1; then
  echo "Instale jq" >&2
  exit 1
fi

echo "# C1 baseline Rust @ ${BASE}"
echo "# interval=${INTERVAL}s samples=${SAMPLES}"
echo "# timestamp_iso,fps_total,cameras_online,capacity_state,load_advisory,frames_received,frames_dropped,cpu_percent"

for ((i=1; i<=SAMPLES; i++)); do
  health="$(curl -fsS --max-time 15 "${BASE}/health")"
  report="$(curl -fsS --max-time 15 "${BASE}/capacity-report" 2>/dev/null || echo '{}')"
  ts="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  fps="$(echo "$health" | jq -r '.fps_total // 0')"
  on="$(echo "$health" | jq -r '.cameras_online // 0')"
  cap="$(echo "$health" | jq -r '.capacity_state // "unknown"')"
  adv="$(echo "$health" | jq -r '.load_advisory // "unknown"')"
  fr="$(echo "$health" | jq -r '.frames_received // 0')"
  metrics="$(curl -fsS --max-time 15 "${BASE}/metrics" 2>/dev/null || echo '{}')"
  drop="$(echo "$metrics" | jq -r '.metrics.frames_dropped // 0')"
  cpu="$(echo "$report" | jq -r '.capacity.cpu.percent // empty')"
  echo "${ts},${fps},${on},${cap},${adv},${fr},${drop},${cpu:-na}"
  if [[ "$i" -lt "$SAMPLES" ]]; then
    sleep "$INTERVAL"
  fi
done

echo "# Python: anotar CPU/RAM do container confvision-worker no EasyPanel na mesma janela" >&2
