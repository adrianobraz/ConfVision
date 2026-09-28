#!/usr/bin/env bash
# Build local Linux (ou CI) do binário profiling — símbolos para perf, sem strip.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

FEATURES="${FEATURES:-ffmpeg-decode}"
PROFILE="${PROFILE:-profiling}"

echo "cargo build --profile $PROFILE --features $FEATURES"
cargo build --profile "$PROFILE" --features "$FEATURES"

BIN="$ROOT/target/$PROFILE/confvision-rust-processor"
if [[ ! -f "$BIN" ]]; then
  echo "Erro: binário não encontrado em $BIN" >&2
  exit 1
fi

echo ""
echo "Binário: $BIN"
if command -v file >/dev/null 2>&1; then
  file "$BIN"
  if file "$BIN" | grep -q ' stripped'; then
    echo "Erro: binário está stripped — verifique [profile.profiling] no Cargo.toml." >&2
    exit 1
  fi
else
  echo "(comando 'file' indisponível — verifique manualmente no Linux)"
fi

if command -v readelf >/dev/null 2>&1; then
  if readelf -S "$BIN" 2>/dev/null | grep -q '\.debug_info'; then
    echo "OK: seção .debug_info presente"
  else
    echo "Aviso: .debug_info não encontrada — perf pode mostrar só endereços." >&2
  fi
fi

echo ""
echo "No host (perf): export CONFVISION_BINARY=$BIN"
