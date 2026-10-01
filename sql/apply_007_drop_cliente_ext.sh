#!/bin/bash
# Remove vis_cliente_ext do Postgres — codigo interno agora em cliente.CodigoInterno (MySQL).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SQL_FILE="${SCRIPT_DIR}/007_drop_vis_cliente_ext.sql"

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

echo "Verificando (vis_cliente_ext nao deve existir)..."
psql "${POSTGRES_URL}" -c "\dt vis_cliente_ext" 2>&1 | grep -q "Did not find" && echo "OK — tabela removida."
