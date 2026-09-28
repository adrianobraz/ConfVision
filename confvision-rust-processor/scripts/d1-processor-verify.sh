#!/usr/bin/env bash
# D1 — verifica 1 ou 2 rust processors (health + capacity-report)
# Uso:
#   d1-processor-verify.sh [URL_PILOT_01]
#   d1-processor-verify.sh URL_01 URL_02
set -euo pipefail

URL1="${1:-https://foxpro-rust-pilot.rkr351.easypanel.host}"
URL2="${2:-}"

strip_trail() { echo "${1%/}"; }
URL1="$(strip_trail "$URL1")"
if [[ -n "$URL2" ]]; then
  URL2="$(strip_trail "$URL2")"
fi

fail=0
LAST_PROC=""

fetch() {
  curl -fsS --max-time 25 "$1"
}

check_one() {
  local base="$1"
  local label="$2"
  echo "=== D1 verify: ${label} @ ${base} ==="
  local health report
  health="$(fetch "${base}/health")" || { echo "FAIL ${base}/health"; fail=1; LAST_PROC=""; return 0; }
  report="$(fetch "${base}/capacity-report" 2>/dev/null || echo '{}')"

  if ! command -v jq >/dev/null 2>&1; then
    echo "$health"
    LAST_PROC=""
    return 0
  fi

  local st proc wid on tot cap r404
  st="$(echo "$health" | jq -r '.status // empty')"
  proc="$(echo "$health" | jq -r '.processor_id // empty')"
  wid="$(echo "$health" | jq -r '.worker_id // empty')"
  on="$(echo "$health" | jq -r '.cameras_online // 0')"
  tot="$(echo "$health" | jq -r '.cameras_total // 0')"
  cap="$(echo "$health" | jq -r '.capacity_state // "unknown"')"
  r404="$(echo "$report" | jq -r '.summary.rtsp_404_count // 0')"

  LAST_PROC="$proc"

  if [[ "$st" != "ok" ]]; then
    echo "FAIL status=$st"
    fail=1
  else
    echo "OK  status=ok processor_id=${proc} worker_id=${wid} cams=${on}/${tot} capacity=${cap} rtsp_404=${r404}"
  fi
  if [[ -n "$proc" && -n "$wid" && "$proc" != "$wid" ]]; then
    echo "WARN processor_id != worker_id (conferir env)"
  fi
  if [[ "$r404" != "0" ]]; then
    echo "WARN rtsp_404_count=$r404"
  fi
}

check_one "$URL1" "pilot-01"
proc1="$LAST_PROC"

if [[ -n "$URL2" ]]; then
  check_one "$URL2" "pilot-02"
  proc2="$LAST_PROC"
  if [[ -n "$proc1" && -n "$proc2" && "$proc1" == "$proc2" ]]; then
    echo "FAIL mesmo processor_id nos dois URLs — env ou roteamento errado"
    fail=1
  elif [[ -n "$proc1" && -n "$proc2" ]]; then
    echo "OK  processors distintos: ${proc1} vs ${proc2}"
  fi
fi

echo ""
if [[ "$fail" -eq 0 ]]; then
  echo "RESULT: D1 verify OK"
  exit 0
fi
echo "RESULT: D1 verify FAIL"
exit 1
