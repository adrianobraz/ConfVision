#!/usr/bin/env bash
# Sonda foxpro: Go API, rust-pilot, confvision (MediaMTX). Worker Python fora da política operacional.
# Uso: bash foxpro-stack-verify.sh [RUST_BASE]
set -euo pipefail

RUST_BASE="${1:-https://foxpro-rust-pilot.rkr351.easypanel.host}"
RUST_BASE="${RUST_BASE%/}"
CONFVISION_URL="${CONFVISION_URL:-https://foxpro-confvision.rkr351.easypanel.host/}"
API_HEALTH="${API_HEALTH:-https://vision.confmonit2.com.br/vis_health}"

fail=0

echo "==> Go API: ${API_HEALTH}"
if api="$(curl -fsS --max-time 20 "${API_HEALTH}")"; then
  echo "    OK ${api}"
else
  echo "    FAIL"
  fail=1
fi

echo "==> Rust pilot: ${RUST_BASE}/health"
if health="$(curl -fsS --max-time 20 "${RUST_BASE}/health")"; then
  if command -v jq >/dev/null 2>&1; then
    echo "    $(echo "$health" | jq -c '{status,cameras_online,cameras_total,capacity_state,load_advisory,load_admission_enabled:.load_admission_enabled}')"
  else
    echo "    OK (install jq for pretty JSON)"
  fi
else
  echo "    FAIL health"
  fail=1
fi

echo "==> MediaMTX/guard: ${CONFVISION_URL}"
code_cv="$(curl -sS -o /tmp/confvision-body.txt -w '%{http_code}' --max-time 20 "${CONFVISION_URL}")"
if [[ "$code_cv" =~ ^2 ]]; then
  echo "    OK HTTP ${code_cv}"
else
  echo "    FAIL HTTP ${code_cv}"
  fail=1
fi

echo "==> confvision-worker: SKIP (descontinuado; analitico = Rust A/B)"

if [[ "$fail" -eq 0 ]]; then
  echo ""
  echo "RESULT: OK stack probes (Go + Rust + confvision)"
else
  echo ""
  echo "RESULT: FAIL"
  exit 1
fi
