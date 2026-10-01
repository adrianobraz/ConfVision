#!/usr/bin/env sh
# Health Guard + MediaMTX (V2). Uso: GUARD_BASE=http://127.0.0.1:8100 ./mtx-node-health.sh
set -eu
GUARD_BASE="${GUARD_BASE:-http://127.0.0.1:8100}"
MTX_API_BASE="${MTX_API_BASE:-}"
MTX_USER="${MEDIAMTX_API_USER:-}"
MTX_PASS="${MEDIAMTX_API_PASS:-}"

echo "=== ConfVision MediaMTX+Guard health ==="
ok=0

if curl -sf --max-time 10 "${GUARD_BASE}/health" >/dev/null; then
  echo "[OK] Guard /health"
else
  echo "[FALHA] Guard ${GUARD_BASE}/health"
  ok=1
fi

code=$(curl -s -o /tmp/cv_ready.json -w "%{http_code}" --max-time 10 "${GUARD_BASE}/health/ready" || true)
if [ "$code" = "200" ]; then
  echo "[OK] Guard /health/ready"
else
  echo "[WARN] Guard /health/ready HTTP ${code}"
  ok=1
fi

if [ -n "$MTX_API_BASE" ]; then
  if curl -sf --max-time 10 -u "${MTX_USER}:${MTX_PASS}" "${MTX_API_BASE}/v3/paths/list?itemsPerPage=1" >/dev/null; then
    echo "[OK] MediaMTX API paths/list"
  else
    echo "[FALHA] MediaMTX API"
    ok=1
  fi
else
  echo "[SKIP] MTX_API_BASE nao definido"
fi

exit "$ok"
