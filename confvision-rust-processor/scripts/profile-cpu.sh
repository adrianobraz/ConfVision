#!/usr/bin/env bash
# Amostragem de CPU real com perf (Linux) — worker confvision-rust-processor.
set -euo pipefail

DURATION="${DURATION:-30}"
FREQ="${FREQ:-99}"
PROC_NAME="${PROC_NAME:-confvision-rust-processor}"
OUT_DIR="${OUT_DIR:-./profile-output}"
THREADS_ONLY=0

usage() {
  cat <<'EOF'
Uso: profile-cpu.sh [PID]

Coleta ~30s de amostras CPU (perf) do worker Rust em produção Linux.

Opções:
  -h, --help           esta ajuda
  --threads-only       só lista threads (delega a rust-process-threads.sh)

Variáveis de ambiente:
  DURATION             segundos de amostragem (default: 30)
  FREQ                 Hz de amostragem (default: 99)
  PROC_NAME            nome do binário
  OUT_DIR              diretório base de saída (default: ./profile-output)
  PERF                 caminho do perf (default: perf)

Saída (em OUT_DIR/run_YYYYMMDD_HHMMSS/):
  perf.data            gravação bruta
  perf-report.txt      relatório textual (perf report)
  perf-top-dso.txt     top por DSO/biblioteca
  threads.txt          snapshot de threads no início
  perf-script.txt      perf script (para flamegraph manual)
  flamegraph.svg       se FlameGraph estiver no PATH (opcional)

Pré-requisitos: linux-tools-perf / perf, permissão para perf record (-p).
EOF
}

resolve_pid() {
  if [[ -n "${1:-}" ]]; then
    echo "$1"
    return 0
  fi
  local pid
  pid="$(pgrep -x "$PROC_NAME" 2>/dev/null | head -n 1 || true)"
  if [[ -z "$pid" ]]; then
    pid="$(pgrep -f "/${PROC_NAME}( |$)" 2>/dev/null | head -n 1 || true)"
  fi
  if [[ -z "$pid" ]]; then
    echo "Erro: processo '$PROC_NAME' não encontrado. Informe o PID como argumento." >&2
    exit 1
  fi
  echo "$pid"
}

check_perf() {
  local perf_bin="${PERF:-perf}"
  if ! command -v "$perf_bin" >/dev/null 2>&1; then
    echo "Erro: '$perf_bin' não encontrado. Instale linux-tools-perf (ou equivalente)." >&2
    exit 1
  fi
  if ! "$perf_bin" list >/dev/null 2>&1; then
    echo "Aviso: perf list falhou — verifique permissões (sudo ou sysctl kernel.perf_event_paranoid)." >&2
  fi
  echo "$perf_bin"
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h | --help)
      usage
      exit 0
      ;;
    --threads-only)
      THREADS_ONLY=1
      shift
      ;;
    *)
      break
      ;;
  esac
done

PID="$(resolve_pid "${1:-}")"

if [[ "$THREADS_ONLY" -eq 1 ]]; then
  exec "$SCRIPT_DIR/rust-process-threads.sh" "$PID"
fi

if ! kill -0 "$PID" 2>/dev/null; then
  echo "Erro: PID $PID inválido ou sem permissão." >&2
  exit 1
fi

PERF_BIN="$(check_perf)"

TS="$(date -u +%Y%m%d_%H%M%S)"
RUN_DIR="$(mkdir -p "$OUT_DIR" && cd "$OUT_DIR" && pwd)/run_${TS}"
mkdir -p "$RUN_DIR"

echo "Profiling PID=$PID por ${DURATION}s @ ${FREQ}Hz -> $RUN_DIR"

{
  echo "pid=$PID"
  echo "duration=${DURATION}"
  echo "freq=${FREQ}"
  echo "proc_name=${PROC_NAME}"
  echo "started_utc=$(date -u -Iseconds)"
  echo "exe=$(readlink -f "/proc/$PID/exe" 2>/dev/null || echo unknown)"
} >"$RUN_DIR/meta.txt"

"$SCRIPT_DIR/rust-process-threads.sh" "$PID" >"$RUN_DIR/threads.txt" 2>&1 || true

echo "Gravando perf.data (pode exigir sudo)..."
set +e
"$PERF_BIN" record \
  -F "$FREQ" \
  -g \
  -p "$PID" \
  -o "$RUN_DIR/perf.data" \
  -- sleep "$DURATION"
RECORD_RC=$?
set -e

if [[ "$RECORD_RC" -ne 0 ]]; then
  echo "Erro: perf record falhou (código $RECORD_RC)." >&2
  echo "Tente: sudo $0 $PID  ou ajuste kernel.perf_event_paranoid (ver docs/CPU_PROFILING_LINUX.md)." >&2
  exit "$RECORD_RC"
fi

echo "Gerando relatórios..."
"$PERF_BIN" report -i "$RUN_DIR/perf.data" --stdio --no-children >"$RUN_DIR/perf-report.txt" 2>&1 || true
"$PERF_BIN" report -i "$RUN_DIR/perf.data" --stdio --sort dso,symbol --percent-limit 0.5 \
  >"$RUN_DIR/perf-top-dso.txt" 2>&1 || true
"$PERF_BIN" script -i "$RUN_DIR/perf.data" >"$RUN_DIR/perf-script.txt" 2>&1 || true

FLAMEGRAPH_SVG="$RUN_DIR/flamegraph.svg"
if command -v stackcollapse-perf.pl >/dev/null 2>&1 && command -v flamegraph.pl >/dev/null 2>&1; then
  echo "Gerando flamegraph.svg (FlameGraph no PATH)..."
  "$PERF_BIN" script -i "$RUN_DIR/perf.data" 2>/dev/null \
    | stackcollapse-perf.pl \
    | flamegraph.pl --title "confvision-rust-processor CPU ($PID)" \
      >"$FLAMEGRAPH_SVG" || rm -f "$FLAMEGRAPH_SVG"
else
  echo "FlameGraph não instalado — pulando SVG (veja docs/CPU_PROFILING_LINUX.md)." >&2
fi

cat >"$RUN_DIR/README.txt" <<EOF
Artefatos gerados em: $RUN_DIR

Relatório rápido:
  less perf-report.txt
  less perf-top-dso.txt

Reabrir interativo:
  perf report -i perf.data

Flamegraph (se tiver FlameGraph instalado manualmente):
  perf script -i perf.data | stackcollapse-perf.pl | flamegraph.pl > flamegraph.svg

Threads no momento da coleta:
  less threads.txt
EOF

echo ""
echo "Concluído."
echo "  perf.data:       $RUN_DIR/perf.data"
echo "  perf-report.txt: $RUN_DIR/perf-report.txt"
if [[ -f "$FLAMEGRAPH_SVG" ]]; then
  echo "  flamegraph.svg:  $FLAMEGRAPH_SVG"
fi
echo "  threads:         $RUN_DIR/threads.txt"
echo "Leia também: docs/CPU_PROFILING_LINUX.md"
