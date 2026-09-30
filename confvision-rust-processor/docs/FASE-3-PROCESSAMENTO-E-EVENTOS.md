# Fase 3 — Processamento analítico e eventos

**Componente:** `confvision-rust-processor` + API Go (`vis_evento`)  
**Pré-requisitos:** [FASE-1-RUST-CORE.md](./FASE-1-RUST-CORE.md), [FASE-2-UNIFICACAO-VIDEO.md](./FASE-2-UNIFICACAO-VIDEO.md)

---

## 0. Pré-requisitos incompletos (auditado 2026-09-30)

| Item Fase 2 | Status real | Impacto Fase 3 |
|-------------|-------------|----------------|
| Motion clip no Rust | **Não** | Eventos de gravação movimento ainda no Python |
| Timelapse no pipeline | **Stub** | Sem eventos timelapse no Rust |
| Sensor no pipeline | **Stub** | Poll sensor ainda Python |
| Shadow Luma×MOG2 medido | **Pendente** | Decisão motion não fechada |

A Fase 3 **não bloqueia** o fluxo analítico YOLO → fila → capture → Postgres, que **já existe**.

---

## 1. Arquitetura atual (código real)

```text
RTSP → decode → motion (luma) → [gate] → YOLO → evaluate_detections
                                              ↓
                                         EventJob
                                              ↓
                              Redis LPUSH confvision:eventos (ou memory)
                                              ↓
                              capture_workers BRPOP (mesmo processo Rust)
                                              ↓
                    create_evento → snapshot/clip ffmpeg → S3 → finalizar_evento
                                              ↓
                                    PostgreSQL vis_evento / vis_evento_clip
```

**Não há** “Event Engine” monolítico — responsabilidades em:

| Módulo | Papel |
|--------|--------|
| `detection/coordinator.rs` | Detection + cooldown + publish |
| `events/queue.rs` | Transporte fila |
| `events/engine.rs` | Fachada publish + logs (Fase 3) |
| `events/dedup.rs` | Cooldown por câmera |
| `capture/process.rs` | Persistência + mídia |
| `motion/session.rs` | Sessão motion (2A) — **não** cria `vis_evento` |

**Go central não consome Redis** para analítico.

---

## 2. Event model (persistido)

### 2.1 Fila interna — `EventJob` (Rust/Python idêntico)

```json
{
  "camera": { "id": 123, "...": "..." },
  "confianca": 0.87,
  "detected_at": 1735689600.123,
  "snapshot_path": "/tmp/...jpg"
}
```

- **Sem `event_id` na fila** — ID nasce no Postgres após `create_evento`.
- Correlação fila → DB: `camera_id` + `detected_at` + logs `evento_id`.

### 2.2 PostgreSQL — `vis_evento` (Go `CreateEvento`)

Campos principais (insert real): `vis_camera_id`, `id_franqueado`, `tipo_deteccao`, `confianca`, `status`, `snapshot_url`, `video_url`, `clip_count`, `processado`, …

**ID canônico:** `vis_evento.id` (integer, RETURNING).

### 2.3 Tipos lógicos (`events/catalog.rs`)

| Kind | Origem | Fila analítica? |
|------|--------|-----------------|
| `analytic_detection` | YOLO Rust | Sim |
| `sensor_pending_capture` | Receptor + poll | Não (API direta) |
| `motion_recording_segment` | motion worker | Não |
| `timelapse_segment` | timelapse worker | Não |
| `camera_operational` | health/sync | Não |

Não inventar `PERSON_DETECTED` enum na API — usar `tipo_deteccao` da câmera no Go.

---

## 3. Event lifecycle (analítico)

```text
Frame → YOLO match → cooldown OK → EventJob publish
    → capture pop → status capturando → mídia → finalizar → status pronto
```

**Incident mode** (`incident_active`): após match, YOLO pausa até motion gate desarmar; snapshots opcionais continuam.

Não há START/UPDATE/END formal no Postgres para pessoa — **1 registro por emissão** (cooldown limita repetição).

Motion sessão (2A) é paralela e **não** unificada a `vis_evento` nesta entrega.

---

## 4. Motion + YOLO (auditado)

- `ANALYSIS_ONLY_ON_MOTION` / `MotionGatedSession`: YOLO só quando gate armado ou frame com motion.
- `yolo_frame_stride`: reduz inferências.
- `yolo_max_inflight` + semáforo: backpressure YOLO.
- YOLO erro/timeout: **não** derruba RTSP (warn + skip).

---

## 5. YOLO

| Item | Código |
|------|--------|
| Backends | `off`, `http` (sidecar), `onnx` (feature) |
| Input | Luma → JPEG se HTTP; tensor se ONNX |
| Output | `PersonDetection` → regras áreas/modo |
| Timeout | `YOLO_HTTP_TIMEOUT_SEC` |
| Async | `YOLO_INFER_ASYNC` + spawn |

---

## 6. Deduplicação

