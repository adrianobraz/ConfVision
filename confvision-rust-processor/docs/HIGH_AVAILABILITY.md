# ConfVision — Alta Disponibilidade

Auditoria: 2026-09-30. Testes de chaos listados em **FASE-5-STORAGE-HA.md** — **não executados** neste ambiente salvo indicação contrária.

---

## 1. Classificação de componentes

| Componente | Classe | Justificativa (código) |
|------------|--------|-------------------------|
| PostgreSQL | **CRÍTICO** | Toda metadata, config, eventos |
| Rust processor (por node) | **CRÍTICO** | Ingest analítico piloto |
| MediaMTX | **CRÍTICO** | RTSP/RTMP/DVR record |
| Go Control Plane API | **IMPORTANTE** | Sync, eventos, credenciais DVR |
| Redis | **IMPORTANTE** | Fila eventos quando `QUEUE_BACKEND=redis` |
| S3 / media storage | **IMPORTANTE** | Mídia durável; pipeline analítico tolera falha parcial |
| sync-agent | **NÃO CRÍTICO** | Cache; Rust usa API direta |
| Workers Python motion/timelapse/sensor | **IMPORTANTE** (legado) | Paralelos ao Rust onde ainda ativos |
| DVR worker | **IMPORTANTE** | Gravação contínua separada |

---

## 2. Single points of failure

| Componente | SPOF? | Impacto | Mitigação atual | Mitigação futura |
|------------|-------|---------|-----------------|------------------|
| PostgreSQL | **Sim** (instância única típica) | Perda total persistência | backup host (não validado) | réplica + failover PG |
| Redis | **Sim** por fila | Capture analítico para; RTSP continua | `QUEUE_BACKEND=memory` só single-process | Redis HA ou outbox PG |
| Control Plane Go | **Sim** por região | Sync/ping/eventos falham | workers com última config em memória até timeout | CP redundante + LB |
| Storage S3 | **Sim** provider | Sem URLs; DVR backlog local | retries upload | multi-region bucket |
| Node VPS | **Sim** | Câmeras daquele node offline | manual reassign `worker_id` | failover auto + lease |
| MediaMTX | **Sim** por node | Streams locais caem | restart container | MTX clustered (complexo) |
| Rust process | **Sim** por worker_id | Analítico para naquele shard | restart systemd/docker | N replicas disjoint worker_id |

*Não inventado:* failover automático de câmera entre nodes **não existe** no código.

---

## 3. Comportamento em falha (inferido)

### 3.1 Redis OFF

| Consumidor | Comportamento |
|------------|---------------|
| Rust RTSP/YOLO | Continua (não depende Redis para decode) |
| Rust publish | DLQ / erro; métrica `errors++` |
| Rust capture | BRPOP falha — fila não drena |
| Python workers | Depende config; fila eventos similar |
| sync-agent cache | Fallback memória se `CONFIG_CACHE_BACKEND` |

### 3.2 PostgreSQL OFF

| Componente | Comportamento |
|------------|---------------|
| Go API | Falha requests |
| Rust sync/ping/evento | Erro HTTP — **não persiste eventos** |
| DVR POST segmento | Falha — arquivo local `.uploaded` pode atrasar |

### 3.3 Storage OFF

| Pipeline | Comportamento |
|----------|---------------|
| Rust capture | Evento criado; upload warn; `finalizar_evento` possivelmente sem mídia |
| DVR | Retry; arquivo permanece em disco |
| MediaMTX record | Continua escrevendo local até disco cheio |

### 3.4 Storage full

Risco: FFmpeg/MTX falha write. **Não testado.** Esperado: falhas isoladas por sub-sistema; monitorar disco `DVR_RECORD_DIR`, `CAPTURE_DIR`, `/recordings`.

### 3.5 Network Node ↔ Control Plane

Rust: sync periódico falha — **câmeras já rodando mantêm estado local** até restart (ADR-012 pull model).

---

## 4. Node failure

Cenário: Node A com N câmeras offline.

