# Fase D — produção foxpro (Rust analítico + Go/Postgres)

**Início:** 2026-09-27  
**Fechamento D5/D6:** 2026-09-28 — ver **[FASE_D_FECHAMENTO.md](./FASE_D_FECHAMENTO.md)** (produção oficial; nomes *piloto* no EasyPanel são legado).

**Stack fixa:** MediaMTX (`confvision`) + **Rust** analítico. Detecção **não** volta ao worker Python.

## Regras desta fase

| Regra | Detalhe |
|-------|---------|
| **Sem Xano novo** | Control plane = **Go** (`https://vision.confmonit2.com.br`) + **Postgres**. |
| **D4 adiado** | GPU / decode NVDEC só após servidor dedicado. |
| **Redis existente** | `REDIS_URL` → ex. `185.130.61.5:6379`; D2 `QUEUE_BACKEND=redis` no Rust. |
| **VPS realista** | Poucas câmeras **por processor** (`MAX_CAMERAS=3–4`); 2 processors no mesmo host = roteamento D5. |

---

## Blocos (escopo VPS)

| Bloco | Objetivo | Status |
|-------|----------|--------|
| **D1** | N processors, `worker_id`, EasyPanel A+B | **Produção foxpro** — [FASE_D_D1_KICKOFF.md](./FASE_D_D1_KICKOFF.md) |
| **D2** | Fila eventos Redis, retry, DLQ | **Produção** — [FASE_D_D2_KICKOFF.md](./FASE_D_D2_KICKOFF.md) |
| **D3** | YOLO / eventos / clips no Rust | **Produção ligada; aceite E2E `vis_evento` em andamento** — [FASE_D_D3_KICKOFF.md](./FASE_D_D3_KICKOFF.md) |
| **D4** | GPU | **Fora de escopo** |
| **D5** | Go: capacity + auto-assign | **Fechado (D5.1–D5.2)** — [FASE_D_D5_KICKOFF.md](./FASE_D_D5_KICKOFF.md) |
| **D6** | Monitor 24/7 + playbooks | **Fechado (D6.1)** — [FASE_D_D6_KICKOFF.md](./FASE_D_D6_KICKOFF.md) |

**Ordem histórica:** D1 → D2 → D3 → (D5 ∥ D6).

---

## Subir stack para clientes (resumo)

1. **Proxmox:** Go `confmonit4confvision` + D5 env.  
2. **EasyPanel:** Start `confvision` → sidecar YOLO → rust A+B (**`confvision-worker` permanece Stop**).  
3. **Postgres / D5:** `worker_id` nas câmeras analíticas.

Detalhe: **[FASE_D_FECHAMENTO.md](./FASE_D_FECHAMENTO.md)** § Colocar ConfVision no ar.

---

## D3 — aceite contínuo

- Sidecar + pilots A/B; testes `d3-online-test.mjs`, `d3-stack-check.mjs`.
- **Pendente produto:** 1 evento E2E → `vis_evento` Postgres; RAM/timeout YOLO conforme carga.

---

## Referências

- Recuperação câmeras / RTSP 404: [RECUPERACAO_CAMERAS.md](./RECUPERACAO_CAMERAS.md)
- Monitor CT111: [C2_MONITOR_INTEGRACAO_EDGE.md](./C2_MONITOR_INTEGRACAO_EDGE.md)
- Admission / capacity: [LOAD_ADMISSION.md](./LOAD_ADMISSION.md)

Verificação:

```bash
bash confvision-rust-processor/scripts/phase-d6-verify.sh
node confvision-rust-processor/scripts/d3-online-test.mjs
```
