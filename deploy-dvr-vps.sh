#!/bin/bash
# Deploy worker DVR na VPS (mesma imagem do confvision-worker, comando diferente)
#   chmod +x deploy-dvr-vps.sh
#   ./deploy-dvr-vps.sh

set -e
TAG="${1:-1}"
IMAGE="confvision-worker:${TAG}"

echo "==> Build ${IMAGE}"
docker build -t "${IMAGE}" .

echo ""
echo "==> Pronto!"
echo "No EasyPanel crie servico confvision-dvr:"
echo "  Imagem: ${IMAGE}"
echo "  Comando: python"
echo "  Arguments: -u dvr_main.py"
echo ""
echo "Volumes:"
echo "  /recordings -> mesmo volume do MediaMTX"
echo ""
echo "Variaveis (.env): XANO_BASE_URL, MEDIAMTX_API_BASE, DVR_RECORD_DIR=/recordings"
echo "Nao precisa CONTABO_S3_* — credenciais vem do Xano (2242) por franqueado."
