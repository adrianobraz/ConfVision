#!/usr/bin/env bash
# D2 — verifica fila Redis nos rust processors (/health)
set -euo pipefail

URL1="${1:-https://foxpro-rust-pilot.rkr351.easypanel.host}"
URL2="${2:-}"
strip() { echo "${1%/}"; }
URL1="$(strip "$URL1")"
[[ -n "$URL2" ]] && URL2="$(strip "$URL2")"

fail=0

check() {
  local base="$1"
  local label="$2"
  echo "=== D2 verify: ${label} @ ${base} ==="
  local h
  h="$(curl -fsS --max-time 25 "${base}/health")" || { echo "FAIL ${base}/health"; fail=1; return; }
  if ! command -v jq >/dev/null 2>&1; then
    echo "$h"
    return
  fi
  local qb ok key depth
  qb="$(echo "$h" | jq -r '.queue_backend // empty')"
  ok="$(echo "$h" | jq -r '.event_queue_redis_ok // false')"
  key="$(echo "$h" | jq -r '.event_queue_key // empty')"
  depth="$(echo "$h" | jq -r '.event_queue_depth // 0')"
  dlq="$(echo "$h" | jq -r '.event_queue_dlq_depth // 0')"
  if [[ "$qb" != "redis" ]]; then
    echo "WARN queue_backend=$qb (esperado redis se D2 ligado)"
  fi
  if [[ "$qb" == "redis" && "$ok" != "true" ]]; then
    echo "FAIL event_queue_redis_ok=$ok"
    fail=1
  else
    echo "OK  queue_backend=${qb} redis_ok=${ok} key=${key} depth=${depth} dlq=${dlq}"
  fi
}

check "$URL1" "pilot-a"
[[ -n "$URL2" ]] && check "$URL2" "pilot-b"

echo ""
if [[ "$fail" -eq 0 ]]; then
  echo "RESULT: D2 verify OK"
  exit 0
fi
echo "RESULT: D2 verify FAILED"
exit 1
