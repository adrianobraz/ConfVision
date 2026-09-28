#!/usr/bin/env bash
# D6 — verificação operacional 24/7 (processors A/B + sidecar YOLO + opcional Go D5)
set -euo pipefail

PILOT_A="${PILOT_A:-https://foxpro-rust-pilot.rkr351.easypanel.host}"
PILOT_B="${PILOT_B:-https://foxpro-rust-pilot-b.rkr351.easypanel.host}"
SIDECAR="${SIDECAR:-https://foxpro-rust-yolo-sidecar.rkr351.easypanel.host}"
GO_API="${CONFVISION_API_URL:-https://vision.confmonit2.com.br}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

fail=0

echo "=== D6 verify $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="

if [[ -x "${SCRIPT_DIR}/d1-processor-verify.sh" ]] || [[ -f "${SCRIPT_DIR}/d1-processor-verify.sh" ]]; then
  bash "${SCRIPT_DIR}/d1-processor-verify.sh" "${PILOT_A}" "${PILOT_B}" || fail=1
else
  for base in "${PILOT_A}" "${PILOT_B}"; do
    curl -fsS --max-time 20 "${base%/}/health" | grep -q '"status":"ok"' || { echo "FAIL health ${base}"; fail=1; }
  done
fi

if command -v node >/dev/null 2>&1 && [[ -f "${SCRIPT_DIR}/d3-online-test.mjs" ]]; then
  node "${SCRIPT_DIR}/d3-online-test.mjs" || fail=1
else
  code="$(curl -sS -o /dev/null -w '%{http_code}' --max-time 30 -X POST "${SIDECAR%/}/v1/detect" \
    -H 'Content-Type: application/json' -d '{"jpeg_base64":""}')"
  if [[ "${code}" != "400" && "${code}" != "200" ]]; then
    echo "FAIL sidecar POST /v1/detect HTTP ${code} (esperado 400 ou 200)"
    fail=1
  else
    echo "OK sidecar detect HTTP ${code}"
  fi
fi

if [[ -n "${VIS_WORKER_API_KEY:-}" ]]; then
  cap="$(curl -fsS --max-time 25 -H "Authorization: Bearer ${VIS_WORKER_API_KEY}" \
    "${GO_API%/}/vis_rust_processor_capacity" 2>/dev/null || echo '{}')"
  if echo "${cap}" | grep -q '"processors"'; then
    echo "OK Go /vis_rust_processor_capacity"
  else
    echo "WARN Go D5 capacity indisponivel ou nao deployado"
  fi
else
  echo "SKIP Go D5 (defina VIS_WORKER_API_KEY para testar ${GO_API}/vis_rust_processor_capacity)"
fi

if [[ "${fail}" -ne 0 ]]; then
  echo "RESULT: FAIL D6"
  exit 1
fi
echo "RESULT: OK D6"
