#!/usr/bin/env bash
# Fase C - verificacao operacional (C2 + C3 + stream + capacity)
# Uso: phase-c-verify.sh [BASE_URL]
#      phase-c-verify.sh --strict-c3 [BASE_URL]
#      phase-c-verify.sh --prometheus-text [BASE_URL]
set -euo pipefail

PROM=0
STRICT_C3=0
ARGS=()
for a in "$@"; do
  case "$a" in
    --prometheus-text) PROM=1 ;;
    --strict-c3) STRICT_C3=1 ;;
    *) ARGS+=("$a") ;;
  esac
done
BASE="${ARGS[0]:-https://foxpro-rust-pilot.rkr351.easypanel.host}"
if [[ "$PROM" -eq 1 && -n "${ARGS[0]:-}" ]]; then
  BASE="${ARGS[0]}"
fi
BASE="${BASE%/}"

fail=0

fetch() {
  curl -fsS --max-time 20 "$1"
}

health="$(fetch "${BASE}/health")" || { echo "FAIL health"; exit 1; }
report="$(fetch "${BASE}/capacity-report" 2>/dev/null || echo '{}')"

if [[ "$PROM" -eq 1 ]]; then
  if ! command -v jq >/dev/null 2>&1; then
    echo "jq required for --prometheus-text" >&2
    exit 1
  fi
  proc="$(echo "$health" | jq -r '.processor_id // "unknown"')"
  on="$(echo "$health" | jq -r '.cameras_online // 0')"
  tot="$(echo "$health" | jq -r '.cameras_total // 0')"
  cap="$(echo "$health" | jq -r '.capacity_state // "unknown"')"
  cap_crit=0
  [[ "$cap" == "critical" ]] && cap_crit=1
  r404="$(echo "$report" | jq -r '.summary.rtsp_404_count // 0')"
  fps="$(echo "$health" | jq -r '.fps_total // 0')"
  recon="$(echo "$health" | jq -r '.reconnects // 0')"
  cat <<EOF
# TYPE rust_up gauge
rust_up{processor_id="${proc}"} 1
# TYPE rust_cameras_online gauge
rust_cameras_online{processor_id="${proc}"} ${on}
# TYPE rust_cameras_total gauge
rust_cameras_total{processor_id="${proc}"} ${tot}
# TYPE rust_capacity_state gauge
rust_capacity_state{processor_id="${proc}"} ${cap_crit}
# TYPE rust_rtsp_404_count gauge
rust_rtsp_404_count{processor_id="${proc}"} ${r404}
# TYPE rust_fps_total gauge
rust_fps_total{processor_id="${proc}"} ${fps}
# TYPE rust_reconnects counter
rust_reconnects{processor_id="${proc}"} ${recon}
EOF
  exit 0
fi

echo "=== Fase C verify @ ${BASE} ==="

if ! command -v jq >/dev/null 2>&1; then
  echo "FAIL jq nao encontrado - instale jq para C2/C3 verify (ou rode no core-4)"
  if [[ "$STRICT_C3" -eq 1 ]]; then
    exit 2
  fi
fi

if command -v jq >/dev/null 2>&1; then
  st="$(echo "$health" | jq -r '.status')"
  if [[ "$st" != "ok" ]]; then
    echo "FAIL status=$st"
    fail=1
  else
    echo "OK  /health status=ok"
  fi

  r404="$(echo "$report" | jq -r '.summary.rtsp_404_count // empty')"
  if [[ -n "$r404" && "$r404" != "0" ]]; then
    echo "FAIL rtsp_404_count=$r404"
    fail=1
  elif [[ -n "$r404" ]]; then
    echo "OK  rtsp_404_count=0"
  fi

  echo "$report" | jq -e '.load.allow_new_camera' >/dev/null 2>&1 && {
    allow="$(echo "$report" | jq -r '.load.allow_new_camera')"
    admission="$(echo "$report" | jq -r '.load.load_admission_enabled // empty')"
    cap="$(echo "$health" | jq -r '.capacity_state')"
    maxc="$(echo "$health" | jq -r '.max_cameras // empty')"
    echo "INFO load.allow_new_camera=$allow load_admission_enabled=${admission:-unknown} capacity_state=$cap max_cameras=${maxc:-?}"
    if [[ "$STRICT_C3" -eq 1 ]]; then
      if [[ "${admission:-}" != "true" ]]; then
        echo "FAIL C3: LOAD_ADMISSION_ENABLED nao refletido - esperado true - ver easypanel.env.fase-c.vps.example"
        fail=1
      elif [[ "$cap" == "critical" && "$allow" == "true" ]]; then
        echo "FAIL C3: critical mas allow_new_camera=true com admission - revisar LOAD_POLICY_MODE=admission"
        fail=1
      else
        echo "OK  C3 admission configurado"
      fi
      if [[ -n "$maxc" && "$maxc" -gt 15 ]]; then
        echo "WARN C3: max_cameras=$maxc alto - recomendado 10 por instancia CPU"
      fi
    elif [[ "${admission:-}" == "true" && "$allow" == "true" && "$cap" == "critical" ]]; then
      echo "WARN admission ON mas allow_new_camera=true em critical"
    elif [[ "${admission:-}" != "true" && "$cap" == "critical" ]]; then
      echo "WARN C3 pendente: critical sem admission - enable_load_admission no capacity-report"
    fi
  } || true

  echo "--- summary ---"
  echo "$health" | jq '{processor_id, cameras_total, cameras_online, capacity_state, load_advisory, max_cameras: .max_cameras}'
  echo "$report" | jq '{summary, recommended_actions: [.recommended_actions[]?.code]}' 2>/dev/null || true
else
  echo "$health"
fi

if [[ "$fail" -eq 0 ]]; then
  echo "=== Fase C verify: OK ==="
  exit 0
fi
echo "=== Fase C verify: FAIL ==="
exit 2
