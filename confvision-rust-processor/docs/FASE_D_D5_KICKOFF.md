# D5 — Control plane Go (assign processor)

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

### Fase D5.1 — Inventário e limites (read-only)

- [ ] Tabela/config de **processors registrados** (URL base, `worker_id`, `vis_mediamtx_node_id`, ativo).
- [ ] Job periódico (cron ou goroutine) que faz **GET /capacity-report** em cada processor ativo.
- [ ] Persistir ou expor **snapshot** (último state, `limiting_resource`, headroom) para admin/API interna.
- [ ] Regra: **não assign** câmera a processor com `load_advisory=reject_admission` ou `estimated_available_cameras=0`.

### Fase D5.2 — Auto-assign (write)

- [ ] Ao criar/reativar câmera analítica (ou fila de “unassigned”): escolher processor com **maior headroom** (mem/CPU conforme `limiting_resource`).
- [ ] `UPDATE cameras SET worker_id = $1 WHERE id = $2` (via camada existente em `visdata/cameras.go`).
- [ ] Respeitar `MAX_CAMERAS` efetivo do processor (ping + capacity-report).
- [ ] Log/auditoria: quem moveu, de/para `worker_id`, motivo (`capacity`, `manual`, `rebalance`).

### Fase D5.3 — Rebalance (opcional piloto)

- [ ] Se `recommended_actions` contém `split_second_processor` / critical prolongado → sugerir ou executar move de câmera (com confirmação ou flag).

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
