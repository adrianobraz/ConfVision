# Architecture Decision Records — Fase 1 (Rust Core)

Somente decisões **já refletidas no código** auditado em 2026-09-30.  
Novas ADRs devem ser append-only neste arquivo.

---

## ADR-001 — Pipeline por câmera com fila bounded drop-oldest

**Título:** FramePipeline isolada por sessão RTSP  

**Contexto:** Múltiplas câmeras no mesmo processo; RTSP pode entregar bursts; decode/YOLO são mais lentos que recepção.  

**Decisão:** Cada sessão usa `FramePipeline` com `DropOldestQueue` (capacidade `FRAME_BUFFER_MAX`). Política **drop-oldest**, nunca fila ilimitada.  

**Motivo:** Análise em tempo real prefere frame recente; evita crescimento de RAM.  

**Alternativas:** Fila global; bloquear producer; acumular todos os frames.  

**Consequências:** `frames_dropped` aumenta sob load; métricas por câmera expõem o fato.  

**Referência:** `src/pipeline/frame_pipeline.rs`

---

## ADR-002 — Fila Redis consumida dentro do mesmo processo Rust

**Título:** Event queue produtor/consumidor co-located  

**Contexto:** Worker Python usava Redis + capture threads no mesmo serviço.  

**Decisão:** `DetectionContext` faz `publish`; `capture::spawn_capture_workers` faz `pop` e chama API Go. **Go central não faz BRPOP.**  

**Motivo:** Paridade comportamental e menor latência; contrato `EventJob` unchanged.  

**Alternativas:** Consumidor no Go; fila in-memory only.  

**Consequências:** Redis é dependência de **captura** quando `capture_enabled`; indisponibilidade Redis não derruba RTSP, mas trava pipeline de eventos.  

**Referência:** `src/events/queue.rs`, `src/capture/workers.rs`

---

## ADR-003 — RTSP via retina (TCP, demux H.264)

**Título:** Cliente RTSP retina em vez de FFmpeg pull contínuo para ingest  

**Contexto:** Necessidade de demux eficiente e cancelamento cooperativo.  

**Decisão:** `connect_rtsp_demuxed` + loop demux; FFmpeg reservado a clip/snapshot pontual.  

**Motivo:** Controle fino de backpressure com `FramePipeline`.  

**Alternativas:** ffmpeg stdin pipe contínuo; GStreamer.  

**Consequências:** Dependência forte de H.264 bem formado; NAL inválido tratado em `stream_policy` / pause.  

**Referência:** `src/rtsp/session.rs`

---

## ADR-004 — Motion gate no Rust (sem confvision-motion)

**Título:** MotionGatedSession para YOLO/analysis only on motion  

**Contexto:** Fase 1 proíbe migrar serviço Python motion.  

**Decisão:** Detector luma em Rust + gate `armed`/miss streak; env `ANALYSIS_ONLY_ON_MOTION`.  

**Motivo:** Reduzir carga YOLO alinhado ao piloto.  

**Alternativas:** MOG2 OpenCV; RPC para confvision-motion.  

**Consequências:** Algoritmo diferente do motion Python legado — comparar thresholds na Fase 2 se unificar.  

**Referência:** `src/motion/gated_session.rs`, `src/motion/detector.rs`

---

## ADR-005 — CameraManager como único dono de workers por camera_id

**Título:** Sync diff start/stop workers  

**Contexto:** Sync periódico da API Go lista câmeras desejadas.  

**Decisão:** `sync_cameras` adiciona/remove tasks; map `handles` impede duplicata no mesmo processo.  

**Motivo:** Evitar dois pipelines no mesmo `camera_id`.  

**Alternativas:** Worker estático por config file.  

**Consequências:** Race curta possível durante sync rápido — mitigado por locks no map.  

**Referência:** `src/camera/manager.rs`

---

## ADR-006 — Health agregado + métricas JSON (não Prometheus embutido)

**Título:** `/health` operacional e `/metrics` com array de câmeras  

**Contexto:** EasyPanel/Kuma precisam probe simples; ADM precisa drill-down.  

**Decisão:** Axum handlers em `health/mod.rs`; Prometheus via scrape sidecar/textfile opcional.  

