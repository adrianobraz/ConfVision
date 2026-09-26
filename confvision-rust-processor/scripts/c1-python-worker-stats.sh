#!/usr/bin/env bash
# C1 — amostra CPU/RAM do container worker Python (host com Docker)
# Uso: c1-python-worker-stats.sh [CONTAINER_NAME_OR_ID] [INTERVAL] [SAMPLES]
set -euo pipefail

CONTAINER="${1:-confvision-worker}"
INTERVAL="${2:-60}"
SAMPLES="${3:-10}"

if ! command -v docker >/dev/null 2>&1; then
  echo "docker não encontrado — anote CPU/RAM manualmente no EasyPanel" >&2
  exit 1
fi

if ! docker ps --format '{{.Names}}' | grep -qE "^${CONTAINER}$|${CONTAINER}"; then
  echo "Container não encontrado: $CONTAINER" >&2
  docker ps --format 'table {{.Names}}\t{{.Status}}' | head -20
  exit 1
fi

echo "# C1 Python worker docker stats — container=$CONTAINER interval=${INTERVAL}s"
echo "# timestamp_utc,cpu_percent,mem_usage,mem_limit,net_io,block_io"

for ((i=1; i<=SAMPLES; i++)); do
  ts="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  line="$(docker stats "$CONTAINER" --no-stream --format '{{.CPUPerc}},{{.MemUsage}},{{.NetIO}},{{.BlockIO}}' | tr -d ' ')"
  echo "${ts},${line}"
  [[ "$i" -lt "$SAMPLES" ]] && sleep "$INTERVAL"
done
