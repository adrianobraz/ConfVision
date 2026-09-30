# Fase 1 — Consolidação do Rust Processing Core

**Projeto:** ConfVision  
**Componente:** `confvision-rust-processor`  
**Branch auditada:** `rust-pilot` (commit de referência no momento da auditoria: ver `git log -1`)  
**Escopo:** consolidar o núcleo por câmera. **Fora de escopo:** motion/timelapse/sensor/dvr/sync-agent Python, Control Plane, MOG2, reescrita YOLO.

---

## 1. Estado inicial

| Item | Situação |
|------|----------|
| `confvision-worker` (Python) | Descontinuado na operação piloto |
| Processamento analítico | `confvision-rust-processor` (EasyPanel foxpro A/B) |
| Serviços legados | `confvision-motion`, `timelapse`, `sensor`, `sync-agent`, `dvr` — **não migrados nesta fase** |
| Testes unitários Rust | **132** testes (`cargo test`) — OK na auditoria |
| Documentação prévia | Fases piloto A/C/D, ARQUITETURA_RUST.md — complementar, não substituto deste doc |

---

## 2. Arquitetura atual (código real)

### 2.1 Módulos (`src/`)

| Módulo | Responsabilidade |
|--------|------------------|
| `main.rs` | Boot, HTTP, sync/ping loops, graceful shutdown |
| `config.rs` | Env global (`Config::from_env`) |
| `camera/` | **CameraManager**, estados, backoff, URL RTSP redigida |
| `worker/` | **run_camera_worker** — loop RTSP → pipeline por câmera |
| `rtsp/` | Sessão **retina** (demux H.264, TCP) |
| `pipeline/` | **FramePipeline** bounded, drop-oldest, consumer async |
| `decode/` | CPU / FFmpeg opcional, aceleração 4.1 |
| `motion/` | Detector + **MotionGatedSession** (gate, não MOG2 externo) |
| `detection/` | YOLO + regras + **publish Redis** |
| `yolo/` | HTTP sidecar e/ou ONNX (feature) |
| `events/` | **EventQueue** (none / memory / redis) |
| `capture/` | Workers BRPOP → **processar_deteccao** → API Go |
| `api/` | Cliente ConfVision (sync, ping, evento, S3 upload) |
| `analytics/` | Runtime compartilhado YOLO + fila + stats detecção |
| `stream_policy/` | 404, pausa analítica, backoff de probe |
| `load/` | Admission + shedding (CPU/RAM) |
| `capacity/` | Estimativa câmeras / capacity-report |
| `health/` | `/health`, `/ready`, `/metrics`, `/capacity-report` |
| `metrics/` | Contadores processador |
| `sharding/` | worker_id / shard sync |
| `redis/` | Log + ping startup fila |

### 2.2 Structs e tasks principais

- **CameraManager** — mapa `camera_id → SharedCameraState`, `JoinHandle` por câmera, sync com API Go, admission, stream health reports.
- **SharedCameraState** — `RwLock<CameraRuntimeState>` **por câmera** (isolamento de métricas/estado).
- **FramePipeline** — fila `Mutex<DropOldestQueue>` **por sessão/câmera** + task consumer.
- **DetectionContext** — por câmera (via spawn async YOLO).
- **Capture workers** — N tasks globais `capture_worker_loop` (compartilham **uma** fila Redis — by design, paridade Python).

### 2.3 Endpoints HTTP

| Rota | Tipo |
|------|------|
| `GET /health` | Liveness + readiness operacional + resumo câmeras + fila Redis + YOLO |
| `GET /ready` | Readiness mínimo (`api_ready`) |
| `GET /metrics` | JSON agregado + **array `cameras[]`** com `CameraRuntimeState` |
| `GET /capacity-report` | Admission / load / estimativas |

---

## 3. Fluxo real (configuração → evento)

```text
Config::from_env + dotenv
    ↓
bootstrap EventQueue (none|memory|redis) + YoloRuntime + AnalyticsRuntime
    ↓
spawn capture_workers (pop fila → processar_deteccao → API Go)
    ↓
CameraManager + sync_loop (GET câmeras ativas por worker_id)
    ↓
sync_cameras: start/stop run_camera_worker por camera_id
    ↓
run_camera_worker loop:
    Reconnecting → connect_rtsp_demuxed (retina)
    ↓
run_rtsp_demux_loop → FramePipeline::push (drop-oldest)
    ↓
run_frame_consumer: decode → motion → (optional) DetectionContext::on_decoded_frame
    ↓
YOLO match + cooldown → EventQueue::publish (Redis LPUSH)
    ↓
capture_worker pop → create_evento / ffmpeg clip / S3 → finalizar_evento
```

