# ConfVision MediaMTX Node — V2 (Guard + MediaMTX)

**Streaming Plane** do ConfVision: ingest RTMP, RTSP interno, HLS, recording, API e métricas — **sem** Control Plane. Código Guard em `confvision-python/` (inalterado na lógica de auth).

| | |
|--|--|
| **MEDIAMTX_VERSION** | `1.21.0` |
| **CONFIG_VERSION** | `confvision-mtx-v2.0.0` |
| **Product version** | `2.0.0` |

## Build (recomendado)

Na raiz do repositório `core4`:

```bash
docker build -f confvision-mediamtx-node/deploy/Dockerfile.v2 -t confvision-mediamtx-node:2.0.0 .
```

V1 (`Dockerfile.v1`) permanece equivalente à V2; use `Dockerfile.v2` em novos deploys.

**Entrega operacional completa:** [docs/MEDIAMTX_GUARD_V2_ENTREGA.md](./docs/MEDIAMTX_GUARD_V2_ENTREGA.md)

## Documentação

- [ARCHITECTURE.md](./ARCHITECTURE.md) — visão completa (entrega A–P)
- [docs/MIGRATION_FROM_LEGACY.md](./docs/MIGRATION_FROM_LEGACY.md) — sair de `Dockerfile.mediamtx`
- [docs/TEST_PLAN.md](./docs/TEST_PLAN.md) · [docs/ROLLBACK.md](./docs/ROLLBACK.md)

## Legado

O serviço EasyPanel anterior usa `confvision-python/Dockerfile.mediamtx`. Migre para `Dockerfile.v1` após smoke test (rolling update por node).

## Princípios

- Path RTMP: **`cam/{hash}`** (inalterado)
- Auth: **RTMP-GUARD** HTTP (código não modificado neste pacote)
- Rust: RTSP `:8554` (não modificado)
- Escala: **muitos nodes**, paths dinâmicos — não milhões de entradas YAML
