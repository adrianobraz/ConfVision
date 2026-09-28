#!/usr/bin/env bash
# Aplica rtsp_url_sec (Opção A) via PUT /vis_camera/{id}
# CSV: id,rtsp_url_sec,despausar_analitico (sem virgulas dentro da URL; senão use apply-rtsp-option-a.ps1)
set -euo pipefail

CSV="${1:-}"
DRY=0
if [[ "${2:-}" == "--dry-run" ]]; then DRY=1; fi

if [[ -z "$CSV" || ! -f "$CSV" ]]; then
  echo "Uso: $0 arquivo.csv [--dry-run]" >&2
  exit 1
fi

command -v jq >/dev/null 2>&1 || { echo "jq obrigatorio" >&2; exit 1; }

BASE="${CONFVISION_API_URL:-https://vision.confmonit2.com.br}"
BASE="${BASE%/}"
KEY="${VIS_WORKER_API_KEY:-}"
if [[ -z "$KEY" ]]; then
  echo "Defina VIS_WORKER_API_KEY" >&2
  exit 1
fi

while IFS=',' read -r id rtsp_url_sec despausar_analitico; do
  id="$(echo "${id:-}" | tr -d '\r\"' | xargs)"
  [[ "$id" == "id" || -z "$id" ]] && continue
  rtsp_url_sec="$(echo "${rtsp_url_sec:-}" | tr -d '\r\"' | xargs)"
  despausar_analitico="$(echo "${despausar_analitico:-}" | tr -d '\r\"' | xargs | tr '[:upper:]' '[:lower:]')"

  [[ -z "$rtsp_url_sec" ]] && { echo "WARN id=$id sem rtsp_url_sec" >&2; continue; }
  case "$rtsp_url_sec" in
    rtsp://*|rtsps://*) ;;
    *) echo "WARN id=$id URL invalida" >&2; continue ;;
  esac
  if [[ "$rtsp_url_sec" == *"/live/"* ]]; then
    echo "WARN id=$id URL legado /live/ ignorada" >&2
    continue
  fi

  json=$(jq -nc --arg r "$rtsp_url_sec" '{rtsp_url_sec: $r}')
  if [[ "$despausar_analitico" =~ ^(1|true|yes|sim)$ ]]; then
    json=$(echo "$json" | jq '. + {analitico_pausado: false}')
  fi

  uri="$BASE/vis_camera/$id"
  if [[ "$DRY" -eq 1 ]]; then
    echo "[dry-run] PUT $uri $json"
    continue
  fi

  code=$(curl -sS -o /tmp/opcao-a-put.json -w "%{http_code}" \
    -X PUT "$uri" \
    -H "X-Vis-Worker-Key: $KEY" \
    -H "Content-Type: application/json" \
    -d "$json")
  if [[ "$code" =~ ^2 ]]; then
    echo "OK id=$id HTTP $code"
  else
    echo "FAIL id=$id HTTP $code $(cat /tmp/opcao-a-put.json)" >&2
  fi
done < <(tail -n +2 "$CSV")
