# Relatório de escalabilidade — ConfVision (Fase 40 — rascunho inicial)

**Data:** 2026-09-30  
**Branch / commit:** `consolidate/confvision-phase7` @ `f982038`  
**Escopo desta entrega:** Fase 0 (inventário) + Fase 1 (build) + consolidação de gaps; **fases 2–39 não executadas** por falta de ambiente de carga.

Documentos relacionados:

- [AUDITORIA_ARQUITETURA.md](./AUDITORIA_ARQUITETURA.md)
- [PLANO_TESTES_ESCALABILIDADE_STATUS.md](./PLANO_TESTES_ESCALABILIDADE_STATUS.md)
- [BENCHMARK_PROCESSOR.md](../confvision-rust-processor/docs/BENCHMARK_PROCESSOR.md)
- [SCALE_TESTING.md](../confvision-rust-processor/docs/SCALE_TESTING.md)

---

## 1. Arquitetura atual

Control plane **Go** (PostgreSQL) + data plane **Rust** (analítico principal) + **MediaMTX** + workers **Python** (legado motion/DVR/timelapse/sensor/sync) + **Redis** (filas/cache) + **S3**. Ver diagrama em [ARCHITECTURE_FINAL.md](../confvision/docs/ARCHITECTURE_FINAL.md).

---

## 2. Componentes

Tabela completa: [AUDITORIA_ARQUITETURA.md](./AUDITORIA_ARQUITETURA.md).

---

## 3. Fluxo de dados

```text
RTMP/RTSP → MediaMTX → Rust (RTSP read) → decode → motion gate → YOLO
  → EventJob → (Redis) → capture → POST vis_evento → PG + S3
Sync: Rust ← GET vis_camera_sync_ativas (worker_id) ← Go ← PG
Heartbeat: Rust → POST vis_worker_ping → vis_worker
Assign (desejado): D5 → capacity-report → UPDATE worker_id (Go)
```

---

## 4. Distribuição de câmeras

- **Desejado:** coluna `vis_camera.worker_id` + D5 score por `/capacity-report`.
- **Implementado no processor:** sync filtra por `worker_id`; sharding hash opcional dentro do processo.
- **Gap:** handler D5 Go **não está** no tree canônico `core4/home/confmonit/v4.0/confvision` (existe em `core4-rust-pilot`).

---

## 5. Redis

- Fila `confvision:eventos` (LIST, DLQ, max size) — **implementado** no Rust.
- Cache sync Python — **implementado**.
- **Gargalo em produção:** **não medido** (Fase 8 pendente).

---

## 6. Banco

- **PostgreSQL:** metadata + eventos; pool limitado (~25).
- **MySQL:** legado opcional.
- **Gargalo:** **não medido** (Fase 9 pendente).

---

## 7. Processors

- **Rust:** analítico YOLO + capture integrado.
- **Python:** analítico legado + motion/DVR/etc.
- **Regra ops:** não duplicar analítico Rust + Python no mesmo `WORKER_ID`.

---

## 8. Sharding

- Modos `worker_id`, `hash`, `auto` — **implementado** (`sharding/mod.rs` + testes unitários).
- Truncagem `MAX_CAMERAS` — **implementado**.
- Distribuição multi-processor **equilibrada:** **não benchmarkada**.

---

## 9. Control Plane

- API `vis_*`, portal, RTMP auth, stream health — **implementado** (Go).
- Auto-assign D5 — **parcial** (script + docs; Go canônico incompleto).
- Escala de `vis_worker_ping` / sync — **não load-tested**.

---

## 10. Data Plane

- `CameraManager` + admission + shedding + per-camera pipeline — **implementado**.
- Capacidade reportada via HTTP — **implementado**; valores **não validados** em carga longa.

---

## 11. Resultados dos testes

| Suite | Status | Evidência |
|-------|--------|-----------|
| Fase 0 inventário | OK | AUDITORIA_ARQUITETURA.md |
| Fase 1 Rust check/test/release | OK | 136 tests |
| Fase 1 clippy -D warnings | FAIL | ~104 warnings |
| Fase 1 Go test/build | OK | visdata |
| Fase 1 Python pytest | SKIP | sem Python no dev host |
| Fases 2–39 carga/HA/sintético | **NE** | PLANO_TESTES_ESCALABILIDADE_STATUS.md |

---

## 12. Gargalos (hipóteses — não confirmados por benchmark)

Ordem **provável** em deploy típico (YOLO on, N câmeras altas):

