# D5 — Control plane Go (assign processor)

**Status:** **Fechado (D5.1–D5.2)** em 2026-09-28 — produção oficial. Ver [FASE_D_FECHAMENTO.md](./FASE_D_FECHAMENTO.md).

## Em uma frase

**D5 — Control plane Go:** fonte da verdade para sync, ping, limites de capacidade e (meta) **auto-assign** câmera → processor. **Sem Xano novo.**

---

## Objetivo

Substituir split manual SQL (D1) por política no **Go** (`https://vision.confmonit2.com.br`):

1. **Ingestão** de saúde/capacidade dos rust-processors.
2. **Decisão** de qual `worker_id` recebe câmera nova ou rebalanceamento.
3. **Persistência** em Postgres (`cameras.worker_id`, pausas, limites).

---

## Fontes de dados (já existem no Rust)

| Fonte | URL | Uso D5 |
|-------|-----|--------|
| Ping | `POST /vis_worker_ping` (Rust → Go) | `worker_id`, `cameras_ativas`, `max_cameras`, `queue_backend`, `yolo_device`, CPU/mem |
| Capacity | `GET {processor}/capacity-report` | `capacity_state`, `limiting_resource`, `estimated_available_cameras`, `load_advisory`, `recommended_actions` |
| Métricas | `GET {processor}/metrics` | Detalhe por câmera (fps, offline, motion) |
| Health D3 | `GET {processor}/health` | `events_published`, `event_queue_depth`, YOLO/capture flags |

Processors foxpro (piloto):

- `https://foxpro-rust-pilot.rkr351.easypanel.host`
- `https://foxpro-rust-pilot-b.rkr351.easypanel.host`

Scripts de referência: [`scripts/capacity-report.sh`](../scripts/capacity-report.sh), [`scripts/d1-processor-verify.sh`](../scripts/d1-processor-verify.sh).

---

## Escopo de implementação Go

### Fase D5.1 — Inventário e limites (read-only) ✅

- [x] URLs via env **`RUST_PROCESSOR_BASE_URLS`** (default foxpro A+B).
- [x] **`GET /vis_rust_processor_capacity`** — agrega `/capacity-report` de cada processor.
- [x] Score + `assign_eligible` (respecta `allow_new_camera`, advisory reject, headroom).

### Fase D5.2 — Auto-assign (write) ✅

- [x] **`POST /vis_camera_assign_processor`** — `camera_id` + opcional `worker_id` manual; senão auto.
- [x] **`POST /ops/d5/auto_assign_analiticas`** — batch (`dry_run`, `limit`).
- [x] **`CreateCamera`** — se analítico sem `worker_id` e `D5_AUTO_ASSIGN_ENABLED` → assign automático (`d5_assign` na resposta).

### Fase D5.3 — Rebalance (opcional piloto)

- [ ] Move automático entre processors em critical prolongado (manual via assign API por enquanto).

### Deploy Go (Ambiente)

```env
RUST_PROCESSOR_BASE_URLS=https://foxpro-rust-pilot.rkr351.easypanel.host,https://foxpro-rust-pilot-b.rkr351.easypanel.host
D5_AUTO_ASSIGN_ENABLED=1
```

Código: `home/confmonit/v4.0/confvision/src/modulos/visdata/rust_processor_d5.go`

---

## Código Go (ponto de partida no repo)

| Área | Path |
|------|------|
| Ping | `home/confmonit/v4.0/confvision/src/modulos/visdata/workers.go` |
| Câmeras / worker_id | `home/confmonit/v4.0/confvision/src/modulos/visdata/cameras.go` |
| Rotas | `home/confmonit/v4.0/confvision/src/modulos/visdata/router.go` |

Rust **não** precisa mudar para D5.1; D5.2 só exige sync Go aplicar `worker_id` que o processor já filtra (`SYNC_FILTER_WORKER_ID`).

---

## Fora de escopo D5

- Xano / novas APIs públicas além do necessário no Go existente.
- Orquestração EasyPanel (criar app processor) — continua manual ou runbook D1.
- GPU / D4.

---

## Verificação

1. Dois processors com loads diferentes → job D5 escolhe o menos carregado para câmera teste.
2. Processor em `critical` + `reject_admission` → **não** recebe câmera nova.
3. `ListCamerasAnaliticas(worker_id=…)` reflete assign após sync.

---

## Referências

- [FASE_D.md](./FASE_D.md) — ordem D3 → (D5 ∥ D6)
- [FASE_D_D1_KICKOFF.md](./FASE_D_D1_KICKOFF.md) — split manual atual
- [LOAD_ADMISSION.md](./LOAD_ADMISSION.md) — admission e capacity-report
