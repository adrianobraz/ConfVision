# syntax=docker/dockerfile:1

FROM rust:1.88-bookworm AS builder
WORKDIR /app
COPY Cargo.toml Cargo.lock* ./
RUN mkdir src && echo "fn main(){}" > src/main.rs && cargo build --release && rm -rf src
COPY src ./src
# Força rebuild do binário real (evita reutilizar artefato do main.rs dummy da camada de cache).
RUN touch src/main.rs && cargo build --release

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/target/release/confvision-rust-processor /usr/local/bin/confvision-rust-processor
ENV HTTP_HOST=0.0.0.0
ENV HTTP_PORT=8090
EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/confvision-rust-processor"]
