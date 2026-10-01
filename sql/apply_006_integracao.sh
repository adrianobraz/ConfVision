#!/bin/bash
# Aplica migration integracao Moni — rodar NO SERVIDOR (rede interna ao Postgres).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SQL_FILE="${SCRIPT_DIR}/006_integracao_schema.sql"

if [[ -z "${POSTGRES_URL:-}" ]]; then
  POSTGRES_URL="postgres://confmonit:20080328DriziN@191.96.156.116:5432/confmonit?sslmode=disable"
fi

echo "Aplicando ${SQL_FILE} ..."
if command -v psql >/dev/null 2>&1; then
  psql "${POSTGRES_URL}" -v ON_ERROR_STOP=1 -f "${SQL_FILE}"
else
  cd "${SCRIPT_DIR}/cmd/apply"
  go run . "${SQL_FILE}"
fi

echo "Verificando tabelas..."
psql "${POSTGRES_URL}" -c "\dt vis_integracao*"
psql "${POSTGRES_URL}" -c "\dt vis_cliente_ext"
psql "${POSTGRES_URL}" -c "\d vis_evento" | grep -E "codigo_imagem|imagem_liberada|vis_integracao_imagem" || true
echo "OK — migration 006 aplicada."