**Importante:** o **Go central não consome Redis**. A fila é **interna ao processo Rust** (produtor = detecção, consumidor = capture workers), igual ao worker Python.

---

## 4. Dependências

| Camada | Uso |
|--------|-----|
| **retina** | RTSP client, demux H.264 |
| **ffmpeg-next** (feature `ffmpeg-decode`) | Decode HW/SW opcional |
| **ort** (feature `yolo-onnx`) | YOLO local opcional |
| **redis** (feature `tokio-comp`) | Fila `confvision:eventos` |
| **reqwest** | API Go + YOLO HTTP sidecar |
| **image** | JPEG snapshot |
| **axum** | HTTP admin |
| **tokio** | Runtime async, 1 task RTSP + 1 consumer por sessão ativa |
| **MediaMTX** | RTSP `rtsp://…/cam/{hash}` (via `MEDIAMTX_RTSP_BASE` + câmera) |
| **Filesystem** | Snapshots temporários em `capture_dir` |

---

## 5. Camera lifecycle

Estados implementados (`camera/types.rs` — `CameraStatus`):

```text
Starting → Online ⇄ Reconnecting → Offline | Error → Stopped
```

Não há enum `DISABLED` explícito: câmera removida do sync → worker parado (`Stopped`).  
**Stream policy** pode pausar analítico localmente (`stream_local_paused`) sem destruir o worker.

Gerenciamento: `CameraManager::sync_cameras` diff desired vs handles — evita duplicar worker para mesmo `camera_id`.

---

## 6. Camera Manager

Responsabilidades **já presentes** em `camera/manager.rs`:

- Iniciar/parar/reiniciar via sync
- `handles: HashMap<i64, CameraWorkerControl>`
- Admission (`LoadAdmissionGate`) e shedding (`load_shed_ids`)
- `pending_admission` para câmeras sem headroom
- Stream health → fila para worker ping Go
- **Não** processa frames (delega ao worker)

---

## 7. RTSP

- Biblioteca: **retina**, transporte **TCP**.
- Timeout/cancel: `CameraCancel`, checagens periódicas no demux loop.
- Reconexão: loop em `camera_worker` com **ReconnectBackoff** (`camera/backoff.rs`).
- URL em logs: **`redact_rtsp_url`** (`camera/stream_url.rs`).
- Simulação: `RTSP_SIMULATE` para testes sem rede.

---

## 8. Reconexão

Fluxo em `worker/camera_worker.rs`:

```text
RTSP falha / sessão encerra
    ↓
status Reconnecting, increment reconnect_count
    ↓
stream_policy (404, NAL inválido, etc.) → pode pausar ou reportar Go
    ↓
backoff.sleep()
    ↓
retry connect
```

Métricas: `reconnect_count`, `rtsp_errors`, `stream_failures_consecutive` em `CameraRuntimeState`.

---

## 9. Decoder

- Default: **CPU** luma path + H264 decoder por sessão (`SessionDecodeContext`).
- Opcional: **ffmpeg-decode**, **ffmpeg-nvdec**, política em `DecodePolicyCoordinator`.
- Métricas por câmera: `frames_decoded`, `decode_errors`, `last_decode_ms`.

---

## 10. Frame pipeline

- **Bounded queue** + **drop-oldest** (`DropOldestQueue` — testes unitários).
- Producer: thread RTSP/demux; consumer: task tokio dedicada por sessão.
- Contadores: `frames_enqueued`, `frames_dropped`, `buffer_full_events`, `frames_processed`.
- **Sem fila global de frames** entre câmeras.

---

## 11. Isolamento entre câmeras

| Isolado por câmera | Compartilhado (global) |
|--------------------|-------------------------|
| FramePipeline | ProcessorMetrics (atomics) |
| CameraRuntimeState (RwLock) | EventQueue Redis (intencional) |
| Task RTSP + consumer | Capture workers (pool N) |
| DetectionContext | YoloRuntime HTTP (semáforo inflight) |
| StreamPolicyState | Sync loop, capacity engine |

