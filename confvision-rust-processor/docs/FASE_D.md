# Fase D — produto (VPS foxpro + Go/Postgres)

**Início:** 2026-09-27  
**Stack fixa:** MediaMTX (`confvision`) + **Rust** analítico. **Worker Python off.**

## Regras desta fase

| Regra | Detalhe |
|-------|---------|
| **Sem Xano novo** | Nenhuma tabela/API/função Xano. Control plane = **Go** (`https://vision.confmonit2.com.br`) + **Postgres**. |
| **D4 adiado** | GPU / GEX44 / decode NVDEC só após **servidor dedicado**. |
| **Redis existente** | `REDIS_URL` → host interno (ex. `185.130.61.5:6379`); D2 liga `QUEUE_BACKEND=redis` no Rust quando código estiver pronto. |
| **VPS realista** | Poucas câmeras **por processor** (`MAX_CAMERAS=3–4`); 2 processors no **mesmo host** = piloto de roteamento, não escala infinita. |

---

## Blocos (escopo VPS)

| Bloco | Objetivo | Status |
|-------|----------|--------|
| **D1** | N processors, `worker_id`, deploy EasyPanel | **Piloto foxpro** (A+B) — [FASE_D_D1_KICKOFF.md](./FASE_D_D1_KICKOFF.md) |
| **D2** | Fila eventos Redis, retry, DLQ | **Código pronto** — [FASE_D_D2_KICKOFF.md](./FASE_D_D2_KICKOFF.md); ligar env + redeploy |
| **D3** | YOLO / eventos / clips no Rust | Pendente dev (piloto = decode + motion) |
| **D4** | GPU | **Fora de escopo** até novo servidor |
| **D5** | Go: limites, auto-assign processor | Pendente dev Go |
| **D6** | SLO 24/7, alertas, playbooks | Parcial (C2); estender monitor D1 |

**Ordem:** D1 → D2 → D3 → (D5 ∥ D6).

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
- **Rust (a implementar):** `REDIS_URL` + `QUEUE_BACKEND=redis`, contrato fila eventos, dead-letter.
- **Não** subir Redis novo na foxpro se `185.130.61.5` for alcançável e estável.
- Senha só no **EasyPanel Ambiente** — nunca no Git.

---

## D3 — YOLO prod (Rust)

- Paridade mínima com ex-worker: motion gate, detecção, snapshot/clip, upload S3, API Go.
- Na VPS **CPU-only**: poucas câmeras; D4 acelera depois.

---

## D5 / D6

- **D5:** Go consome `capacity-report` / ping para assign `worker_id` → processor (sem Xano).
- **D6:** Alertas replica 0/1, OOM, `rtsp_404_count`, playbooks existentes (`RUNBOOK_CAMERAS_404_ATIVO.md`, etc.).

---

## Referências Fase C (fechamento operacional)

- Admission C3, verify: `scripts/phase-c-verify.sh`, `easypanel.env.fase-c.vps.example`.
- Monitor C2: `docs/C2_MONITOR_INTEGRACAO_EDGE.md`.
- Decisão produto: **Rust-only analítico**, worker desligado.

---

## Próximo passo imediato

**D2:** [FASE_D_D2_KICKOFF.md](./FASE_D_D2_KICKOFF.md) — `QUEUE_BACKEND=redis` nos pilots A/B, depois:

```bash
bash confvision-rust-processor/scripts/d2-redis-verify.sh \
  https://foxpro-rust-pilot.rkr351.easypanel.host \
  https://foxpro-rust-pilot-b.rkr351.easypanel.host
```
