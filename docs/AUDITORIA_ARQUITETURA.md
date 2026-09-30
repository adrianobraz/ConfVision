# ConfVision — Auditoria de arquitetura (Fase 0)

**Data:** 2026-09-30  
**Repositório:** `C:\sistemaconfmonit\core4` · branch `consolidate/confvision-phase7` · commit `f982038`  
**Escopo:** inventário read-only do que existe no código e docs; **sem** alterar arquitetura.  
**Referências:** [ESTRUTURA_ECOSISTEMA.md](./ESTRUTURA_ECOSISTEMA.md), [confvision/docs/ARCHITECTURE_FINAL.md](../confvision/docs/ARCHITECTURE_FINAL.md)

---

## Legenda de classificação

| Tag | Significado |
|-----|-------------|
| **IMPLEMENTADO** | Código presente e usado no fluxo documentado |
| **PARCIAL** | Existe mas incompleto, só em worktree/legado, ou só docs/ops |
| **PLANEJADO** | Descrito em ADR/plano; sem implementação verificável |
| **NÃO ENCONTRADO** | Não localizado no `core4` canônico |

---

## Mapa de componentes

| Componente | Localização | Responsabilidade | Dependências | Estado | Observações |
|------------|-------------|------------------|--------------|--------|-------------|
| **Control Plane (Go)** | `home/confmonit/v4.0/confvision/` | Portal, API `vis_*`, CRUD câmeras, eventos, workers, RTMP auth | PostgreSQL, Redis (cache RTMP/sync), S3, Xano (opcional) | **IMPLEMENTADO** | Pool PG ~25 conexões; `visdata` + `visapi` |
| **PostgreSQL (metadata)** | `visdata/*.go`, `sql/` | `vis_camera`, `vis_worker`, `vis_evento`, stream policy, coleta | — | **IMPLEMENTADO** | Migrations em `sql/migrations/` (stream, coleta) |
| **MySQL legado** | `visdata/cliente_codigo.go`, etc. | Cliente/franqueado legado ConfMonit | Opcional (`ConexaoMySQL`) | **PARCIAL** | Só quando env configurado; não é store principal ConfVision vídeo |
| **Rust Processor** | `confvision-rust-processor/` | Analítico: sync pull, RTSP, decode, motion, YOLO, eventos, capture, health | Go API, MediaMTX RTSP, Redis (fila), S3 | **IMPLEMENTADO** | ~101 módulos `.rs`; endpoints `/health`, `/ready`, `/metrics`, `/capacity-report` |
| **CameraManager** | `src/camera/manager.rs` | Dono de workers por `camera_id`; reconcile no sync | Config, admission, shedding, stream policy | **IMPLEMENTADO** | ADR-005; sem lease global |
| **Sync pull câmeras** | `api/client.rs`, loop em `main.rs` | `GET /vis_camera_sync_ativas?worker_id=` | Go, `WORKER_ID` | **IMPLEMENTADO** | Intervalo configurável |
| **Sharding** | `src/sharding/mod.rs`, `confvision/sharding.py` | `worker_id`, hash `camera_id % total`, `MAX_CAMERAS` | Env `SHARD_*` | **IMPLEMENTADO** | Modos: worker_id, hash, auto |
| **Capacity Engine** | `src/capacity/` | Coleta CPU/RAM/GPU, estimativa `estimated_available_cameras` | `/capacity-report` | **IMPLEMENTADO** | Heurística; **medição de carga real pendente** |
| **Load Admission** | `src/load/mod.rs` (`LoadAdmissionGate`) | Bloqueia novas câmeras sem headroom | Capacity policy | **IMPLEMENTADO** | `pending_admission` no manager |
| **Load Shedding** | `src/load/shedding.rs` | Descarte sob CPU/RAM | Host metrics | **IMPLEMENTADO** | Loop em `main.rs` |
| **Frame pipeline / backpressure** | `src/pipeline/frame_pipeline.rs` | `DropOldestQueue` por câmera | `FRAME_BUFFER_MAX` | **IMPLEMENTADO** | Testes unitários |
| **Detection / YOLO scheduler (in-process)** | `src/detection/coordinator.rs` | Semaphore `yolo_max_inflight`, cooldown emit | YOLO HTTP/ONNX | **IMPLEMENTADO** | Não é scheduler cluster separado |
| **Event Queue (Redis)** | `src/events/queue.rs` | `LPUSH`/`BRPOP`, DLQ, max size | `REDIS_URL`, `confvision:eventos` | **IMPLEMENTADO** | Fallback `memory`/`none` |
| **Capture workers (Rust)** | `src/capture/workers.rs` | Consome fila → `vis_evento` + S3 | Go API | **IMPLEMENTADO** | Paralelo ao path Python legado |
| **Stream policy / auto-pausa** | `src/stream_policy/`, Go `stream_health.go` | 404/NAL → ping pausa analítico | Postgres `stream_*` | **PARCIAL** | Depende deploy Go + migrations + worker ping |
| **D5 assign (capacity-score)** | `core4-rust-pilot/.../rust_processor_d5.go` | `GET /capacity-report` → assign `worker_id` | Processors HTTP | **PARCIAL** | **Não presente** no Go canônico `core4/home/.../confvision` (só scripts + docs) |
| **Node Agent (binário)** | — | — | — | **NÃO ENCONTRADO** | Funções embutidas no Rust (sync/ping) |
| **Lease / ownership Redis** | Docs Fase 4/HA | Anti split-brain | — | **NÃO ENCONTRADO** | Ownership = `vis_camera.worker_id` + env |
| **Failover automático** | `docs/HIGH_AVAILABILITY.md` | — | — | **PLANEJADO** | Reassign manual `worker_id` |
| **Regionalização** | Docs Fase 6 | — | — | **NÃO ENCONTRADO** | Multi-tenant por `id_franqueado` em queries |
| **Python analítico legado** | `confvision/main.py`, `distributed_main.py` | YOLO + Redis + capture threads | Redis, Go | **IMPLEMENTADO** | Não usar junto com Rust no mesmo `worker_id` |
| **Python motion / timelapse / sensor / DVR** | `motion_main.py`, `dvr_main.py`, … | Gravação paralela | MediaMTX | **IMPLEMENTADO** | Responsabilidades distintas do analítico Rust |
| **sync-agent** | `sync_agent_main.py` | Cache config Redis `confvision:sync:*` | Go, Redis | **IMPLEMENTADO** | Workers Python |
| **MediaMTX** | `confvision/mediamtx/mediamtx.yml` | RTSP/RTMP/HLS/record | — | **IMPLEMENTADO** | Deploy EasyPanel |
| **RTMP Guard** | `confvision/rtmp_guard.py`, `rtmp_ban.py` | Auth publish, ban IP, stream policy | Go `rtmp_auth` | **IMPLEMENTADO** | Commit `f982038` |
| **YOLO sidecar (Python GPU)** | `yolo_gpu_worker.py`, `gpu_scheduler.py` | Inferência HTTP/GPU | — | **PARCIAL** | Rust preferencial em produção piloto |
| **XanoScript APIs** | `core4/apis/`, `tables/` | Ecossistema ConfMonit (não data plane vídeo) | Xano | **IMPLEMENTADO** | Workspace separado do processor |
| **Tenant stack / deploy** | `deploy/tenant-stack/`, `confvision/easypanel/` | Env, compose, runbooks | — | **IMPLEMENTADO** | Ops |
| **Simulador 1M câmeras** | — | — | — | **NÃO ENCONTRADO** | Fases 26–28 requerem desenvolvimento |
| **Reconciliation tool** | — | CP vs processor vs Redis | — | **NÃO ENCONTRADO** | Fase 32 gap |
| **Observabilidade distribuída (trace_id)** | Logs estruturados parciais | — | — | **PARCIAL** | Sem trace end-to-end padronizado |

