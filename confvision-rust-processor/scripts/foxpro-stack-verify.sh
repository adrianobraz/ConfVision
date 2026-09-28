#!/usr/bin/env bash
# Sonda foxpro: rust-pilot, confvision (MediaMTX), confvision-worker (EasyPanel).
# Uso: bash foxpro-stack-verify.sh [RUST_BASE]
set -euo pipefail

RUST_BASE="${1:-https://foxpro-rust-pilot.rkr351.easypanel.host}"
RUST_BASE="${RUST_BASE%/}"
CONFVISION_URL="${CONFVISION_URL:-https://foxpro-confvision.rkr351.easypanel.host/}"
WORKER_URL="${WORKER_URL:-https://foxpro-confvision-worker.rkr351.easypanel.host/}"
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
echo "    HTTP ${code_cv}"
head -c 120 /tmp/confvision-body.txt 2>/dev/null | tr '\n' ' '
echo ""

echo "==> confvision-worker (EasyPanel proxy): ${WORKER_URL}"
code_w="$(curl -sS -o /tmp/worker-body.txt -w '%{http_code}' --max-time 20 "${WORKER_URL}")"
body_w="$(cat /tmp/worker-body.txt 2>/dev/null || true)"

echo "    HTTP ${code_w}"
if echo "$body_w" | grep -q 'Service is not started'; then
  echo "    DIAG: servico PARADO no EasyPanel (nao e crash de app — falta Start ou start falhou antes de criar container)."
  echo "    ACAO: foxpro / confvision-worker -> Play (Start). Se erro de imagem -> Implantar (build) e Start."
  fail=1
elif echo "$body_w" | grep -q 'No such image'; then
  echo "    DIAG: imagem Docker ausente."
  fail=1
elif [[ "$code_w" == "502" || "$code_w" == "503" ]]; then
  echo "    DIAG: proxy sem backend (container parado ou crash loop)."
  fail=1
else
  echo "    body: $(echo "$body_w" | head -c 100 | tr '\n' ' ')"
fi

echo ""
echo "NOTA: worker Python nao expoe HTTP; dominio publico pode dar 502 mesmo com processo OK."
echo "      Confiar em: EasyPanel Logs ([START] ConfVision worker) e CPU/mem > 0."

if [[ "$fail" -eq 0 ]]; then
  echo "RESULT: stack probes OK (worker pode ainda estar parado se nao checou Start)."
else
  echo "RESULT: FAIL — ver worker / EasyPanel acima."
  exit 1
fi
