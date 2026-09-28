#!/usr/bin/env bash
# Verifica Opção A: sync analítico + preview de RTSP (sem vazar senha)
set -euo pipefail

BASE="${CONFVISION_API_URL:-https://vision.confmonit2.com.br}"
BASE="${BASE%/}"
KEY="${VIS_WORKER_API_KEY:-}"
WORKER_ID="${WORKER_ID:-}"

if [[ -z "$KEY" ]]; then
  echo "Defina VIS_WORKER_API_KEY" >&2
  exit 1
fi

qs=""
if [[ -n "$WORKER_ID" ]]; then
  qs="?worker_id=$(printf %s "$WORKER_ID" | jq -sRr @uri)"
fi

echo "== GET $BASE/vis_camera_sync_ativas$qs"
body=$(curl -sS -H "X-Vis-Worker-Key: $KEY" "$BASE/vis_camera_sync_ativas$qs")
count=$(echo "$body" | jq '(.cameras // .dados // []) | length')
echo "cameras no sync: $count"

echo "$body" | jq -r '
  (.cameras // .dados // [])[]
  | "  id=\(.id) worker=\(.worker_id // "?") rtsp_sec=\(
      if (.rtsp_url_sec // "") | length > 0 then "sim" else "NAO" end
    ) pausado=\(.analitico_pausado // false)"
'

if [[ "$count" -eq 0 ]]; then
  echo ""
  echo "AVISO: sync vazio — confira analitico_pausado=false, deteccao_humano=true, ativo=true"
  exit 2
fi

missing=$(echo "$body" | jq '[ (.cameras // .dados // [])[] | select((.rtsp_url_sec // "") == "") ] | length')
if [[ "$missing" -gt 0 ]]; then
  echo ""
  echo "AVISO: $missing camera(s) sem rtsp_url_sec — Rust usara cam/{hash} (404 sem worker)"
  exit 3
fi

echo ""
echo "OK Opção A: sync com RTSP sec preenchido"
