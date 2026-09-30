# Fase 6 — Escala global + operação

**Auditoria:** 2026-09-30 · código em `core4-rust-pilot`, `core4/confvision`, Go `visdata`.

Princípio: **medir antes de distribuir**. Esta entrega documenta o que existe; não adiciona clusters artificiais.

---

## 1. Arquitetura real (hoje)

```text
                    CONTROL PLANE (Go visapi/visdata + PostgreSQL)
                              │
         ┌────────────────────┼────────────────────┐
         │                    │                    │
    D5 Scheduler         vis_worker           vis_mediamtx_node
    (capacity-report)    (heartbeat)          (ingest pool)
         │                    │                    │
         └──────────┬─────────┴─────────┬──────────┘
                    │                   │
              Rust worker A       Rust worker B     … Python DVR/motion/…
              WORKER_ID uniq      WORKER_ID uniq
                    │                   │
                 MediaMTX            MediaMTX
                    │                   │
                 Cameras             Cameras
                    └─────────┬─────────┘
                              │
                    Redis (fila) + S3 (mídia)
```

**Futuro compatível (não implementado):** Regional CP → pools → nodes. IDs globais PG (`vis_evento.id` serial) já são centralizados.

---

## 2. Unidade de escala

| Decisão | Valor |
|---------|--------|
| Unidade básica | **Instância `confvision-rust-processor`** identificada por `WORKER_ID` |
| Agrupamento lógico | **Node Pool** via env `RUST_PROCESSOR_BASE_URLS` (`servidor_id\|url`) e/ou `vis_mediamtx_node_id` |
| Anti-pattern evitado | Hardcode `servidor-001` no código — usar registry env + DB |

---

## 3. Node Pool (avaliação)

| Pool | Mecanismo real | Capabilities |
|------|----------------|--------------|
| Analítico CPU/GPU | Rust processors registrados D5 | `/capacity-report`, YOLO env, admission |
| Ingest RTSP/RTMP | `vis_mediamtx_node` | pontos/capacidade MTX (Go) |
| DVR | `dvr_main.py`, `worker_tipo=dvr` | gravação contínua, disco |
| Motion/timelapse | workers Python separados | OpenCV/FFmpeg, disco |

**Não criado:** tabela `vis_node_pool`. Pools são **operacionais** (env + tipo worker).

---

## 4. Scheduler (auditado)

**Implementação:** `rust_processor_d5.go`

Algoritmo (simples, previsível):

1. Listar processors do registry (`FetchProcessorCapacityScoped`).
2. Eliminar unreachable / `AllowNewCamera=false` / advisory reject.
3. Score = `estimated_available×10` + estado − câmeras online − RTSP 404.
4. `PickBestProcessor` → `AssignCameraWorkerID`.

**Não implementado:** rebalance automático, migration STOP/RELEASE/lease, scheduler multi-instância com lock.

**Flag:** `D5_AUTO_ASSIGN_ENABLED` (default on).

---

## 5. Capacity model (Rust)

| Sinal | Fonte |
|-------|--------|
| CPU/RAM/GPU/VRAM | `capacity/collector` + histórico |
| FPS / drops / latency | `CapacityEngine` + câmeras |
| `estimated_available_cameras` | `capacity/estimator.rs` (heurística, não só contagem) |
| Admission | `LoadAdmissionGate` + `allow_new_camera` |
| Shedding | `LoadSheddingCoordinator` (CPU/RAM real) |

**Camera weight:** **não implementado** — mesma heurística para todas as câmeras no estimador. Resolução/FPS entram indiretamente via FPS medido.

---

## 6. Zero duplicate processing

| Mecanismo | Eficácia |
|-----------|----------|
| `vis_camera.worker_id` + sync filter | **Um worker_id por câmera** no modelo desejado |
| Hash shard (`SHARD_MODE=hash`) | Partição dentro de um processo |
| Lease / epoch | **Ausente** — duplicar `WORKER_ID` = **bug crítico** |

Migration segura: **pendente** (Fase 5/6 gap).

---

## 7. PostgreSQL scale

| Item | Estado |
|------|--------|
| Connection pool Go | **Singleton** `sql.DB`, `MaxOpenConns=25`, `MaxIdleConns=5` (`conexao/postgres.go`) |
| N× workers Rust | HTTP para API — **não** 1 conexão PG por câmera |
| Read replica / partitioning | **Não** no código |
| Sharding dados | **Não** — multi-tenant por `id_franqueado` em queries |

