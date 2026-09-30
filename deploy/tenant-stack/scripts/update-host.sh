#!/usr/bin/env bash
# U2 — Atualiza stack do host (MTX + YOLO + Rust exemplo).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

ENV_FILE="${ENV_FILE:-.env.host}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.host.example.yml}"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Faltando $ENV_FILE — copie env/host.env.example e configure."
  exit 1
fi

# shellcheck disable=SC1090
source "$ENV_FILE"

if [[ -f .env.registry ]]; then
  # shellcheck disable=SC1091
  source .env.registry
  if [[ "${COMPOSE_IMAGE_PULL:-0}" == "1" && -n "${RUST_IMAGE:-}" ]]; then
    echo "Pull imagens registry (IMAGE_TAG=${IMAGE_TAG:-?})..."
    docker pull "${MEDIAMTX_IMAGE:-}"
    docker pull "${YOLO_IMAGE:-}"
    docker pull "${RUST_IMAGE:-}"
  fi
fi

echo "Rebuild/up: $COMPOSE_FILE env=$ENV_FILE HOST_ID=${HOST_ID:-?}"
docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d --build

echo "Ordem de warm-up: aguardando YOLO..."
sleep 5
YOLO_URL="http://127.0.0.1:${YOLO_HTTP_PORT:-8091}/health"
for i in $(seq 1 30); do
  if curl -sf "$YOLO_URL" >/dev/null 2>&1; then
    echo "YOLO OK"
    break
  fi
  sleep 2
done

RUST_URL="${RUST_HEALTH_URL:-http://127.0.0.1:${RUST_HTTP_PORT:-8090}/health}"
if curl -sf "$RUST_URL" >/dev/null 2>&1; then
  echo "Rust OK: $RUST_URL"
else
  echo "AVISO: Rust health falhou em $RUST_URL"
fi

if [[ -x ./scripts/tenant-stack-smoke.sh ]]; then
  ./scripts/tenant-stack-smoke.sh || true
fi

echo "Update host concluído."
