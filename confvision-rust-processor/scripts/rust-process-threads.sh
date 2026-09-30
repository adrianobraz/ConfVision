#!/usr/bin/env bash
# Lista threads do worker Rust (PID/TID/nome/CPU quando disponível).
set -euo pipefail

PROC_NAME="${PROC_NAME:-confvision-rust-processor}"

usage() {
  cat <<'EOF'
Uso: rust-process-threads.sh [PID]

  PID opcional. Sem PID, tenta localizar confvision-rust-processor via pgrep.

Variáveis:
  PROC_NAME   nome do binário (default: confvision-rust-processor)
EOF
}

resolve_pid() {
  if [[ $# -ge 1 && -n "${1:-}" ]]; then
    echo "$1"
    return 0
  fi
  local pid
  pid="$(pgrep -x "$PROC_NAME" 2>/dev/null | head -n 1 || true)"
  if [[ -z "$pid" ]]; then
    pid="$(pgrep -f "/${PROC_NAME}( |$)" 2>/dev/null | head -n 1 || true)"
  fi
  if [[ -z "$pid" ]]; then
    echo "Erro: processo '$PROC_NAME' não encontrado." >&2
    exit 1
  fi
  echo "$pid"
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

PID="$(resolve_pid "${1:-}")"

if ! kill -0 "$PID" 2>/dev/null; then
  echo "Erro: PID $PID inválido ou sem permissão." >&2
  exit 1
fi

echo "Processo: PID=$PID ($(readlink -f "/proc/$PID/exe" 2>/dev/null || echo '?'))"
echo ""
printf "%-8s %-8s %6s %-12s %-16s %s\n" "PID" "TID" "%CPU" "TIME" "COMM" "CMD (truncado)"
echo "--------------------------------------------------------------------------------"

if ps -T -p "$PID" -o pid=,tid=,pcpu=,time=,comm=,args= 2>/dev/null; then
  :
else
  # BusyBox / ambientes mínimos
  for tid_path in "/proc/$PID/task"/*; do
    tid="$(basename "$tid_path")"
    comm="$(cat "$tid_path/comm" 2>/dev/null || echo '?')"
    echo "$PID $tid ? ? $comm"
  done
fi

echo ""
echo "Detalhe por TID (comm em /proc):"
for tid_path in "/proc/$PID/task"/*; do
  tid="$(basename "$tid_path")"
  comm="$(cat "$tid_path/comm" 2>/dev/null || echo '?')"
  echo "  TID=$tid comm=$comm"
done
