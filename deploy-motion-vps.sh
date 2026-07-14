#!/bin/bash
# Deploy worker gravacao por movimento (mesma imagem confvision-worker)
#   chmod +x deploy-motion-vps.sh
#   ./deploy-motion-vps.sh

set -e
TAG="${1:-1}"
IMAGE="confvision-worker:${TAG}"

echo "==> Build ${IMAGE}"
docker build -t "${IMAGE}" .

echo ""
echo "==> Pronto!"
echo "No EasyPanel crie servico confvision-motion:"
echo "  Imagem: ${IMAGE}"
echo "  Comando: python"
echo "  Arguments: -u motion_main.py"
echo ""
echo "Variaveis (.env): XANO_BASE_URL, MEDIAMTX_RTSP_BASE, WORKER_ID"
echo "Opcional: MOTION_CLIP_MAX_SEC=300, MOTION_POST_ROLL_SEC=5, MOTION_MIN_AREA=1500"
echo "Nao precisa CONTABO_S3_* — credenciais vem do Xano (2242) por franqueado."
