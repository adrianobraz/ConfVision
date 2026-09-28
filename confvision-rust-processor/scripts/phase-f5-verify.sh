#!/usr/bin/env bash
# Fase 5 — smoke: endpoints de métricas prontos para rampa (sem cargo)
set -euo pipefail

PILOT_A="${PILOT_A:-https://foxpro-rust-pilot.rkr351.easypanel.host}"
PILOT_B="${PILOT_B:-https://foxpro-rust-pilot-b.rkr351.easypanel.host}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

fail=0

echo "=== Fase 5 verify $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="

for base in "$PILOT_A" "$PILOT_B"; do
  base="${base%/}"
  echo "-- ${base}"
  for path in health capacity-report metrics; do
    code="$(curl -sS -o /dev/null -w '%{http_code}' --max-time 20 "${base}/${path}")"
    if [[ "$code" =~ ^2 ]]; then
      echo "  OK /${path} HTTP ${code}"
    else
      echo "  FAIL /${path} HTTP ${code}"
      fail=1
    fi
  done
done

echo ""
echo "-- snapshot test (header + 1 linha)"
F5_HEADER=1 "${SCRIPT_DIR}/phase-f5-snapshot.sh" | head -1
F5_STAGE=0 F5_NOTE=verify "${SCRIPT_DIR}/phase-f5-snapshot.sh" "$PILOT_A" | head -1

echo ""
if [[ "$fail" -ne 0 ]]; then
  echo "RESULT: FAIL Fase 5 verify"
  exit 1
fi
echo "RESULT: OK Fase 5 verify (infra pronta; rampa requer câmeras no sync)"
echo "Próximo: bash confvision-rust-processor/scripts/phase-f5-ramp-run.sh"
