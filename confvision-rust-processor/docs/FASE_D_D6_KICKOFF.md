# D6 — SLO operacional 24/7

## Em uma frase

**D6 — SLO operacional:** operação **24/7** — alertas (C2), playbooks (RTSP 404, OOM, réplica 0/1), documentação de recuperação de câmeras.

---

## Objetivo

Fechar o loop **observar → alertar → runbook** para a stack foxpro Rust-only:

- 2× rust-pilot (A/B)
- `rust-yolo-sidecar`
- Redis eventos (D2)
- MediaMTX `confvision`
- Go + Postgres (control plane)

---

## Baseline (C2 — já parcial)

| Item | Referência |
|------|------------|
| Cron verify | [C2_MONITOR_INTEGRACAO_EDGE.md](./C2_MONITOR_INTEGRACAO_EDGE.md) — CT 111, `phase-c-verify.sh` |
| Regras Prometheus | [`deploy/observability/alerts-rust-pilot.rules.yml`](../deploy/observability/alerts-rust-pilot.rules.yml) |
| Opções scrape | [`deploy/observability/README.md`](../deploy/observability/README.md) |

**Gap D6:** estender monitoramento para **piloto B**, **sidecar YOLO**, e sinais **D3** (fila/eventos).

---

## Checklist D6.1 — Monitors (Uptime Kuma / cron)

| Alvo | Check | Severidade |
|------|--------|------------|
| `foxpro-rust-pilot` | `GET /health` → `status=ok` | P1 |
| `foxpro-rust-pilot-b` | idem | P1 |
| `foxpro-rust-yolo-sidecar` | `POST /v1/detect` body mínimo → **400** ou infer → **200** (não 502) | P1 |
| `foxpro-confvision` | MediaMTX/guard up | P1 |
| Redis D2 | `event_queue_redis_ok=true` nos dois `/health` | P2 |

Script local: `node confvision-rust-processor/scripts/d3-online-test.mjs`  
Stack detalhado: `node confvision-rust-processor/scripts/d3-stack-check.mjs`

---

## Checklist D6.2 — Alertas (Prometheus ou verify --strict)

Campos prioritários (já documentados em observability):

| Sinal | Onde | Ação |
|-------|------|------|
| `rtsp_404_count > 0` | `/capacity-report` | [RUNBOOK_CAMERAS_404_ATIVO.md](./RUNBOOK_CAMERAS_404_ATIVO.md) |
| `capacity_state=critical` prolongado | `/health`, capacity-report | Reduzir câmeras, D5 assign, RAM VPS |
| `load_advisory=reject_admission` | `/health` | Não admitir câmeras; rebalance D5 |
| `frames_dropped` subindo | `/metrics` | Buffer/CPU/mem |
| Processor **réplica 0/1** (EasyPanel) | Painel / uptime | Redeploy, OOM kill |
| OOM / restart loop | Logs EasyPanel | Aumentar RAM ou reduzir `MAX_CAMERAS` / YOLO stride |
| `events_published=0` com `motion_detected>0` por N h | `/health` + `/metrics` | Investigar YOLO timeout, JPEG, zonas (D3 E2E) |
| `event_queue_depth` alto estável | `/health` | Captura/Go/S3 travado |
| Sidecar infer > 15s | Rust `YOLO_HTTP` timeout | Aumentar timeout Rust ou RAM sidecar; warm-up |

---

## Checklist D6.3 — Playbooks (doc)

| Playbook | Arquivo |
|----------|---------|
| RTSP 404 / offline | [RUNBOOK_CAMERAS_404_ATIVO.md](./RUNBOOK_CAMERAS_404_ATIVO.md) |
| Segundo processor | [deploy/phase-c/EASYPANEL_SECOND_PROCESSOR.md](../deploy/phase-c/EASYPANEL_SECOND_PROCESSOR.md) |
| Admission / capacity | [LOAD_ADMISSION.md](./LOAD_ADMISSION.md) |
| Rollback D3 | [FASE_D_D3_KICKOFF.md](./FASE_D_D3_KICKOFF.md) — `YOLO_ENABLED=0`, `CAPTURE_ENABLED=0` |

**Novo (D6):** runbook curto *“Sidecar YOLO down”* — redeploy app, testar `d3-online-test.mjs`, `YOLO_HTTP_URL` interno.

---

## Checklist D6.4 — SLO piloto (foxpro)

Metas realistas VPS CPU-only (não SLA comercial):

| SLO | Meta piloto |
|-----|-------------|
| Availability `/health` processors | > 99% (exclui deploy planejado) |
| Tempo detecção problema | < 15 min (cron */5 + alerta) |
| RTSP 404 conhecidos | 0 no capacity-report após runbook |
| Evento E2E (D3 aceite) | ≥ 1 `vis_evento` / semana em câmera piloto até prod “confiável” |

---

## Verificação D6

```bash
bash confvision-rust-processor/scripts/phase-c-verify.sh --strict-c3 \
  https://foxpro-rust-pilot.rkr351.easypanel.host
bash confvision-rust-processor/scripts/d1-processor-verify.sh \
  https://foxpro-rust-pilot.rkr351.easypanel.host \
  https://foxpro-rust-pilot-b.rkr351.easypanel.host
node confvision-rust-processor/scripts/d3-online-test.mjs
```

Atualizar cron C2 para **duas URLs** + log rotacionado.

---

## Referências

- [FASE_D.md](./FASE_D.md)
- [FASE_D_D5_KICKOFF.md](./FASE_D_D5_KICKOFF.md) — assign reduz `critical` crônico