- **Detecção:** `vis_worker.ultimo_ping_em` envelhece (POST `/vis_worker_ping`).
- **Estado CP:** worker row existe; **não há** flag automática `OFFLINE` universal auditada além de `ativo` + idade ping (consumidor portal = **A DEFINIR**).
- **Recuperação automática câmeras em outro node:** **NÃO IMPLEMENTADO**.
- **Manual:** UPDATE `vis_camera.worker_id` + restart Rust destino.

---

## 5. Camera recovery (restart node)

Fluxo esperado:

```text
Node boot → Rust → GET sync ativas → CameraManager reconcile → RTSP sessions
         → POST ping (actual)
```

Depende de Postgres + API disponíveis após boot. Ordem serviços: **não há orchestrator** — Docker compose order ≠ garantia.

---

## 6. Graceful shutdown (Rust)

`main.rs`: SIGINT/SIGTERM → `manager.shutdown_signal` → workers encerram sessões RTSP.

**Não verificado:** drain completo fila Redis in-flight jobs.

---

## 7. Lease, ownership, split brain

Fase 4: ownership = `vis_camera.worker_id` + `WORKER_ID` env.

| Mecanismo | Status |
|-----------|--------|
| Lease TTL Redis | **NÃO IMPLEMENTADO** |
| Epoch fencing | **NÃO IMPLEMENTADO** |
| Split brain (dois Rust mesmo worker_id) | **RISCO CRÍTICO** — duplicidade processamento |
| Split brain (worker_id diferente mesmo câmera) | Postgres último write wins; sem lock |

**Teste lease expiration:** **NÃO APLICÁVEL** (sem lease).

---

## 8. Failover levels

| Nível | Status |
|-------|--------|
| 1 — Detecção falha | Parcial (ping, health HTTP Rust) |
| 2 — Reassign manual | Suportado via DB |
| 3 — Reassign automático | **Não** |
| 4 — Recovery completo auto | **Não** |

---

## 9. RPO / RTO

| Categoria | RPO | RTO |
|-----------|-----|-----|
| Configuração (PG) | **A DEFINIR** | **A DEFINIR** |
| Eventos metadata | **A DEFINIR** (backup PG) | **A DEFINIR** |
| Fila Redis pendente | ~0–N jobs não drenados | restart worker |
| Snapshots/clips S3 | **A DEFINIR** | **A DEFINIR** |
| DVR segmentos | backlog disco + fila upload | **A DEFINIR** |
| Timelapse | **A DEFINIR** | **A DEFINIR** |

---

## 10. Restart safety (testes)

| Teste | Resultado | Tempo recuperação | Dados perdidos |
|-------|-----------|-------------------|----------------|
| Rust restart | **NÃO EXECUTADO** | — | — |
| MediaMTX restart | **NÃO EXECUTADO** | — | — |
| Node Agent restart | N/A (agent não separado) | — | — |
| Node restart | **NÃO EXECUTADO** | — | — |
| Redis failure | **NÃO EXECUTADO** | — | — |
| PostgreSQL failure | **NÃO EXECUTADO** | — | — |
| Storage failure | **NÃO EXECUTADO** | — | — |
| Network interruption | **NÃO EXECUTADO** | — | — |
| Node failure | **NÃO EXECUTADO** | — | — |
| Storage full | **NÃO EXECUTADO** | — | — |
| Restart completo stack | **NÃO EXECUTADO** | — | — |

---

## 11. Capacidade node (observado vs inventado)

Fontes: ping `cpu_percent`, `mem_percent`, `cameras_ativas`; endpoint capacidade D5.

**Não definir** “100 câmeras/node” sem benchmark. Documentar margens após teste carga Fase 6.

---

## 12. Alertas recomendados

Consolidar (evitar duplicata com coleta existente `vis_sistema_health`):

- Node/worker ping stale
- `event_queue_redis_ok = false`
- Profundidade fila / DLQ > limiar
- Disco `DVR_RECORD_DIR` > 85%
- `stream_falhas_consecutivas` alto (PG)
- S3 upload error rate (logs `[DVR]`, `media upload falhou`)

---

## 13. Riscos abertos

1. Split brain por `worker_id` duplicado.
2. Sem purge S3/PG alinhado — custo e orphan growth.
3. Rust S3 upload sem retry vs Python.
4. Backup/restore PG não validado.
5. Failover automático ausente — RTO manual depende de ops.
