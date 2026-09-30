#!/usr/bin/env bash
# Gera .env.host + .env.mtx + .env.yolo + .env.rust para docker-compose.host.example.yml
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

HOST_ID="${HOST_ID:-srv_confvision_lab}"
TENANT_ID="${TENANT_ID:-ct_cli_example}"
MAX_CAMERAS="${MAX_CAMERAS:-200}"

echo "Host: ${HOST_ID} Tenant: ${TENANT_ID} MAX_CAMERAS=${MAX_CAMERAS}"

cp -f env/host.env.example .env.host
sed -i "s/^HOST_ID=.*/HOST_ID=${HOST_ID}/" .env.host
sed -i "s/^TENANT_ID=.*/TENANT_ID=${TENANT_ID}/" .env.host
sed -i "s/^MAX_CAMERAS=.*/MAX_CAMERAS=${MAX_CAMERAS}/" .env.host

cp -f env/mtx.env.example .env.mtx
cp -f env/yolo-sidecar.env.example .env.yolo
cp -f env/rust-processor.env.example .env.rust

sed -i "s/\${TENANT_ID}/${TENANT_ID}/g" .env.rust
sed -i "s/^MAX_CAMERAS=.*/MAX_CAMERAS=${MAX_CAMERAS}/" .env.rust
sed -i "s/^PROCESSOR_ID=.*/PROCESSOR_ID=${TENANT_ID}/" .env.rust
sed -i "s/^WORKER_ID=.*/WORKER_ID=${TENANT_ID}/" .env.rust
sed -i "s|^MEDIAMTX_RTSP_BASE=.*|MEDIAMTX_RTSP_BASE=rtsp://mtx:8554|" .env.rust

if ! grep -q '^YOLO_HTTP_URL=' .env.rust; then
  echo "YOLO_HTTP_URL=http://yolo:8091" >> .env.rust
else
  sed -i "s|^YOLO_HTTP_URL=.*|YOLO_HTTP_URL=http://yolo:8091|" .env.rust
fi

echo "Criados: .env.host .env.mtx .env.yolo .env.rust"
echo "Edite segredos antes de: ./scripts/update-host.sh"
