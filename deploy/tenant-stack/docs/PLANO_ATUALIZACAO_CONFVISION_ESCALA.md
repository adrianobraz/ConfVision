# Plano de escala ConfVision (U0–U7)

Objetivo: repetir o piloto **foxpro** em **N servidores Proxmox**, com **1 container cliente (`ct_cli_*`)** por fatia de câmeras, **API Go central ConfVision**, **sem Xano** nos fluxos novos.

Decisões v1 (aplicadas neste repositório):

| Tópico | Decisão |
|--------|---------|
| **Imagens** | Hoje: build local (EasyPanel) + binário Go via `build.ps1` + FileZilla. **U2:** registry único (GHCR) + tag semver; host faz `docker pull`. |
| **YOLO** | **1 sidecar por host físico**; todos os Rust do host usam a mesma `YOLO_HTTP_URL` interna. |
| **MediaMTX** | **1 instância por host** (paths `cam/{hash}`); MTX por tenant só se isolamento exigir. |
| **Coleta U3** | **Go central** (`coleta_operacional.go` + `FetchAllProcessorCapacity`); lista de URLs em `RUST_PROCESSOR_BASE_URLS`. |
| **600 cams / host** | Planejamento + `MAX_CAMERAS` + admission Rust; meta operacional (vários `ct_cli_*`); shedding **off** em operação normal. |
| **D5 / U6** | Registry `servidor_id\|https://rust...` em `RUST_PROCESSOR_BASE_URLS`; filtro `?servidor_id=` e assign com `servidor_id` no POST. |

---

## Modelo lógico

```text
SERVIDOR (Proxmox GPU, srv-confvision-042)
├── MediaMTX (1×)           — RTMP/RTSP host
├── rust-yolo-sidecar (1×)  — :8091
├── ct_cli_cliente_a        — Rust MAX_CAMERAS=200, WORKER_ID=ct_cli_cliente_a
├── ct_cli_cliente_b        — Rust MAX_CAMERAS=200
└── (opcional) ingest publisher por tenant

Go central ConfVision (vision.confmonit2.com.br)
├── Postgres vis_camera.worker_id → tenant Rust
├── RUST_PROCESSOR_BASE_URLS → todos os processors (com servidor_id)
├── D5 assign / auto_assign
└── Coleta → vis_sistema_health / vis_sistema_metric
```

**NOITERSEC / A+B:** dois processors no mesmo host (URLs distintas, mesmo `servidor_id`).

---

## Fases

### U0 — Baseline e rastreio

- [ ] Tag git `confvision-u0-YYYY-MM-DD` nos repos após deploy estável.
- [ ] Preencher [RELEASE_BASELINE.md](./RELEASE_BASELINE.md).
- [ ] Foxpro: `LOAD_SHEDDING_ENABLED=0` em produção normal; D5 200; smoke `phase-d6-verify`.
- [ ] Rotacionar `VIS_WORKER_API_KEY` se exposta.

### U1 — Template tenant (`ct_cli_*`)

- [ ] `scripts/tenant-init-env.sh` ou `.ps1` com `TENANT_ID=ct_cli_<slug>`, `MAX_CAMERAS=200`.
- [ ] [FASE1_PROVISIONAMENTO_TENANT.md](./FASE1_PROVISIONAMENTO_TENANT.md) + [docker-compose.host.example.yml](../docker-compose.host.example.yml).
- [ ] Postgres: câmeras com `worker_id = TENANT_ID`.
- [ ] Registrar URLs no Go: `srv-confvision-042|https://rust-a...,srv-confvision-042|https://rust-b...`.

### U2 — Update por host

- [ ] [scripts/update-host.sh](../scripts/update-host.sh) / [update-host.ps1](../scripts/update-host.ps1).
- [ ] [env/registry.env.example](../env/registry.env.example).

### U3 — Coleta multi-host

- [ ] Central: expandir `RUST_PROCESSOR_BASE_URLS` (formato `servidor|url`).
- [ ] `COLETA_RELATORIO_ENABLED=1`, intervalo 15–30m.

### U4 — Dados / Postgres

- [ ] Campo `servidor_id` em câmeras (fase posterior); até lá: `worker_id` = tenant.

### U5 — UI Administrator

- [ ] `/relatorio-operacional` — [U5_ADMINISTRATOR_UI.md](./U5_ADMINISTRATOR_UI.md)

### U6 — D5 multi-host

- [ ] Go: parse `servidor_id|url`, GET capacity `?servidor_id=`, POST assign com `servidor_id`.

### U7 — Escala massiva

- [ ] Automação Proxmox (Terraform/Ansible) fora deste doc.

---

## Referências

- [IMPLANTAR_FASES_0_4.md](./IMPLANTAR_FASES_0_4.md)
- Deploy Go central: `core4/home/confmonit/v4.0/confvision/docs/DEPLOY_GO_CONFVISION.md`