| Mecanismo | Onde |
|-----------|------|
| `cooldown_seg` câmera | `EmitCooldownGate` / `DetectionContext` |
| `incident_active` | Suprime YOLO repetido na mesma incidência |
| `CaptureLocks` | 1 captura/câmera — job extra **descartado** |
| Fila cheia | publish falha, snapshot temporário apagado |

Métrica: `events_suppressed_cooldown` em `/health`.

---

## 7. Correlation

```text
detected_at + camera_id  →  (logs publish)
evento_id                →  (logs create_evento / finalizar)
snapshot S3 key          →  evento_snapshot_key(franqueado, evento_id)
clip S3 key              →  evento_clip_key(..., seq)
```

---

## 8. Snapshot / Clip

- Pré-YOLO: `write_detection_snapshot` → path no `EventJob`.
- Capture: reutiliza JPEG ou ffmpeg RTSP (`capture/process.rs`).
- Clip: `CLIP_DURACAO_SEG`, ffmpeg subprocess, upload S3, `finalizar_evento`.

---

## 9. Event Bus — decisão

**Não criar bus novo.**

| Transporte | Uso atual |
|------------|-----------|
| Redis list `LPUSH`/`BRPOP` | Fila analítica co-located |
| DLQ key | Jobs rejeitados (queue full / redis error) |
| HTTP Go | Persistência + sensor pendentes |
| Canais Tokio | Frame pipeline por câmera |

Volume: bounded `EVENT_QUEUE_MAX_SIZE`; perda aceitável em queue full (DLQ + discard).

---

## 10. Event delivery

| Etapa | Retry | Idempotência |
|-------|-------|--------------|
| Redis publish | `QUEUE_PUBLISH_RETRIES` | Não — novo job por detecção |
| Capture | loop worker | Descarta job duplicado se captura ativa |
| create_evento | single attempt | **Novo id** cada job |
| Sensor (Python) | loop 3s | Evento já existe no DB |

**At-least-once** na fila possível; consumidor não deduplica por `event_id` na fila (não existe).

---

## 11. Ordering

Ordem FIFO por fila Redis **por processo**. Multi-processor: mesma fila → competição OK se 1 captura/câmera.

Motion session ordering: local por câmera — sem bus global.

---

## 12. Redis (auditado)

- Key default: `confvision:eventos`
- DLQ: `confvision:eventos:dlq` (env)
- Backend: `none` | `memory` | `redis`
- Métricas: depth, published, queue_full, errors, redis_connected

Indisponibilidade Redis: publish → Error/DLQ; RTSP continua; capture pop falha com retry sleep.

---

## 13. PostgreSQL

- Escrita via API Go apenas (Rust não SQL direto).
- Tabelas: `vis_evento`, `vis_evento_clip`
- Sensor: `vis_evento_query_sensor_pendentes`

Sem migrações nesta fase.

---

## 14. Storage

Metadados: Postgres. Mídia: S3 (`media/upload_file`). Paths temporários: `CAPTURE_DIR`.

---

## 15. Observabilidade

Logs estruturados (Fase 3): `component=events|capture`, `operation=publish|create_evento|finalizar`, `camera_id`, `evento_id`, `detected_at`.

`/health`: `events_published`, `events_suppressed_cooldown`, `events_queue_full`, `events_captured`, `yolo_*`.

---

## 16. Backpressure

| Camada | Comportamento |
|--------|---------------|
| Frame pipeline | drop-oldest |
| YOLO | skip se busy / timeout |
| Event queue | queue full → DLQ/drop |
| Capture | lock por câmera |

---

## 17. Testes

| Teste | Status |
|-------|--------|
| `events/dedup.rs` unit | ✅ |
| Demais checklist §31–36 | ⚠️ pendente ops |

---

## 18. Alterações desta entrega Fase 3

- Doc + catálogo tipos + dedup + engine fachada + métrica cooldown + logs correlação.
- **Sem** novo Event Bus, **sem** mudança contrato `EventJob` / API Go.

---

## 19. Checklist §40 (honesto)

| Critério | Status |
|----------|--------|
| Eventos auditados | ✅ |
| Event model documentado | ✅ |
| Event ID (vis_evento.id) | ✅ documentado |
| Detection vs Event separado conceitualmente | ✅ engine fachada |
| Deduplicação cooldown | ✅ + métrica |
| Motion/YOLO/Sensor no Event Engine unificado | ⚠️ parcial (só analítico) |
| Testes carga/falha | ⚠️ |
| Fase 3 concluída | ❌ — consolidação + analítico; unificação motion/sensor pendente Fase 2 |

---

## 20. Próxima fase

**FASE 4 — Control Plane** (assignment, registry, sync) — fora deste escopo.

---

## Referências

- [ARCHITECTURE_DECISIONS.md](./ARCHITECTURE_DECISIONS.md) ADR-008, ADR-009
- `core4/confvision/event_queue.py`
- `home/confmonit/v4.0/confvision/src/modulos/visdata/eventos.go`