**Motivo:** Evitar duplicar formatos no binary.  

**Alternativas:** Só Prometheus; só `/health` plano.  

**Consequências:** Integração Prometheus requer conversão ou textfile exporter.  

**Referência:** `src/health/mod.rs`, `deploy/observability/`

---

## ADR-007 — Shadow mode Luma vs MOG2 antes de desligar motion Python

**Título:** Comparação JSONL dual-write  

**Contexto:** Fase 2 proíbe remover `confvision-motion` cedo; algoritmos luma (Rust) e MOG2 (Python) não são equivalentes.  

**Decisão:** Com `MOTION_SHADOW_COMPARE=1`, Rust escreve `rust_cam_{id}.jsonl` e Python `python_cam_{id}.jsonl` em `MOTION_SHADOW_LOG_DIR`. Script `motion_shadow_compare.py` estima concordância de `detected` em janela temporal.  

**Motivo:** Evidência antes de MOG2 no Rust ou desativação do serviço legado.  

**Alternativas:** Desligar motion e confiar só no luma; portar MOG2 imediatamente.  

**Consequências:** I/O extra em disco; serviços legados permanecem com RTSP duplicado durante shadow.  

**Referência:** `src/motion/shadow.rs`, `core4/confvision/motion_worker.py`, `docs/FASE-2-UNIFICACAO-VIDEO.md`

---

## ADR-008 — Fila Redis LPUSH/BRPOP como “event transport” analítico

**Título:** Não introduzir Event Bus novo na Fase 3  

**Contexto:** Spec menciona Event Bus; Python e Rust já usam lista Redis in-process.  

**Decisão:** Manter `EventJob` + `confvision:eventos` (+ DLQ). Go **não** consome fila.  

**Motivo:** Contrato estável, paridade Python, bounded queue + DLQ suficiente para volume analítico atual.  

**Alternativas:** Redis Streams, Pub/Sub, NATS, bus interno Tokio broadcast.  

**Consequências:** Ordering global limitado; multi-processor compartilha fila; idempotência na fila não existe (id nasce no Postgres).  

**Referência:** `src/events/queue.rs`, `confvision/event_queue.py`

---

## ADR-009 — Identificador canônico de evento = `vis_evento.id`

**Título:** event_id após persistência Go  

**Contexto:** Fila transporta detecção sem ID.  

**Decisão:** Correlação pré-DB via `camera_id` + `detected_at` + logs; ID oficial após `create_evento`.  

**Motivo:** Compatibilidade com worker Python e API existente.  

**Alternativas:** UUID no EventJob (quebra consumidores).  

**Consequências:** Rastreamento distribuído exige logs estruturados ou query Postgres.  

**Referência:** `src/capture/process.rs`, `src/events/job.rs`, `visdata/eventos.go`

---

## ADR-010 — Assignment analítico via `vis_camera.worker_id`

**Título:** Coluna worker_id como vínculo câmera → processor  

**Contexto:** Múltiplos hosts Rust A/B; necessidade de evitar todos processarem todas as câmeras.  

**Decisão:** Go persiste `worker_id` na câmera; sync API filtra por query param; D5 auto-assign escolhe processor por capacity-report.  

**Motivo:** Já em produção piloto; simples; SQL auditável.  

**Alternativas:** Tabela assignment N:N; lease Redis; sharding só hash sem DB.  

**Consequências:** Duplicidade se dois processos usarem o mesmo `WORKER_ID`; migração = UPDATE worker_id + sync.  

**Referência:** `visdata/cameras.go`, `rust_processor_d5.go`

---

## ADR-011 — Registry de nó via `vis_worker` + ping

**Título:** Heartbeat POST substitui Node Registry separado  

**Contexto:** Spec pede Node Registry; código já tem `vis_worker`.  

**Decisão:** `UpsertWorkerPing` é o registro; chave lógica `(worker_id, worker_tipo, vis_mediamtx_node_id)`.  

**Motivo:** Evitar tabela duplicada; métricas CPU/RAM já no ping.  

**Alternativas:** Nova tabela `vis_video_node`.  

**Consequências:** “Node Agent” não é entidade separada no DB.  

**Referência:** `visdata/workers.go`

---