---

## Fluxo de dados (resumo)

```text
NVR/DVR ──RTMP──► MediaMTX ──RTSP──► Rust (por câmera)
                         │
Portal ──HTTPS──► Go (PG) ◄──sync/ping── Rust
                         │
Rust: decode → motion → YOLO → EventJob → Redis (opcional) → Capture → vis_evento + S3
```

**Assignment:** desejado em PG (`worker_id`); execução real = processo Rust com `WORKER_ID` matching + sync filtrado.

---

## Sharding e risco de duplicata

| Mecanismo | Estado |
|-----------|--------|
| Filtro API `worker_id` | **IMPLEMENTADO** |
| Hash local `camera_id % shard_total` | **IMPLEMENTADO** (modo hash) |
| Lease / lock distribuído | **NÃO ENCONTRADO** |
| Dois processors com mesmo `WORKER_ID` | **Risco operacional** (split brain possível) |
| Dois processors com `worker_id` distintos na mesma câmera | Mitigado se PG tiver um só `worker_id` |

---

## Redis — papéis

| Uso | Chave / padrão | Produtor | Consumidor |
|-----|----------------|----------|------------|
| Fila analítica | `confvision:eventos` (+ DLQ) | Rust detection / Python | Rust capture / Python capture_workers |
| Cache sync | `confvision:sync:*` | sync-agent | Python workers |
| RTMP auth cache | (Go) | Go | RTMP Guard |

---

## Banco — papéis

| Store | Carga esperada | Gargalo potencial |
|-------|----------------|-------------------|
| PostgreSQL | CRUD portal, sync lists, eventos, pings | Pool 25, queries sync por worker |
| MySQL | Integração legado | Opcional |
| S3 | Mídia evento/DVR | Upload bandwidth |

---

## Gaps arquiteturais confirmados (não inventar)

1. **D5 Go** documentado e scriptado; **código ausente** no painel Go canônico `core4/home/...` (existe em `core4-rust-pilot`).
2. **Capacidade sustentável** não medida (ver Fase 2–15).
3. **Lease / failover / rebalance** automáticos: **PLANEJADO**, não implementado.
4. **Testes de carga Redis/PG/API**: runbooks existem; execução **pendente**.
5. **Simulação massiva (1M+)**:** NÃO ENCONTRADO**.

---

## Próximo passo (Fase 1+)

Ver [RELATORIO_ESCALABILIDADE_CONFVISION.md](./RELATORIO_ESCALABILIDADE_CONFVISION.md) e [PLANO_TESTES_ESCALABILIDADE_STATUS.md](./PLANO_TESTES_ESCALABILIDADE_STATUS.md).