**Risco conhecido:** sidecar YOLO lento → `yolo_skipped_busy`; **não** bloqueia RTSP das outras câmeras.  
**Risco conhecido:** shedding/admission CPU global pode **parar** câmeras sob load (política operacional, não bug de pipeline).

---

## 12. Motion (Fase 1 — gate only)

- **Não** migrar `confvision-motion` Python.
- Algoritmo: `motion/detector.rs` (diff luma, sensibilidade config).
- Gate: `MotionGatedSession` + `MotionEnqueueGate` + `ANALYSIS_ONLY_ON_MOTION`.
- Documentado em env: `MOTION_*`, `MOTION_GATE_MISS_FRAMES`.

---

## 13. YOLO

- Modos: `YOLO_BACKEND=http` (sidecar) ou `onnx` (feature).
- Timeout async: `YOLO_HTTP_TIMEOUT_SEC`, `YOLO_INFER_ASYNC`, `YOLO_MAX_INFLIGHT`.
- Falha YOLO: log + skip frame; processo e RTSP continuam.

---

## 14. Snapshot

- `capture/snapshot.rs`, `write_detection_snapshot` antes do enqueue.
- Path local; upload S3 na fase capture via API Go.

---

## 15. Clip

- `capture/ffmpeg.rs` — RTSP secundário para clip; Rust **coordena**, FFmpeg executa.
- Não é encoder contínuo tipo DVR.

---

## 16. Eventos

- **EventJob** (`events/job.rs`) — paridade Python `EventJob.to_dict`.
- Publish: após detecção + cooldown.
- Consumo: capture workers → `create_evento` / `finalizar_evento` (API Go).
- Retry/DLQ: `events/queue.rs` (`confvision:eventos:dlq`).

---

## 17. Redis

- Backends: `none`, `memory`, `redis`.
- `QUEUE_BACKEND=redis` exige `REDIS_URL`.
- Indisponibilidade: publish → erro/DLQ; **pop** log + sleep 500ms (capture loop).
- Startup: `redis::startup_redis_check` + loop métricas depth 30s.

---

## 18. Health

| Endpoint | Semântica |
|----------|-----------|
| `/health` | Processo vivo + câmeras + capacity + fila + YOLO |
| `/ready` | HTTP/API bootstrapped |
| `/metrics` | Detalhe por câmera (`cameras[]`) |

---

## 19. Metrics (por câmera em `/metrics`)

Campos em `CameraRuntimeState` cobrem: status, uptime implícito (`started_at`), reconnects, frames received/dropped/processed, decode, motion, buffer, stream policy.

Contadores globais detecção/capture: `/health` → `yolo_inferences`, `events_published`, `events_captured` (via `AnalyticsRuntime` / capture stats).

Prometheus textfile: ver `deploy/observability/` (opcional, não duplicado no core).

---

## 20. Logging

- `tracing` + `RUST_LOG` / `LOG_LEVEL`.
- Campos estruturados frequentes: `camera_id`, `processor_id`, `worker_id`.
- RTSP redigido; revisar logs de erro third-party (FFmpeg) para vazamento de URL — **melhoria futura** se aparecer credencial em stderr.

---

## 21. Configuração

- **Global:** `config.rs` — centenas de env documentados em `.env.example` / easypanel examples.
- **Por câmera:** payload JSON do sync Go (`CameraRecord`) — áreas, confiança, flags analítico, stream policy generation.

---

## 22. Segurança

- `redact_rtsp_url` no estado exposto.
- Não logar `VIS_WORKER_API_KEY` / `REDIS_URL` no boot (redis loga só host lógico via `log_redis_status`).

---

## 23. Graceful shutdown

`main.rs`:

```text
SIGINT / SIGTERM (unix)
    ↓
manager.shutdown_signal → stop_all
    ↓
abort sync, ping, HTTP
    ↓
send_final_ping
```

Workers respeitam `watch` shutdown na sessão RTSP.

---

## 24. Resource management

- `frame_buffer_max`, pool global opcional (`buffer/pool.rs`).
- `MAX_CAMERAS`, admission, shedding.
- jemalloc feature opcional.

---

## 25. Testes

| Tipo | Onde |
|------|------|
| Unitários | pipeline, motion, queue, backoff, config, decode, stream_policy, … |
| Integração RTSP | `RTSP_SIMULATE`, testes em `camera_worker` / manager |
| Falha | stream_policy tests, queue full, redis mock patterns |
| Concorrência | `frame_pipeline` drop-oldest sob carga simulada |