## ADR-012 — Modelo pull (sync HTTP) sem comandos push CP→Node

**Título:** Processors puxam desired state  

**Contexto:** Avaliar WebSocket/Redis commands.  

**Decisão:** Rust loop `GET /vis_camera_sync_ativas` + reconciliação local; ping envia actual resumido.  

**Motivo:** Menor superfície; CP down não mata pipeline imediatamente.  

**Alternativas:** Push START/STOP camera.  

**Consequências:** Latência de assign = intervalo sync; drift detection não automática.  

**Referência:** `confvision-rust-processor/src/main.rs`, `camera/manager.rs`

---

## ADR-013 — Sync-agent Redis como cache opcional (não source of truth)

**Título:** Postgres+Go API > config_cache  

**Contexto:** sync-agent escreve Redis; Rust fala API direta.  

**Decisão:** Manter sync-agent para workers Python legados; verdade permanece Postgres; migrar consumidores gradualmente.  

**Motivo:** Duas fontes conflitantes seriam perigosas.  

**Alternativas:** Desligar sync-agent já; forçar todos via Redis.  

**Consequências:** Repo `core4/confvision` deve alinhar arquivos `config_cache.py` ou build context monorepo.  

**Referência:** `core4-rust-pilot/sync_agent.py`, `config_cache.py`

---

## ADR-014 — Metadata em PostgreSQL, mídia em object storage

**Título:** Separação metadata (PG) vs arquivos (S3)  

**Contexto:** Fase 5 exige não usar PG como blob store.  

**Decisão:** `vis_evento` / `vis_gravacao_segmento` guardam URLs e keys; uploads via Contabo S3 (path-style). Temp local em `CAPTURE_DIR`, `DVR_RECORD_DIR`, `MOTION_RECORD_DIR`.  

**Motivo:** Escalabilidade e custo; padrão já usado em Rust e Python.  

**Alternativas:** BYTEA em PG; NFS único sem índice.  

**Consequências:** Consistência eventual orphan/missing; lifecycle S3 é operacional.  

**Referência:** `visdata/eventos.go`, `gravacao.go`, `media/upload.rs`, `gravacao_storage.py`

---

## ADR-015 — Falha de upload não aborta registro de evento (Rust)

**Título:** Isolamento falha mídia no pipeline analítico  

**Contexto:** Storage pode falhar sob carga ou credencial.  

**Decisão:** `MediaStorage::put_or_log` + `finalizar_evento` mesmo sem URL.  

**Motivo:** Metadado de detecção vale mais que bloquear pipeline RTSP.  

**Alternativas:** Rollback evento; retry infinito.  

**Consequências:** `MISSING MEDIA` possível; portal deve tolerar URL vazia.  

**Referência:** `src/capture/process.rs`, `src/media/storage.rs`

---

## ADR-016 — Abstração MediaStorage incremental (sem migrar backend)

**Título:** Wrapper S3 no Rust, local continua permitido  

**Contexto:** Evitar paths hardcoded espalhados; preparar S3/MinIO/NAS.  

**Decisão:** Struct `MediaStorage` delega a `upload_file`; Python mantém `storage.py` / `gravacao_storage.py` até unificação futura.  

**Motivo:** Menor diff; paridade com produção Contabo.  

**Alternativas:** Trait object + drivers múltiplos na Fase 5.  

**Consequências:** Dois stacks upload (evento global vs DVR per-franqueado) permanecem.  

**Referência:** `src/media/storage.rs`

---

## ADR-017 — Retenção declarada no plano, enforcement storage incompleto

**Título:** Retencao_dias em licença/câmera sem purge automático auditado  

**Contexto:** Planos timelapse/DVR definem dias em Go (`plano.go`, `licencas.go`).  

**Decisão:** Documentar **NÃO DEFINIDO** para job de limpeza S3+disco+PG; não inventar cron.  

**Motivo:** Evitar delete indiscriminado (requisito Fase 5).  

**Alternativas:** Implementar janitor na Fase 5.  

**Consequências:** Crescimento storage até policy ops manual.  

**Referência:** `plano.go`, `dvr_segment.py` (sem delete pós-upload)

---

## ADR-018 — HA nível 1 apenas; sem failover automático de câmera

**Título:** Heartbeat + manual reassign  

