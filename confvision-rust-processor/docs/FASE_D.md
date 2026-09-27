# Fase D — produto (VPS foxpro + Go/Postgres)

**Início:** 2026-09-27  
**Stack fixa:** MediaMTX (`confvision`) + **Rust** analítico. **Worker Python off.**

## Regras desta fase

| Regra | Detalhe |
|-------|---------|
| **Sem Xano novo** | Nenhuma tabela/API/função Xano. Control plane = **Go** (`https://vision.confmonit2.com.br`) + **Postgres**. |
| **D4 adiado** | GPU / GEX44 / decode NVDEC só após **servidor dedicado**. |
| **Redis existente** | `REDIS_URL` → host interno (ex. `185.130.61.5:6379`); D2 liga `QUEUE_BACKEND=redis` no Rust. |
| **VPS realista** | Poucas câmeras **por processor** (`MAX_CAMERAS=3–4`); 2 processors no **mesmo host** = piloto de roteamento, não escala infinita. |

---

## Blocos (escopo VPS)

| Bloco | Objetivo | Status |
|-------|----------|--------|
| **D1** | N processors, `worker_id`, deploy EasyPanel | **Piloto foxpro** (A+B) — [FASE_D_D1_KICKOFF.md](./FASE_D_D1_KICKOFF.md) |
| **D2** | Fila eventos Redis, retry, DLQ | **Produção** (pilots A/B) — [FASE_D_D2_KICKOFF.md](./FASE_D_D2_KICKOFF.md) |
| **D3** | YOLO / eventos / clips no Rust | **Piloto ligado; E2E `vis_evento` pendente** — [FASE_D_D3_KICKOFF.md](./FASE_D_D3_KICKOFF.md) |
| **D4** | GPU | **Fora de escopo** até novo servidor |
| **D5** | Go: limites, auto-assign processor | **Kickoff** — [FASE_D_D5_KICKOFF.md](./FASE_D_D5_KICKOFF.md) |
| **D6** | SLO 24/7, alertas, playbooks | **Kickoff** (estende C2) — [FASE_D_D6_KICKOFF.md](./FASE_D_D6_KICKOFF.md) |

**Ordem:** D1 → D2 → D3 → **(D5 ∥ D6)**.

### Referência rápida (não se perder)

| Bloco | O quê |
|-------|--------|
| **D5 — Control plane Go** | Fonte da verdade: sync, ping, limites de capacidade, **auto-assign** câmera → processor. |
| **D6 — SLO operacional** | Operação 24/7: alertas (C2), playbooks (RTSP 404, OOM, réplica 0/1), doc recuperação câmeras. |

---

## D3 — estado atual (foxpro)

**Feito (piloto):**

- Sidecar `foxpro/rust-yolo-sidecar` (Torch CPU), pilots A/B com D2+D3 env.
- Testes: `d3-online-test.mjs`, `d3-stack-check.mjs`.
- Motion detectado (ex. câmera 22); sidecar HTTP 200 em inferência.

**Pendente (aceite produto):**

- **1 evento E2E:** movimento + pessoa → `events_published` → fila Redis → `POST /vis_evento` → registro Postgres.
- Revisar antes de prod **confiável**:
  - **YOLO HTTP timeout** Rust = 15s vs inferência CPU sidecar (cold ~30s+; risco de falha intermitente).
  - **JPEG** (`Corrupt Huffman` no sidecar) — qualidade snapshot no Rust.
  - **RAM** — 2 pilots + sidecar no mesmo host (`capacity_state=critical`, `limiting_resource=memory`).

---

## D1 — entregáveis

1. App EasyPanel **`foxpro-rust-pilot-02`** (clone env de [`easypanel.env.fase-c.processor-02.example`](../easypanel.env.fase-c.processor-02.example)).
2. Postgres: split `worker_id` — [`sql/phase_d_foxpro_split_vps_conservative.sql`](../sql/phase_d_foxpro_split_vps_conservative.sql) (pausar **id 2**, split **5/18/19** vs **22**). Alternativa fleet maior: [`phase_c_foxpro_split_pilot02.sql`](../sql/phase_c_foxpro_split_pilot02.sql).
3. **`confvision-worker` Stop** (Rust-only).
4. Verificação: [`scripts/d1-processor-verify.sh`](../scripts/d1-processor-verify.sh).
5. Runbook deploy: [`deploy/phase-c/EASYPANEL_SECOND_PROCESSOR.md`](../deploy/phase-c/EASYPANEL_SECOND_PROCESSOR.md).

---

## D2 — Redis (infra já existe)

- Worker/sync legado: `CONFIG_CACHE_BACKEND=redis`, chaves `confvision:sync:…`.
- **Rust:** `REDIS_URL` + `QUEUE_BACKEND=redis`, fila `confvision:eventos`.
- **Não** subir Redis novo na foxpro se `185.130.61.5` for alcançável e estável.
- Senha só no **EasyPanel Ambiente** — nunca no Git.

---

## D3 — YOLO prod (Rust)

- Paridade mínima com ex-worker: motion gate, detecção, snapshot/clip, upload S3, API Go.
- Na VPS **CPU-only**: poucas câmeras; D4 acelera depois.
- Detalhes: [FASE_D_D3_KICKOFF.md](./FASE_D_D3_KICKOFF.md).

---

## D5 / D6

- **D5:** [FASE_D_D5_KICKOFF.md](./FASE_D_D5_KICKOFF.md) — Go consome `capacity-report` / ping; assign `worker_id` → processor.
- **D6:** [FASE_D_D6_KICKOFF.md](./FASE_D_D6_KICKOFF.md) — alertas, playbooks, cron A+B+sidecar.

---

## Referências Fase C (fechamento operacional)

- Admission C3, verify: `scripts/phase-c-verify.sh`, `easypanel.env.fase-c.vps.example`.
- Monitor C2: `docs/C2_MONITOR_INTEGRACAO_EDGE.md`.
- Decisão produto: **Rust-only analítico**, worker desligado.

---

## Próximo passo imediato

**Paralelo:**

1. **D3 aceite:** 1 evento E2E na câmera piloto (22) + ajuste timeout/RAM se necessário.
2. **D5.1:** job Go lendo `/capacity-report` dos dois pilots.
3. **D6.1:** estender cron/monitor para pilot-b + sidecar; alertas `critical`/404.

Verificação rápida:

```bash
node confvision-rust-processor/scripts/d3-online-test.mjs
node confvision-rust-processor/scripts/d3-stack-check.mjs
bash confvision-rust-processor/scripts/d1-processor-verify.sh \
  https://foxpro-rust-pilot.rkr351.easypanel.host \
  https://foxpro-rust-pilot-b.rkr351.easypanel.host
```
