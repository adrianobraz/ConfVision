#!/usr/bin/env bash
# Fase 5 — rampa medida: amostra em cada degrau enquanto operador sobe carga no Postgres/cadastro
#
# Pré-requisitos: câmeras elegíveis ao sync (ativo, deteccao_humano, !pausado), RTSP Opção A,
# worker_id assign A/B, processors no ar.
#
# Env:
#   F5_STAGES="1 5 10 20 40 50"   degraus alvo (soma A+B cameras_online)
#   F5_HOLD_SEC=180                 segundos amostrando em cada degrau
#   F5_SAMPLE_EVERY=30              intervalo entre snapshots
#   F5_OUT=ramp-f5.csv              arquivo de saída
#   F5_INTERACTIVE=1                pausa antes de cada degrau (default 1)
#
set -euo pipefail

PILOT_A="${PILOT_A:-https://foxpro-rust-pilot.rkr351.easypanel.host}"
PILOT_B="${PILOT_B:-https://foxpro-rust-pilot-b.rkr351.easypanel.host}"
STAGES="${F5_STAGES:-1 5 10 20 40 50}"
HOLD="${F5_HOLD_SEC:-180}"
EVERY="${F5_SAMPLE_EVERY:-30}"
OUT="${F5_OUT:-ramp-f5-$(date -u +%Y%m%dT%H%M%SZ).csv}"
INTERACTIVE="${F5_INTERACTIVE:-1}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if ! command -v jq >/dev/null 2>&1; then
  echo "jq obrigatorio" >&2
  exit 1
fi

sum_online() {
  local a b
  a="$(curl -fsS --max-time 15 "${PILOT_A%/}/health" | jq -r '.cameras_online // 0')"
  b="$(curl -fsS --max-time 15 "${PILOT_B%/}/health" | jq -r '.cameras_online // 0')"
  echo $((a + b))
}

echo "=== Fase 5 ramp $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="
echo "stages=${STAGES} hold=${HOLD}s sample_every=${EVERY}s out=${OUT}"

F5_HEADER=1 bash "${SCRIPT_DIR}/phase-f5-snapshot.sh" > "$OUT"

for target in $STAGES; do
  echo ""
  echo "--- Degrau alvo: ${target} câmeras online (A+B) ---"
  if [[ "$INTERACTIVE" == "1" ]]; then
    echo "Operador: assign/despausar/RTSP até ~${target} online; Enter para amostrar ${HOLD}s (Ctrl+C aborta)"
    read -r _
  fi

  start=$(date +%s)
  end=$((start + HOLD))
  while [[ $(date +%s) -lt $end ]]; do
    online="$(sum_online)"
    echo "[$(date -u +%H:%M:%S)] cameras_online total=${online} (meta ${target})"
    F5_STAGE="$target" F5_NOTE="online=${online}" bash "${SCRIPT_DIR}/phase-f5-snapshot.sh" >> "$OUT"
    sleep "$EVERY"
  done

  online="$(sum_online)"
  if [[ "$online" -lt "$target" ]]; then
    echo "WARN: degrau ${target} não atingido (online=${online}). Cadastro/sync/RTSP?"
    echo "      Pode continuar rampa ou corrigir antes do próximo degrau."
  else
    echo "OK degrau ${target} (online=${online})"
  fi

  rep_a="$(curl -fsS --max-time 15 "${PILOT_A%/}/capacity-report" 2>/dev/null || echo '{}')"
  cap="$(echo "$rep_a" | jq -r '.capacity.state // "?"')"
  lim="$(echo "$rep_a" | jq -r '.capacity.limiting_resource // "?"')"
  adv="$(echo "$rep_a" | jq -r '.load.advisory // .load_advisory // "?"')"
  echo "  capacity_state=${cap} limiting=${lim} load_advisory=${adv}"
  if [[ "$cap" == "critical" ]]; then
    echo "  AVISO: capacity critical — considere parar rampa antes do próximo degrau"
  fi
done

echo ""
echo "RESULT: Fase 5 ramp concluída → ${OUT}"
echo "Preencha docs/FASE_5_RESULTADO.template.md com conclusões."
