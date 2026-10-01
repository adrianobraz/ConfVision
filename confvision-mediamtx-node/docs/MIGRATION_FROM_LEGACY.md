# Migração — legado → ConfVision MediaMTX Node V1

## Legado

- Path: `confvision-python/mediamtx/mediamtx.yml`
- Docker: `confvision-python/Dockerfile.mediamtx`
- Imagem: `bluenviron/mediamtx:1` (floating tag)
- HLS: `hlsAlwaysRemux: yes`
- Metrics: desligado

## V1

- Path: `confvision-mediamtx-node/config/mediamtx.confvision.yml`
- Docker: `confvision-mediamtx-node/deploy/Dockerfile.v1`
- Imagem: `bluenviron/mediamtx:1.21.0`
- HLS on-demand, metrics `:9998`, playback `:9996`

## Passos

1. Build: `docker build -f confvision-mediamtx-node/deploy/Dockerfile.v1 -t confvision-mediamtx-node:1.0.0 .` (raiz `core4`)
2. Copiar env de `easypanel.env.example` + secrets existentes do serviço foxpro.
3. Manter volume `/recordings`.
4. Smoke: publish + RTSP + API paths list.
5. Rolling update por node (ROLLING_UPDATE.md).

## Inalterado

- RTMP-GUARD código e fluxo auth
- Go, Postgres, Redis, Rust
