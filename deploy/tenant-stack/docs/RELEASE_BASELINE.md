# Release baseline (U0)

Preencher a cada deploy significativo em produção.

| Componente | Versão / tag | Commit / notas | Data |
|------------|--------------|----------------|------|
| ConfVision Go (central) | | branch `fix/confvision-rtmp-postgres-nal` ou `main` | |
| confvision-rust-processor | `PROCESSOR_VERSION` env | branch `rust-pilot` | |
| MediaMTX image | Dockerfile-mediamtx | | |
| rust-yolo-sidecar | Dockerfile.yolo-sidecar | | |
| Postgres schema coleta | `coleta_operacional_migration.sql` | | |

## Env críticos (foxpro / host)

| Variável | Produção recomendada |
|----------|----------------------|
| `LOAD_SHEDDING_ENABLED` | `0` (normal); `1` só emergência CPU |
| `D5_AUTO_ASSIGN_ENABLED` | `1` central Go |
| `RUST_PROCESSOR_BASE_URLS` | `servidor\|url,...` |
| `YOLO_HTTP_URL` (Rust) | `http://yolo:8091` (rede host) |
| `MAX_CAMERAS` | `200` por tenant |

## Smoke

```bash
# Rust pilot repo
./scripts/d3-online-test.sh quick
./deploy/tenant-stack/scripts/tenant-stack-smoke.sh
# Go (onde existir)
./scripts/phase-d6-verify.sh
```

## Tag sugerida

```bash
git tag -a confvision-u0-2026-03-30 -m "Baseline pós D5/coleta/shedding"
```
