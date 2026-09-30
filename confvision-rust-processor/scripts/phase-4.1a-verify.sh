#!/usr/bin/env bash
# Fase 4.1A — alinhamento workers (sem cargo; OK no ReceptorTeste / CT111)
set -euo pipefail

PILOT_A="${PILOT_A:-https://foxpro-rust-pilot.rkr351.easypanel.host}"
PILOT_B="${PILOT_B:-https://foxpro-rust-pilot-b.rkr351.easypanel.host}"
GO_API="${CONFVISION_API_URL:-https://vision.confmonit2.com.br}"
GO_API="${GO_API%/}"
KEY="${VIS_WORKER_API_KEY:-}"

fail=0
wid_a=""
wid_b=""
_last_worker_id=""

echo "=== Fase 4.1A verify $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="

check_processor() {
  local base="$1"
  local label="$2"
  _last_worker_id=""
  echo "=== ${label} @ ${base} ==="
  local health
  health="$(curl -fsS --max-time 25 "${base%/}/health")" || {
    echo "FAIL ${base}/health"
    fail=1
    return 1
  }
  if ! command -v jq >/dev/null 2>&1; then
    echo "$health"
    return 0
  fi
  local st wid proc maxc
  st="$(echo "$health" | jq -r '.status // empty')"
  wid="$(echo "$health" | jq -r '.worker_id // empty')"
  proc="$(echo "$health" | jq -r '.processor_id // empty')"
  maxc="$(echo "$health" | jq -r '.max_cameras // empty')"
  echo "  status=$st worker_id=$wid processor_id=$proc max_cameras=$maxc"
  if [[ "$st" != "ok" ]]; then
    echo "FAIL status=$st"
    fail=1
    return 1
  fi
  if [[ -z "$wid" ]]; then
    echo "FAIL worker_id vazio no /health (env WORKER_ID)"
    fail=1
    return 1
  fi
  if [[ -n "$proc" && -n "$wid" && "$proc" != "$wid" ]]; then
    echo "  INFO processor_id != worker_id (PROCESSOR_ID=métricas, WORKER_ID=assign Postgres)"
  fi
  echo "OK ${label}"
  _last_worker_id="$wid"
}

check_processor "$PILOT_A" "rust A" || true
wid_a="$_last_worker_id"
check_processor "$PILOT_B" "rust B" || true
wid_b="$_last_worker_id"

if [[ -n "$wid_a" && -n "$wid_b" && "$wid_a" == "$wid_b" ]]; then
  echo "FAIL mesmo worker_id nos dois processors"
  fail=1
elif [[ -n "$wid_a" && -n "$wid_b" ]]; then
  echo "OK worker_id distintos: ${wid_a} vs ${wid_b}"
fi

if [[ -n "$KEY" && -n "$wid_a" ]]; then
  echo ""
  echo "== Go sync por worker_id (SHARD_MODE=worker_id / SYNC_FILTER_WORKER_ID)"
  for w in "$wid_a" "$wid_b"; do
    [[ -z "$w" ]] && continue
    qs="?worker_id=$(printf %s "$w" | jq -sRr @uri)"
    n="$(curl -sS -H "X-Vis-Worker-Key: $KEY" "${GO_API}/vis_camera_sync_ativas${qs}" \
      | jq '(.cameras // .dados // []) | length')"
    echo "  worker_id=$w → cameras no sync: $n"
  done
  echo ""
  echo "Postgres: SELECT worker_id, worker_tipo, shard_index, shard_total, max_cameras, cameras_ativas"
  echo "          FROM vis_worker WHERE worker_tipo = 'rust_processor' ORDER BY worker_id;"
else
  echo "SKIP sync Go (defina VIS_WORKER_API_KEY para testar por worker_id)"
fi

echo ""
if [[ "$fail" -ne 0 ]]; then
  echo "RESULT: FAIL Fase 4.1A"
  exit 1
fi
echo "RESULT: OK Fase 4.1A"
