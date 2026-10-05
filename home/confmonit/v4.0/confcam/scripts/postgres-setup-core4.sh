#!/bin/bash
# Executar NO servidor core-4 (185.130.61.4) como root.
# Aplica schema se necessario e sincroniza licencas do franqueado teste.
set -euo pipefail

PGURL="${POSTGRES_URL:-postgres://confmonit:20080328DriziN@191.96.156.116:5432/confmonit?sslmode=disable}"
REPO_SQL="${1:-/home/confmonit/v4.0/confvision/sql}"

echo "== Postgres ConfVision setup =="
echo "URL: ${PGURL/@*/@***}"

psql "$PGURL" -c "SELECT version();" >/dev/null
echo "Conexao OK"

for f in 002_central_schema.sql 003_extend_schema.sql 014_capacidade_processamento.sql; do
  if [ -f "$REPO_SQL/$f" ]; then
    echo "Aplicando $f ..."
    psql "$PGURL" -f "$REPO_SQL/$f" || true
  fi
done

echo "== Contagens =="
psql "$PGURL" -c "SELECT COUNT(*) AS vis_licenca FROM vis_licenca;"
psql "$PGURL" -c "SELECT COUNT(*) AS nodes_ativos FROM vis_mediamtx_node WHERE status='ativo';"

SYNC_FILE="$REPO_SQL/sync_franqueado_2025050602383727046281876.sql"
if [ -f "$SYNC_FILE" ]; then
  echo "Aplicando sync licencas Xano ..."
  psql "$PGURL" -f "$SYNC_FILE"
  psql "$PGURL" -c "SELECT status, COUNT(*) FROM vis_licenca WHERE id_franqueado='2025050602383727046281876' GROUP BY status;"
else
  echo "Arquivo sync nao encontrado: $SYNC_FILE"
  echo "Apos deploy do binario novo, abra Minhas licencas no portal para sync automatico."
fi

echo "Concluido."
