#!/bin/bash
# Executar NO servidor Postgres (191.96.156.116 / Hosting Now) como root.
# Libera acesso externo ao banco confmonit para um IP de desenvolvimento.
#
# Uso:
#   ./pg_hba_liberar_ip_externo.sh 186.233.17.187
#
set -euo pipefail

DEV_IP="${1:-}"
if [ -z "$DEV_IP" ]; then
  echo "Informe o IP externo do desenvolvedor. Ex.: $0 186.233.17.187"
  exit 1
fi

PG_HBA=$(sudo -u postgres psql -tAc "SHOW hba_file;" | tr -d ' ')
if [ ! -f "$PG_HBA" ]; then
  echo "pg_hba.conf nao encontrado: $PG_HBA"
  exit 1
fi

MARKER="# confmonit dev externo $DEV_IP"
if grep -qF "$MARKER" "$PG_HBA" 2>/dev/null; then
  echo "Entrada ja existe para $DEV_IP"
else
  echo "$MARKER" | sudo tee -a "$PG_HBA" >/dev/null
  echo "host    confmonit    confmonit    ${DEV_IP}/32    scram-sha-256" | sudo tee -a "$PG_HBA" >/dev/null
  echo "Adicionado em $PG_HBA"
fi

# Garantir escuta em todas as interfaces (porta 5432 ja responde externamente)
PG_CONF=$(sudo -u postgres psql -tAc "SHOW config_file;" | tr -d ' ')
if [ -f "$PG_CONF" ] && ! grep -qE "^listen_addresses\s*=\s*'\*'" "$PG_CONF"; then
  echo "AVISO: confira listen_addresses em $PG_CONF (recomendado '*')"
fi

sudo systemctl reload postgresql || sudo service postgresql reload
echo "PostgreSQL recarregado. Teste de fora:"
echo "  psql \"postgres://confmonit:SENHA@191.96.156.116:5432/confmonit?sslmode=disable\" -c \"SELECT 1\""
