#!/bin/bash
# Deploy worker Timelapse Inteligente (mesma imagem confvision-worker)
#   chmod +x deploy-timelapse-vps.sh
#   ./deploy-timelapse-vps.sh

set -e
TAG="${1:-1}"
IMAGE="confvision-worker:${TAG}"

echo "==> Build ${IMAGE}"
docker build -t "${IMAGE}" .

echo ""
echo "==> Pronto!"
echo "No EasyPanel crie servico confvision-timelapse:"
echo "  Source: GitHub adrianobraz/ConfVision (branch main)"
echo "  Comando: python"
echo "  Arguments: -u timelapse_main.py"
echo ""
echo "Variaveis (.env): XANO_BASE_URL, MEDIAMTX_RTSP_BASE, WORKER_ID"
echo "Opcional motion: MOTION_CLIP_MAX_SEC=300, MOTION_POST_ROLL_SEC=5, MOTION_MIN_AREA=1500"
echo "Timelapse producao: TIMELAPSE_FRAME_INTERVALO_SEG=720, TIMELAPSE_FRAMES_POR_SEGMENTO=60"
echo "Timelapse teste:    TIMELAPSE_FRAME_INTERVALO_SEG=20,  TIMELAPSE_FRAMES_POR_SEGMENTO=24"
echo "Nao precisa CONTABO_S3_* — credenciais vem do Xano por franqueado."
echo ""
echo "Documentacao completa: DEPLOY_EASYPANEL.md"