**Comando:** `cargo test` → **132 passed** (auditoria 2026-09-30).

---

## 26. Stress tests

Scripts operacionais (não CI automático):

- `scripts/c1-ab-baseline.sh`, capacity-report, piloto env `MAX_CAMERAS`.
- Capacidade **medida** em foxpro: ~2–3 cams CPU com `critical` admission — ver `/health` produção.

Não afirmar 100 câmeras sem benchmark dedicado (Fase 4.1 / GPU).

---

## 27. Problemas encontrados (auditoria)

1. **`events_published=0`** em produção com YOLO ativo → gargalo **antes** da fila (`yolo_matches=0`), não Redis.
2. **Redis central 185.130.61.5** pode estar vazio (`KEYS confvision:*`) enquanto `/health` reporta redis ok → confirmar **mesmo REDIS_URL** no EasyPanel A/B.
3. **`/ready`** não verifica sync Go bem-sucedido — só flag pós-bind HTTP.
4. **Capture workers globais** — lock por câmera em `CaptureLocks`; job descartado se captura em andamento (by design).
5. WIP no repo: muitos arquivos modificados não relacionados à Fase 1 no working tree — não misturar commits.

---

## 28. Problemas não resolvidos nesta fase

- Eventos E2E produção (pessoa + publish + capture + Postgres).
- Stress 25–100 câmeras documentado com números.
- Liveness/readiness Kubernetes-style estrito (sync falhou permanente).
- Métricas Prometheus nativas no binary (hoje JSON `/metrics`).

---

## 29. Riscos

- CPU decode denso → admission `reject` → sensação de “câmera caiu”.
- YOLO sidecar lento → menos inferências, zero eventos.
- Dois processors A/B no **mesmo** Redis sem coordenação de shard → competição na fila (OK se workers distintos por câmera).

---

## 30. Alterações realizadas nesta entrega Fase 1

| Ação | Detalhe |
|------|---------|
| Auditoria código | Este documento |
| ADRs | `ARCHITECTURE_DECISIONS.md` (decisões já embodied no código) |
| Código | **Nenhuma reescrita** — consolidação = documentação + validação testes |

*(Implementações futuras devem ser commits pequenos separados.)*

---

## 31. Checklist de conclusão (§33 do briefing)

| Critério | Status |
|----------|--------|
| Rust auditado | ✅ |
| Camera Manager consolidado | ✅ (já existia; documentado) |
| lifecycle definido | ✅ (`CameraStatus`) |
| RTSP robusto | ✅ (retina + policy); melhorias incrementais possíveis |
| reconexão robusta | ✅ backoff + stream_policy |
| decoder validado | ✅ CPU + optional FFmpeg |
| frame pipeline validado | ✅ drop-oldest + testes |
| isolamento entre câmeras | ✅ validado em design; shedding global documentado |
| Motion preservado | ✅ gate only |
| YOLO preservado | ✅ |
| eventos preservados | ✅ contrato EventJob |
| Redis preservado | ✅ |
| health validado | ✅ |
| métricas disponíveis | ✅ `/metrics` per cam |
| logs estruturados | ⚠️ parcial (tracing ok; padronização contínua) |
| graceful shutdown | ✅ |
| gerenciamento recursos | ✅ admission/shedding/bounded queues |
| testes unitários | ✅ 132 |
| testes integração | ⚠️ simulate + ops scripts |
| testes falha | ⚠️ parcial unit |
| teste múltiplas câmeras | ⚠️ foxpro 2–3 cams medido |
| documentação | ✅ este doc |
| Fase 2 não migrada | ✅ |

**Fase 1 documental/consolidação:** concluída.  
**Fase 1 “núcleo confiável em produção E2E eventos”:** pendente validação operacional (eventos + stress).

---

## 32. Próxima fase

**FASE 2 — Unificação do vídeo** (fora deste escopo): evolução motion/algoritmo, possivelmente integração timelapse/dvr — **somente após** E2E eventos estável no Rust core.

---

## Referências

- `docs/ARCHITECTURE_DECISIONS.md`
- `docs/ARQUITETURA_RUST.md`
- `docs/FASE_D_D2_KICKOFF.md`, `FASE_D_D3_KICKOFF.md`
- `deploy/tenant-stack/` (escala multi-tenant — não confundir com Fase 1 core)
