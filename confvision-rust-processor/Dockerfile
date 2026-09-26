# syntax=docker/dockerfile:1.4
#
# FFmpeg Bookworm: libavcodec59 / libavutil57 / libavformat59 (FFmpeg 5.1.x; decode usa feature "format").
# Build: cargo-chef (deps cache) + BuildKit cache mounts + um único cargo build do crate app.
# Produção: cargo build --release --features ffmpeg-decode

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
COPY src ./src
RUN cargo chef prepare --recipe-path recipe.json

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
    cargo build --release --features ffmpeg-decode \
    && cp /app/target/release/confvision-rust-processor /app/confvision-rust-processor

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
    && ldd /usr/local/bin/confvision-rust-processor | grep 'not found' && exit 1 || true
ENV HTTP_HOST=0.0.0.0
ENV HTTP_PORT=8090
EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/confvision-rust-processor"]
