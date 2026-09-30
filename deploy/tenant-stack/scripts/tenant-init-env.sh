#!/usr/bin/env bash
# Gera .env.tenant, .env.mtx, .env.publisher, .env.rust a partir de env/*.example
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

TENANT_ID="${TENANT_ID:-ct_cli_example}"
MAX_CAMERAS="${MAX_CAMERAS:-200}"

echo "Tenant: ${TENANT_ID} MAX_CAMERAS=${MAX_CAMERAS}"

cp -f env/tenant.env.example .env.tenant
sed -i "s/^TENANT_ID=.*/TENANT_ID=${TENANT_ID}/" .env.tenant
sed -i "s/^MAX_CAMERAS=.*/MAX_CAMERAS=${MAX_CAMERAS}/" .env.tenant

cp -f env/mtx.env.example .env.mtx
cp -f env/publisher.env.example .env.publisher
cp -f env/rust-processor.env.example .env.rust

sed -i "s/\${TENANT_ID}/${TENANT_ID}/g" .env.publisher .env.rust
sed -i "s/^MAX_CAMERAS=.*/MAX_CAMERAS=${MAX_CAMERAS}/" .env.publisher .env.rust
sed -i "s/^PROCESSOR_ID=.*/PROCESSOR_ID=${TENANT_ID}/" .env.rust
sed -i "s/^WORKER_ID=.*/WORKER_ID=${TENANT_ID}/" .env.rust

echo "Criados: .env.tenant .env.mtx .env.publisher .env.rust"
echo "Edite segredos (VIS_WORKER_API_KEY, RTMP_PUBLISH_SECRET, CONFVISION_API_URL, REDIS_URL) antes do compose up."
