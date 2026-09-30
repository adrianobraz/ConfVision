#!/usr/bin/env bash
# Recria easypanel/foxpro/confvision-worker:latest na VPS foxpro (SSH root).
# Uso: bash foxpro-fix-worker-image.sh [BRANCH]
set -euo pipefail

BRANCH="${1:-main}"
DIR="${CONFVISION_BUILD_DIR:-/opt/confvision-build}"
IMAGE_LOCAL="confvision-worker:1"
IMAGE_EP="easypanel/foxpro/confvision-worker:latest"

echo "==> Clone/update ConfVision branch=${BRANCH} em ${DIR}"
if [[ -d "${DIR}/.git" ]]; then
  git -C "${DIR}" fetch origin
  git -C "${DIR}" checkout "${BRANCH}"
  git -C "${DIR}" pull origin "${BRANCH}"
else
  git clone -b "${BRANCH}" https://github.com/adrianobraz/ConfVision.git "${DIR}"
fi

if [[ ! -f "${DIR}/main.py" || ! -f "${DIR}/Dockerfile" ]]; then
  echo "FAIL: ${DIR} nao tem main.py + Dockerfile na raiz — confira branch/pasta"
  exit 1
fi

echo "==> docker build"
docker build -t "${IMAGE_LOCAL}" "${DIR}"

echo "==> tag ${IMAGE_EP}"
docker tag "${IMAGE_LOCAL}" "${IMAGE_EP}"

echo "==> imagens"
docker images | grep -E 'confvision-worker|foxpro/confvision-worker' || true

echo ""
echo "OK. No EasyPanel: foxpro / confvision-worker -> Start (ou Deploy se rebuild interno)."