**Contexto:** Spec Fase 5 proíbe failover prematuro.  

**Decisão:** Detecção via `vis_worker` ping; reassignment = alterar `worker_id` no PG; **sem** lease Redis.  

**Motivo:** Split brain e fencing não implementados.  

**Alternativas:** Lease + auto-migrate (Fase 6+).  

**Consequências:** RTO node failure = intervenção humana.  

**Referência:** `visdata/workers.go`, `FASE-4-CONTROL-PLANE.md`, `HIGH_AVAILABILITY.md`

---

## ADR-019 — Unidade de escala = instância Rust (`WORKER_ID`)

**Título:** Horizontal scale por processor, não por hostname  

**Contexto:** Fase 6 proíbe lógica fixa servidor-00N.  

**Decisão:** Cada réplica analítica tem `WORKER_ID` único; registry D5 via `RUST_PROCESSOR_BASE_URLS`.  

**Motivo:** Assignment PG já usa `worker_id`.  

**Alternativas:** Kubernetes pod name; auto UUID sem registro CP.  

**Consequências:** Colisão de `WORKER_ID` = processamento duplicado.  

**Referência:** `rust_processor_d5.go`, `control_plane/mod.rs`

---

## ADR-020 — Scheduler v1 = D5 capacity-score (sem scheduler separado)

**Título:** Assign baseado em `/capacity-report`  

**Contexto:** Spec pede scheduler simples.  

**Decisão:** Reutilizar `scoreProcessorForAssign` + `PickBestProcessor`; auto-assign opcional.  

**Motivo:** Já em produção piloto; evita segundo serviço.  

**Alternativas:** Scheduler worker com lease; Kubernetes descheduler.  

**Consequências:** Sem rebalance/migration automática; CP deve serializar assigns manuais.  

**Referência:** `visdata/rust_processor_d5.go`

---

## ADR-021 — Node Pool lógico via env e `vis_mediamtx_node`

**Título:** Pools sem tabela dedicada  

**Contexto:** CPU vs GPU vs DVR são capacidades distintas.  

**Decisão:** Pool Rust = registry D5 `servidor_id`; pool ingest = `vis_mediamtx_node`; pool DVR/motion = `worker_tipo`.  

**Motivo:** Reflete deploy atual sem migration schema.  

**Alternativas:** `vis_node_pool` table.  

**Consequências:** Documentação ops deve manter mapa pool→hosts.  

**Referência:** `mediamtx.go`, `config_cache.py` node id

---

## ADR-022 — Camera weight explícito adiado

**Título:** Capacidade via FPS/heurística agregada  

**Contexto:** 4K vs 720p não têm peso no assign.  

**Decisão:** Usar `CapacityEngine` (CPU/GPU/FPS/drops) até existir benchmark por perfil.  

**Motivo:** Evitar fórmula inventada.  

**Alternativas:** `camera_weight` column.  

**Consequências:** Assign D5 pode subestimar câmeras pesadas até MEDIDO.  

**Referência:** `capacity/estimator.rs`

---

## ADR-023 — Backpressure em camadas (pipeline + fila + shedding)

**Título:** Sem fila ilimitada  

**Contexto:** Escala exige limites.  

**Decisão:** `DropOldestQueue`; Redis `max_size` + DLQ; `LoadSheddingCoordinator` em CPU/RAM.  

**Motivo:** Já implementado Fase 1 + load policy.  

**Alternativas:** Prioridade multi-fila eventos.  

**Consequências:** Perda de frames/eventos sob overload — intencional para tempo real.  

**Referência:** `frame_pipeline.rs`, `events/queue.rs`, `load/shedding.rs`

---

## ADR-024 — Pool PostgreSQL fixo por instância Go

**Título:** Max 25 conexões abertas  

**Contexto:** Escala multi-tenant não deve abrir 1 conn/câmera.  

**Decisão:** Singleton `sql.DB` com `SetMaxOpenConns(25)`.  

**Motivo:** Padrão atual confvision Go.  

**Alternativas:** PgBouncer; pgxpool por request.  

**Consequências:** Escalar API = N instâncias × 25 conns — monitorar PG `max_connections`.  

**Referência:** `conexao/postgres.go`
