#!/bin/bash
# Rode na VPS, dentro da pasta confvision (apos enviar os arquivos via WinSCP):
#   chmod +x deploy-vps.sh
#   ./deploy-vps.sh

set -e
TAG="${1:-1}"
IMAGE="confvision-worker:${TAG}"

echo "==> Build ${IMAGE}"
docker build -t "${IMAGE}" .

echo ""
echo "==> Pronto!"
echo "No EasyPanel -> confvision-worker -> Origem -> Imagem Docker:"
echo "  ${IMAGE}"
echo ""
echo "Avancado -> Comando e Arguments: vazios"
echo "Depois clique Implantar"
