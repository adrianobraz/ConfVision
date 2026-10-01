#!/bin/bash
# Instala Redis central ConfVision no core-4 (Proxmox)
set -euo pipefail

DEST=/opt/confvision/redis
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

mkdir -p "$DEST"
cp "$SCRIPT_DIR/docker-compose.redis.yml" "$DEST/"
if [[ ! -f "$DEST/.env ]]; then
  cp "$SCRIPT_DIR/.env.example" "$DEST/.env"
  echo "[OK] Criado $DEST/.env — revise REDIS_PASSWORD"
fi

cd "$DEST"
docker compose -f docker-compose.redis.yml up -d
docker compose -f docker-compose.redis.yml ps

echo ""
echo "Teste local:"
echo "  docker exec confvision-redis redis-cli -a \"\$(grep REDIS_PASSWORD .env | cut -d= -f2)\" ping"
echo ""
echo "Firewall (liberar só VPS/srvN):"
echo "  ufw allow from IP_VPS to any port 6379 proto tcp"
