#!/usr/bin/env bash
# Fase 4.1 — gate fmt + testes + builds (CPU sempre; FFmpeg se pkg-config disponível)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "=== Fase 4.1 verify $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="

if ! command -v cargo >/dev/null 2>&1; then
  echo "SKIP: cargo não instalado (normal no CT111/ReceptorTeste — monitor sem build Rust)."
  echo "      Rode este script no host com Rust ou: docker build -f confvision-rust-processor/Dockerfile ."
  echo "      Para alinhamento workers em monitor: bash confvision-rust-processor/scripts/phase-4.1a-verify.sh"
  exit 0
fi

echo "-- cargo fmt --check"
cargo fmt --all -- --check

echo "-- cargo test (default)"
cargo test

echo "-- cargo build --release (CPU default)"
cargo build --release

if command -v pkg-config >/dev/null 2>&1 \
  && pkg-config --exists libavcodec libavutil 2>/dev/null; then
  echo "-- cargo test --features ffmpeg-decode,ffmpeg-nvdec"
  cargo test --features ffmpeg-decode,ffmpeg-nvdec
  echo "-- cargo build --release --features ffmpeg-decode"
  cargo build --release --features ffmpeg-decode
  echo "-- cargo build --release --features ffmpeg-decode,ffmpeg-nvdec"
  cargo build --release --features ffmpeg-decode,ffmpeg-nvdec
else
  echo "SKIP FFmpeg builds (pkg-config / libav dev não encontrados — use Docker Bookworm ou CI Linux)"
fi

echo "RESULT: OK Fase 4.1 verify"
