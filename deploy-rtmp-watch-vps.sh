#!/bin/bash
# Deploy monitor de falhas RTMP (mesma imagem confvision-worker)
#   chmod +x deploy-rtmp-watch-vps.sh
#   ./deploy-rtmp-watch-vps.sh

set -e
TAG="${1:-1}"
IMAGE="confvision-worker:${TAG}"

echo "==> Build ${IMAGE}"
docker build -t "${IMAGE}" .

echo ""
echo "==> Pronto!"
echo "No EasyPanel crie servico confvision-rtmp-watch:"
echo "  Imagem: ${IMAGE}"
echo "  Comando: python"
echo "  Arguments: -u rtmp_watch_main.py"
echo ""
echo "Mount (mesmo volume do MediaMTX):"
echo "  /opt/confvision/recordings -> /recordings"
echo ""
echo "Variaveis:"
echo "  MTX_LOG_FILE=/recordings/mediamtx.log"
echo "  RTMP_WATCH_JSON=/recordings/rtmp_falhas.json"
echo "  RTMP_WATCH_HTTP_PORT=8099"
echo ""
echo "No ConfVision Go (.env):"
echo "  RTMP_WATCH_URL=http://foxpro_confvision-rtmp-watch:8099"
echo ""
echo "MediaMTX precisa gravar log em arquivo (mediamtx.yml):"
echo "  logDestinations: [stdout, file]"
echo "  logFile: /recordings/mediamtx.log"
