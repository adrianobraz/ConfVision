#!/usr/bin/env bash
# Smoke test — rede RTSP + health Rust (Fase 0)
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ -f .env.tenant ]]; then
  # shellcheck disable=SC1091
  source .env.tenant
fi

MEDIAMTX_RTSP_HOST="${MEDIAMTX_RTSP_HOST:-mtx}"
RUST_HEALTH_URL="${RUST_HEALTH_URL:-http://127.0.0.1:8090/health}"
TENANT_ID="${TENANT_ID:-ct_cli_example}"

echo "== Smoke tenant ${TENANT_ID} =="

echo "-- DNS MTX host: ${MEDIAMTX_RTSP_HOST} --"
if getent hosts "${MEDIAMTX_RTSP_HOST}" >/dev/null 2>&1; then
  echo "OK getent hosts ${MEDIAMTX_RTSP_HOST}"
else
  echo "FAIL: não resolve ${MEDIAMTX_RTSP_HOST} (container mtx na mesma rede?)"
  exit 1
fi

echo "-- TCP 8554 --"
if command -v nc >/dev/null 2>&1; then
  if nc -z -w 3 "${MEDIAMTX_RTSP_HOST}" 8554; then
    echo "OK nc ${MEDIAMTX_RTSP_HOST}:8554"
  else
    echo "WARN: porta 8554 fechada em ${MEDIAMTX_RTSP_HOST}"
  fi
else
  echo "SKIP nc (instalar netcat-openbsd)"
fi

echo "-- Rust /health --"
if command -v curl >/dev/null 2>&1; then
  code="$(curl -sS -o /tmp/tenant-smoke-health.json -w "%{http_code}" "${RUST_HEALTH_URL}" || true)"
  if [[ "${code}" == "200" ]]; then
    echo "OK HTTP ${code} ${RUST_HEALTH_URL}"
    if command -v jq >/dev/null 2>&1; then
      jq -r '"cameras_online=\(.cameras_online) frames_received=\(.frames_received // .metrics.frames_received // "n/a")"' /tmp/tenant-smoke-health.json 2>/dev/null || cat /tmp/tenant-smoke-health.json
    else
      head -c 400 /tmp/tenant-smoke-health.json
      echo
    fi
  else
    echo "FAIL: HTTP ${code} em ${RUST_HEALTH_URL} (rust-processor running?)"
    exit 1
  fi
else
  echo "SKIP curl"
fi

echo "== Smoke concluído =="
