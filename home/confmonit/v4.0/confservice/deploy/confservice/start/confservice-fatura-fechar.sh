#!/bin/bash
# Fecha a quinzena atual do ConfService (POST /internal/fatura/fechar).
# Agendado nos dias 1 e 16 via /etc/cron.d/confservice-fatura

set -euo pipefail

ENV_FILE="/home/confmonit/v4.0/confservice/.env"
URL="http://127.0.0.1:2020/internal/fatura/fechar"
LOG="/var/log/confservice-fatura.log"
API_KEY=""

if [[ -f "$ENV_FILE" ]]; then
  # shellcheck disable=SC1090
  API_KEY="$(grep -E '^API_KEY=' "$ENV_FILE" | head -1 | cut -d= -f2- | tr -d '\r' | tr -d '"' | tr -d "'")"
fi

if [[ -z "$API_KEY" ]]; then
  echo "$(date -Iseconds) ERRO: API_KEY vazio em $ENV_FILE" >>"$LOG"
  exit 1
fi

{
  echo "$(date -Iseconds) inicio fechamento fatura"
  curl -sS -X POST "$URL" \
    -H "X-Api-Key: ${API_KEY}" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    --max-time 60
  echo
  echo "$(date -Iseconds) fim"
} >>"$LOG" 2>&1
