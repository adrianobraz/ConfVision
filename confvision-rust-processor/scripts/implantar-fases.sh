#!/usr/bin/env bash
# CT111 / Linux — mesma ordem que implantar-fases.ps1 (cadastro via API se curl+key)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_BASE="${CONFVISION_API_URL:-https://vision.confmonit2.com.br}"
RUST_A="${RUST_A:-https://foxpro-rust-pilot.rkr351.easypanel.host}"
RUST_B="${RUST_B:-https://foxpro-rust-pilot-b.rkr351.easypanel.host}"
fail=0

echo "######## Fase 0/D6 — stack ########"
bash "${SCRIPT_DIR}/phase-d6-verify.sh" || fail=1

echo "######## Fase 3 — D2 Redis ########"
bash "${SCRIPT_DIR}/d2-redis-verify.sh" "$RUST_A" "$RUST_B" || true

echo "######## Fase 0 — health A+B ########"
for u in "$RUST_A" "$RUST_B"; do
  curl -fsS --max-time 25 "${u%/}/health" | head -c 500
  echo ""
done

echo "######## D3 online ########"
if command -v node >/dev/null 2>&1; then
  node "${SCRIPT_DIR}/d3-online-test.mjs" --quick || fail=1
fi

echo "######## Fase 5 snapshot ########"
bash "${SCRIPT_DIR}/phase-f5-health-metrics.sh" || true

if [[ "$fail" -eq 0 ]]; then
  echo "RESULT: implantar-fases.sh OK"
  exit 0
fi
echo "RESULT: implantar-fases.sh FAIL"
exit 1
