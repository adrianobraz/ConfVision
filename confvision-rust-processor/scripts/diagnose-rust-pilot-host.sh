#!/usr/bin/env bash
# Diagnóstico no HOST da VPS (SSH), quando EasyPanel mostra 0% CPU, log vazio e /health = 502.
# Uso: bash scripts/diagnose-rust-pilot-host.sh [filtro_nome_container]
set -euo pipefail

FILTER="${1:-rust-pilot}"

echo "=== Disco (Docker) ==="
df -h /var/lib/docker 2>/dev/null || df -h /

echo ""
echo "=== Memória ==="
free -h 2>/dev/null || true

echo ""
echo "=== Containers (filtro: ${FILTER}) ==="
docker ps -a --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' 2>/dev/null | grep -i "${FILTER}" || docker ps -a --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' | head -20

CID="$(docker ps -aq --filter "name=${FILTER}" | head -1)"
if [[ -z "${CID}" ]]; then
  echo ""
  echo "Nenhum container encontrado com nome contendo '${FILTER}'."
  echo "No EasyPanel: serviço parado ou nome interno diferente (ex.: foxpro_rust-pilot_1)."
  exit 1
fi

echo ""
echo "=== Inspect (State / OOM / ExitCode) ==="
docker inspect "${CID}" --format 'Name={{.Name}} Status={{.State.Status}} ExitCode={{.State.ExitCode}} OOMKilled={{.State.OOMKilled}} Error={{.State.Error}}'

echo ""
echo "=== Últimos 80 linhas de log (inclui container parado) ==="
docker logs --tail 80 "${CID}" 2>&1 || true

echo ""
echo "=== Teste de bibliotecas FFmpeg no container (se ainda existir) ==="
docker start "${CID}" 2>/dev/null || true
sleep 2
docker exec "${CID}" ldd /usr/local/bin/confvision-rust-processor 2>&1 | grep -E 'not found|libav' || echo "(ldd indisponível ou processo não sobe)"

echo ""
echo "=== Curl interno /health (rede do container) ==="
docker exec "${CID}" bash -c 'command -v curl >/dev/null && curl -sS -o /dev/null -w "HTTP=%{http_code}\n" http://127.0.0.1:8090/health || wget -qO- http://127.0.0.1:8090/health | head -c 200' 2>&1 || echo "Falhou: binário não escuta em 8090 ou container parado."

echo ""
echo "=== Dica ==="
echo "- config error / VIDEO_ACCELERATION=gpu → ajuste env no EasyPanel (use auto ou cpu)."
echo "- error while loading shared libraries → redeploy imagem com libavformat59 no Dockerfile runtime."
echo "- OOMKilled=true → reduza MAX_CAMERAS ou libere RAM; pare ollama/outros serviços."
echo "- Disco 100% → docker system prune (cuidado) ou expanda volume."
