# Stack mínima por tenant (`ct_cli_*`)

Provisionamento **multi-tenant** (Proxmox / Docker Compose) para o modelo atual:

**Câmera → RTSP Opção A (`rtsp_url_sec`) / MediaMTX → Rust A/B + sidecar YOLO**. Worker Python **descontinuado**.

Sem Xano. Control plane: API Go (`CONFVISION_API_URL` / `XANO_BASE_URL` no worker Python).

---

## Documentos

| Fase | Arquivo |
|------|---------|
| **0–4 — Implantar tudo (ops)** | [docs/IMPLANTAR_FASES_0_4.md](./docs/IMPLANTAR_FASES_0_4.md) |
| **0 — Checklist produção** | [docs/FASE0_CHECKLIST_OPS.md](./docs/FASE0_CHECKLIST_OPS.md) |
| **1 — Novo tenant** | [docs/FASE1_PROVISIONAMENTO_TENANT.md](./docs/FASE1_PROVISIONAMENTO_TENANT.md) |
| Recuperação incidentes | [../../confvision-rust-processor/docs/RECUPERACAO_CAMERAS.md](../../confvision-rust-processor/docs/RECUPERACAO_CAMERAS.md) |

---

## Arquivos deste diretório

| Arquivo | Uso |
|---------|-----|
| `docker-compose.tenant.example.yml` | Referência Compose (MTX + publisher + Rust) |
| `env/tenant.env.example` | Variáveis comuns (`TENANT_ID`, segredos) |
| `env/mtx.env.example` | MediaMTX + Guard |
| `env/publisher.env.example` | Worker Python (ingest RTMP) |
| `env/rust-processor.env.example` | `confvision-rust-processor` |
| `scripts/tenant-init-env.sh` | Gera `.env.*` local a partir do template |
| `scripts/tenant-stack-smoke.sh` | Smoke test (DNS RTSP, health Rust) |

---

## Início rápido (lab / Proxmox CT)

```bash
cd deploy/tenant-stack
export TENANT_ID=ct_cli_rva
export MAX_CAMERAS=200
./scripts/tenant-init-env.sh
# Editar .env.tenant (segredos reais — não commitar)
docker compose -f docker-compose.tenant.example.yml --env-file .env.tenant up -d --build
./scripts/tenant-stack-smoke.sh
```

EasyPanel foxpro: **MediaMTX + sidecar + 2× Rust** (sem `confvision-worker`). Novo tenant: ver Fase 1.

---

## Escala (visão)

- **~400 câmeras com detecção / servidor GPU** → 1–2 tenants de 200 (`ct_cli_*`) por host.
- **Softpark 200k** → ~500 hosts × stack repetível (automação Ansible/Terraform fora deste doc).