1. **YOLO** (HTTP/GPU ou CPU ONNX)
2. **Decode** (CPU ou NVDEC se habilitado)
3. **RTSP rede** (WAN instável)
4. **Redis** (só se fila saturada — TBD)
5. **Go API** (sync/ping massivo — TBD)

**Confirmar apenas com Fases 2–3 e 15.**

---

## 13. Falhas encontradas (Fase 1 / inventário)

- **Clippy strict** falha por hygiene (imports mortos) — não impede runtime.
- **D5 Go** divergência entre worktree piloto e `core4` canônico — risco operacional de assign manual.
- **Sem lease:** risco de duplicata se misconfig `WORKER_ID`.

---

## 14. Limites medidos

**Nenhum limite de capacidade comprovado nesta entrega.**

---

## 15. Capacidade sustentável

**TBD** — exige Fase 2/15 em hardware representativo (≥ 30 min por degrau).

---

## 16. Capacidade máxima observada

**TBD** — não confundir com sustentável.

---

## 17. Escalabilidade horizontal

- **Teórica:** múltiplos processors com `worker_id` distintos + assign no PG.
- **Comprovada:** **não** (Fase 14 pendente).
- Overhead sync/ping/redis **não medido**.

---

## 18. Escalabilidade regional

**Não implementada** (multi-tenant lógico por franqueado apenas).

---

## 19. Limitações

- Sem simulador massivo de câmeras.
- Sem lease/failover automático.
- Benchmarks doc &gt; medição (admitido em FASE-6 docs).
- Dev Windows sem Python/GPU lab nesta sessão.

---

## 20. Gaps

| Gap | Prioridade sugerida |
|-----|---------------------|
| Merge D5 Go para tree canônico | Alta (control plane) |
| Benchmark processor + pipeline | Alta |
| Lease ou fencing anti-duplicata | Média-alta |
| Load test Redis/PG/API | Média |
| Simulador controle 100k–1M metadados | Baixa-média (planejamento) |
| Reconciliation tool | Média |

---

## 21. Roadmap (testes, não implementação)

1. Staging FoxPro: smoke `/health`, `/capacity-report`, assign D5 (após merge Go).
2. Fase 2 ladder 1→100 câmeras, 30 min cada, preencher BENCHMARK_PROCESSOR.md.
3. Fase 3 A–D (profiling por camada).
4. Fase 7 documentar comportamento split-brain (teste controlado).
5. Fases 8–9 redis-benchmark / pgbench isolado + app queries.
6. Fase 14 scale-out 1→4 processors com mesma carga/câmera count.

---

## Respostas finais (18 perguntas) — estado honesto

| # | Pergunta | Resposta |
|---|----------|----------|
| 1 | Capacidade atual de 1 processor? | **Desconhecida (não benchmarkada).** |
| 2 | Capacidade sustentável? | **TBD.** |
| 3 | Gargalo? | **Hipótese: YOLO/decode** — **não confirmado.** |
| 4 | Scale-out funciona? | **Parcialmente desenhado** (`worker_id`); **não provado.** |
| 5 | Sharding funciona? | **Sim em lógica local** (unit tests); **não provado em cluster.** |
| 6 | Risco câmera duplicada? | **Sim**, se dois processors mesmo `WORKER_ID` ou erro PG — **sem lease.** |
| 7 | Ownership? | **Fraco:** `worker_id` PG + convenção env; **sem lease Redis.** |
| 8 | Failover? | **Manual** (reassign); **sem automático.** |
| 9 | Redis é gargalo? | **Não medido.** |
| 10 | Banco é gargalo? | **Não medido.** |
| 11 | YOLO é gargalo? | **Provável** em carga analítica; **não medido.** |
| 12 | Control Plane escala? | **Não load-tested.** |
| 13 | Event Plane escala? | **Não load-tested.** |
| 14 | Regionalização existe? | **Não.** |
| 15 | Preparado para milhões? | **Não** — só particionamento lógico (`worker_id`, franqueado); **sem registry/lease/simulador.** |
| 16 | O que falta desenvolver? | Lease/failover, D5 no Go canônico, benchmarks, simulador CP, reconciliation, regional (se produto exigir). |
| 17 | Maior limitação arquitetural atual? | **Assignment/ownership fraco + ausência de medição de capacidade real.** |
| 18 | Testes que faltam? | **Fases 2–39** (ver PLANO_TESTES_ESCALABILIDADE_STATUS.md). |

---

## Regra de capacidade (cumprimento)

Este relatório **não** afirma suporte a N câmeras. Qualquer número futuro deve citar run id, hardware, duração ≥ 30 min, drops, latência e limiting_resource do `/capacity-report`.
