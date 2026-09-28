#!/usr/bin/env bash
# Verificação pós-deploy: política stream + endpoints relacionados.
# Uso: verify-stream-policy.sh [RUST_BASE] [API_BASE]
set -euo pipefail

RUST_BASE="${1:-https://foxpro-rust-pilot.rkr351.easypanel.host}"
API_BASE="${2:-https://vision.confmonit2.com.br}"
RUST_BASE="${RUST_BASE%/}"
API_BASE="${API_BASE%/}"

echo "=== Stream policy verification ==="
echo "Rust: $RUST_BASE"
echo "API:  $API_BASE"

fail=0

check_http() {
  local url="$1"
  local label="$2"
  local code
  code="$(curl -sS -o /dev/null -w '%{http_code}' --max-time 15 "$url" 2>/dev/null || echo "000")"
  if [[ "$code" =~ ^2 ]]; then
    echo "OK  $label (HTTP $code)"
  else
    echo "FAIL $label (HTTP $code)"
    fail=1
  fi
}

check_http "$RUST_BASE/health" "rust /health"
check_http "$RUST_BASE/capacity-report" "rust /capacity-report"

check_http "$API_BASE/vis_health" "API /vis_health"

react_code="$(curl -sS -o /dev/null -w '%{http_code}' -X POST "$API_BASE/vis_camera_stream_reactivate?camera_id=0" --max-time 15 2>/dev/null || echo "000")"
if [[ "$react_code" == "404" ]]; then
  echo "FAIL API /vis_camera_stream_reactivate -> 404 (API Go ainda sem deploy stream_health)"
  fail=1
elif [[ "$react_code" =~ ^[24] ]]; then
  echo "OK  API /vis_camera_stream_reactivate (HTTP $react_code)"
else
  echo "WARN API /vis_camera_stream_reactivate (HTTP $react_code)"
fi

if command -v jq >/dev/null 2>&1; then
  if metrics="$(curl -fsS "$RUST_BASE/metrics" 2>/dev/null)"; then
    echo "--- rust metrics (stream fields sample) ---"
    echo "$metrics" | jq '[.cameras[]? | {camera_id, stream_failures_consecutive, stream_next_probe_at, stream_local_paused}] | .[0:3]'
  else
    echo "SKIP /metrics (unreachable)"
    fail=1
  fi
fi

if [[ "$fail" -eq 0 ]]; then
  echo "=== Resultado: OK ==="
  exit 0
fi
echo "=== Resultado: FAIL — subir rust-pilot (commit 4a6d36f+) e redeploy API Go ==="
exit 2
