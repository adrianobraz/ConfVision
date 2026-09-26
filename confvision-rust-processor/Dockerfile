# syntax=docker/dockerfile:1.4
#
# FFmpeg Bookworm: libavcodec59 / libavutil57 / libavformat59 (FFmpeg 5.1.x; decode usa feature "format").
# Build: cargo-chef (deps cache) + BuildKit cache mounts + cargo build do app após COPY src real.
# Produção: cargo build --release --features ffmpeg-decode
#
# IMPORTANTE: chef prepare usa main stub só para cache de deps. O binário publicado vem SOMENTE do
# build após COPY src, com remoção do artefato stale em /app/target (cache mount pode reter stub ~300KB).

FROM rust:1.88-bookworm AS base
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config \
    libavcodec-dev \
    libavutil-dev \
    libavformat-dev \
    libclang-dev \
    && rm -rf /var/lib/apt/lists/*
RUN cargo install cargo-chef --locked

FROM base AS planner
COPY Cargo.toml Cargo.lock ./
RUN mkdir -p src \
    && printf 'fn main() {}\n' > src/main.rs \
    && cargo chef prepare --recipe-path recipe.json

FROM base AS builder
COPY --from=planner /app/recipe.json recipe.json
COPY Cargo.toml Cargo.lock ./
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/usr/local/cargo/git \
    --mount=type=cache,target=/app/target \
    cargo chef cook --release --features ffmpeg-decode --recipe-path recipe.json
COPY src ./src
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/usr/local/cargo/git \
    --mount=type=cache,target=/app/target \
    rm -f target/release/confvision-rust-processor \
    && rm -rf target/release/.fingerprint/confvision-rust-processor-* \
    && rm -f target/release/deps/confvision_rust_processor-* \
    && cargo build --release --features ffmpeg-decode -p confvision-rust-processor \
    && cp /app/target/release/confvision-rust-processor /app/confvision-rust-processor \
    && test "$(stat -c%s /app/confvision-rust-processor)" -gt 1500000 \
    && grep -aq 'confvision-rust-processor starting' /app/confvision-rust-processor

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    bash \
    procps \
    libavcodec59 \
    libavutil57 \
    libavformat59 \
    libswresample4 \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/confvision-rust-processor /usr/local/bin/confvision-rust-processor
COPY scripts/profile-cpu.sh scripts/rust-process-threads.sh /app/scripts/
RUN chmod 755 /app/scripts/profile-cpu.sh /app/scripts/rust-process-threads.sh \
    && test "$(stat -c%s /usr/local/bin/confvision-rust-processor)" -gt 1500000 \
    && grep -aq 'confvision-rust-processor starting' /usr/local/bin/confvision-rust-processor \
    && ldd /usr/local/bin/confvision-rust-processor | grep 'not found' && exit 1 || true
ENV HTTP_HOST=0.0.0.0
ENV HTTP_PORT=8090
EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/confvision-rust-processor"]
