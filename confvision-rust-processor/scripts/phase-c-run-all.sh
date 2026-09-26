#!/usr/bin/env bash
# Executa verificações C1–C4 (read-only) contra URL(s) do rust-processor.
# Uso: phase-c-run-all.sh [BASE_URL_PILOT_01] [BASE_URL_PILOT_02_opcional]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
URL1="${1:-https://foxpro-rust-pilot.rkr351.easypanel.host}"
URL2="${2:-}"

run_one() {
  local base="$1"
  local label="$2"
  echo ""
  echo "######## ${label}: ${base} ########"
  bash "${SCRIPT_DIR}/phase-c-verify.sh" "$base" || return 1
  bash "${SCRIPT_DIR}/capacity-report.sh" "$base" || true
  if command -v jq >/dev/null 2>&1; then
    report="$(curl -fsS --max-time 20 "${base%/}/capacity-report")"
    echo "--- C3 load ---"
    echo "$report" | jq '.load'
    echo "--- C4 actions ---"
    echo "$report" | jq '[.recommended_actions[] | {code, priority}]'
    ids="$(curl -fsS --max-time 20 "${base%/}/metrics" | jq -c '[.cameras[].camera_id]')"
    echo "--- cameras --- $ids"
  fi
}

fail=0
run_one "$URL1" "C2/C3 pilot-01" || fail=1

if [[ -n "$URL2" ]]; then
  run_one "$URL2" "C4 pilot-02" || fail=1
else
  echo ""
  echo "INFO C4: URL pilot-02 omitida — após criar serviço, passe 2º argumento."
fi

echo ""
echo "=== C1 ==="
echo "Rodar manualmente: bash scripts/c1-ab-baseline.sh \"${URL1}\" 60 10"
echo "Template: docs/c1-ab-results.template.md"

if [[ "$fail" -eq 0 ]]; then
  echo "=== phase-c-run-all: OK (operacional) ==="
  exit 0
fi
echo "=== phase-c-run-all: FAIL ==="
exit 2
