#!/usr/bin/env bash
# Libera TCP 5432 (Postgres) para um IP cliente via UFW.
# Executar como root NO SERVIDOR onde o Postgres escuta (não na VPS foxpro, salvo Postgres local).
#
# Uso:
#   postgres-allow-client-ip.sh 203.0.113.50
#   postgres-allow-client-ip.sh 203.0.113.50 31.97.173.119   # + VPS foxpro

set -euo pipefail

if [[ "${1:-}" == "" ]]; then
  echo "Uso: $0 CLIENT_IP [OUTRO_IP ...]"
  echo "Ex.: $0 \$(curl -fsS ifconfig.me) 31.97.173.119"
  exit 1
fi

if ! command -v ufw >/dev/null 2>&1; then
  echo "ufw não encontrado — configure firewall manualmente (iptables/nftables/painel cloud)."
  exit 1
fi

for ip in "$@"; do
  if [[ ! "$ip" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "IP inválido: $ip"
    exit 1
  fi
  echo "Liberando 5432/tcp de $ip ..."
  ufw allow from "$ip" to any port 5432 proto tcp comment "confmonit psql $ip"
done

ufw status numbered | grep -E '5432|Status' || true
echo "OK. Se ainda falhar auth, ajuste pg_hba.conf e: systemctl reload postgresql"
