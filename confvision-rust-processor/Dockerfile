# syntax=docker/dockerfile:1
#
# FFmpeg Bookworm: libavcodec59 / libavutil57 (FFmpeg 5.1.x).
# Build com feature ffmpeg-decode; runtime sem metapackage ffmpeg CLI.

FROM rust:1.88-bookworm AS builder
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config \
    libavcodec-dev \
    libavutil-dev \
    libavformat-dev \
    libclang-dev \
    && rm -rf /var/lib/apt/lists/*
COPY Cargo.toml Cargo.lock* ./
RUN mkdir src && echo "fn main(){}" > src/main.rs \
    && cargo build --release --features ffmpeg-decode \
    && rm -rf src
COPY src ./src
RUN touch src/main.rs && cargo build --release --features ffmpeg-decode

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libavcodec59 \
    libavutil57 \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/target/release/confvision-rust-processor /usr/local/bin/confvision-rust-processor
ENV HTTP_HOST=0.0.0.0
ENV HTTP_PORT=8090
EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/confvision-rust-processor"]
