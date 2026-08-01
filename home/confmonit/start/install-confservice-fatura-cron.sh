#!/bin/bash
# Instala o cron de fechamento quinzenal do ConfService.
# Use apos subir via FileZilla para /home/confmonit/v4.0/confservice/cron/
# (FileZilla nao escreve em /etc/cron.d — precisa deste passo com sudo)

set -euo pipefail

SRC_CRON="/home/confmonit/v4.0/confservice/cron/confservice-fatura"
DST_CRON="/etc/cron.d/confservice-fatura"
SRC_SH_PKG="/home/confmonit/v4.0/confservice/start/confservice-fatura-fechar.sh"
DST_SH="/home/confmonit/start/confservice-fatura-fechar.sh"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Rode com sudo: sudo $0"
  exit 1
fi

if [[ ! -f "$SRC_CRON" ]]; then
  echo "ERRO: nao achei $SRC_CRON"
  echo "Suba o arquivo via FileZilla para essa pasta e tente de novo."
  exit 1
fi

# Script de execucao (se veio no pacote start/ do confservice)
if [[ -f "$SRC_SH_PKG" ]]; then
  mkdir -p /home/confmonit/start
  cp -f "$SRC_SH_PKG" "$DST_SH"
  chmod +x "$DST_SH"
  echo "OK: script -> $DST_SH"
elif [[ -f "$DST_SH" ]]; then
  chmod +x "$DST_SH"
  echo "OK: script ja existia em $DST_SH"
else
  echo "ERRO: falta confservice-fatura-fechar.sh"
  echo "Suba tambem start/confservice-fatura-fechar.sh para $SRC_SH_PKG"
  exit 1
fi

# /etc/cron.d exige: dono root, sem write para group/other, newline final
cp -f "$SRC_CRON" "$DST_CRON"
chown root:root "$DST_CRON"
chmod 644 "$DST_CRON"
# garante newline no final (cron.d e chato com isso)
sed -i -e '$a\' "$DST_CRON" 2>/dev/null || true

echo "OK: cron -> $DST_CRON"
echo "Agenda: 00:05 nos dias 1 e 16"
echo ""
echo "Teste manual:"
echo "  $DST_SH"
echo "  tail -20 /var/log/confservice-fatura.log"