**Gargalo potencial:** writes `vis_evento` + listagens; medir com `pg-audit` e slow query log (**pendente**).

---

## 8. Redis scale

| Uso | Backpressure |
|-----|----------------|
| `confvision:eventos` | `EVENT_QUEUE_MAX_SIZE`, LPUSH room check, DLQ |
| Publish retry | `queue_publish_retries` + backoff |
| Memory backend | single-process only |

**Cluster Redis / Streams:** não auditado como requisito.

---

## 9. Event pipeline (real)

```text
RTSP → decode → motion gate → YOLO → publish (Redis)
  → capture workers → POST vis_evento → PG
  → FFmpeg/S3 (opcional)
```

**Idempotência:** cooldown por câmera (`EmitCooldownGate`); **não** idempotency key global de evento na API.

**Prioridades CRITICAL/HIGH:** **não** na fila Redis atual.

---

## 10. API scale

| Item | Estado |
|------|--------|
| Rate limit | **Somente** imagem pública por IP (`rate_limit.go`) |
| Load test 100–1000 rps | **NÃO EXECUTADO** nesta fase |
| visapi + visdata | mesmo processo Go típico — horizontal = **N instâncias** atrás LB (**não deployado** no repo) |

---

## 11. Multi-tenant

Isolamento por **`id_franqueado` / `id_cliente`** em queries Go — **auditar cada rota** antes de escala 1000 clientes.

Storage DVR: bucket/credencial **por franqueado** (`gravacao_storage.py`).

---

## 12. Network / bandwidth

Modelo ( **ESTIMADO** sem medição produção):

```text
ingress_node ≈ Σ (bitrate_câmera × streams_ativos)
egress ≈ RTSP interno + upload S3 + API
```

Separar ingress câmera vs tráfego S3 em monitoramento futuro.

---

## 13. GPU / YOLO / frame policy

Documentado em Fase 1–2:

```text
RTSP FPS → FramePipeline (drop-oldest)
  → motion gate (ANALYSIS_ONLY_ON_MOTION)
  → YOLO stride / async / max inflight
```

Benchmark YOLO 1/10/50/100 câmeras: **NÃO EXECUTADO** (ambiente piloto limitado).

---

## 14. Observability

| Fonte | Conteúdo |
|-------|----------|
| Rust `/health`, `/metrics`, `/capacity-report` | câmeras, fila, capacity, phase62 |
| `vis_worker` ping | CPU/RAM, versão, cameras_ativas |
| `COLETA_RELATORIO` | `vis_sistema_health`, `vis_sistema_metric`, stream relatório dedupe |
| Dashboard global único | **Não consolidado** — dados existem em PG + JSON metrics |

**Alert dedupe:** parcial em `vis_stream_relatorio.referencia_dedupe` (30 min).

---

## 15. Versioning / deploy

Rust ping envia `processor_version`; workers Python `*_WORKER_VERSION`.

Rolling/canary/rollback: **procedimento ops** em `OPERATIONS_RUNBOOK.md` — **sem** orquestrador no repo.

---

## 16. Graceful degradation (código)

Ver Fase 5 `HIGH_AVAILABILITY.md`: CP off → Rust continua com último sync; Redis off → capture para; storage off → evento sem mídia.

---

## 17. Limites

| Tipo | Valor |
|------|--------|
| **TESTED LIMIT** | Piloto: poucas câmeras por worker (foxpro A/B) — **sem número formal documentado em benchmark** |
| **THEORETICAL ARCHITECTURAL LIMIT** | Decomposição horizontal por `WORKER_ID` + PG central; teto real = PG + Redis + S3 + rede |

---

## 18. Documentos relacionados

- [CAPACITY_PLANNING.md](./CAPACITY_PLANNING.md)
- [OPERATIONS_RUNBOOK.md](./OPERATIONS_RUNBOOK.md)
- [SCALE_TESTING.md](./SCALE_TESTING.md)
- [DISASTER_RECOVERY.md](./DISASTER_RECOVERY.md)
- [FASE-5-STORAGE-HA.md](./FASE-5-STORAGE-HA.md)

---

## 19. Checklist Fase 6 (honesto)

Documentação e auditoria: **feito**. Medidas de throughput, load/chaos test, camera weight, migration com lease, API load test: **pendente** ou **parcial**.
