# Plano técnico de migração — ConfVision × Rust

**Status:** validação pré-implementação (nenhum serviço desligado, nenhum código alterado neste passo).

**Base:** [AUDITORIA_MIGRACAO_SERVICOS_RUST.md](./AUDITORIA_MIGRACAO_SERVICOS_RUST.md) (§22 validação), código em `core4/confvision/` e `core4-rust-pilot/confvision-rust-processor/`.

---

## 1. Matriz de decisão (função × tecnologia)

| Função | Atual | Rust hoje | Decisão | Pré-requisito |
|--------|-------|-----------|---------|---------------|
| RTSP analítico | worker (off) / rust | Sim (`rtsp/`, `camera_worker.rs`) | **Manter Rust** | Assignment por câmera; 1 RTSP por owner |
| Decode analítico | Rust ffmpeg / retort | Sim (`decode/`) | **Manter Rust** | Feature `ffmpeg-decode` no deploy |
| Motion gate analítico | Rust luma | Sim (`motion/detector.rs`) | **Manter Rust** | Tunagem env `MOTION_*_THRESHOLD` |
| MOG2 gravação | motion, timelapse, worker legacy | **Não** | **Não substituir pelo gate luma sem testes** | Paridade A/B ou módulo MOG2/ffmpeg gravação |
| YOLO inferência | Rust HTTP (+ ONNX opcional) | HTTP sim; ONNX com feature | **Manter HTTP sidecar** | `YOLO_HTTP_URL`, admission CPU |
| Snapshot evento | Rust + sensor Python | Sim (`capture/snapshot.rs`) | **Rust** (sensor via mesmo pipeline) | S3 env no processor |
| Event clip analítico | Rust + sensor Python | Sim (`capture/ffmpeg.rs`) | **Rust** | Evitar 2º RTSP (frame buffer ou owner único) |
| Motion clip gravação | motion Python | **Não** | **Migrar depois** (Rust módulo ou Go+frames) | `process_segment_file` + `tipo=movimento` |
| Timelapse | timelapse Python | **Não** | **Migrar depois** | Frames esparsos + concat; sem 2º RTSP |
| Sensor poll | sensor Python | **Não** | **Go control plane / bus** | Substituir poll HTTP ad hoc |
| Sensor capture | event_capture Python | Parcial (`capture/process.rs`) | **Absorver no Rust owner** | Job por evento_id; sem RTSP duplicado |
| DVR record | MediaMTX | **Não** (correto) | **Manter MTX** | — |
| DVR watcher + S3 + API | dvr Python | **Não** | **Go uploader** (ou manter Python) | Idempotência `vis_gravacao_segmento` |
| S3 upload evento | Rust `media/upload.rs` | Parcial (env) | **Rust** para analítico | Credenciais / env alinhados ao Python |
| S3 upload gravação | gravacao_storage Python | **Não** | **Go ou Python** até migração | Mesmas keys/bucket |
| Redis fila eventos | Rust | Sim (`events/queue.rs`) | **Rust** | `QUEUE_BACKEND=redis` |
| Redis config sync | sync-agent Python | **Não** | **Go CP** | Deprecar `get_cameras_*_cached` |
| Events API | Rust coordinator + capture | Sim | **Rust** analítico | Fila + DLQ monitorada |
| Health / metrics | Rust | Sim (`/health`, `/metrics`) | **Rust** | Sidecar/probes EasyPanel |
| Metrics Python workers | ping apenas | N/A | **Manter ping** até desligar | — |

---

## 2. NÃO migrar para Rust

| Item | Justificativa |
|------|----------------|
| **Gravação contínua DVR (fMP4)** | MediaMTX já grava no filesystem; Rust não ganha decode aqui — só I/O e API. |
| **Watcher filesystem + upload segmento contínuo** | Trabalho I/O-bound, retries, idempotência — **Go** (ou Python atual) encaixa melhor que monolito Rust de vídeo. |
| **POST `vis_gravacao_segmento`** | Contrato de gravação comercial; separar do pipeline analítico reduz blast radius. |
| **Poll `vis_evento_query_sensor_pendentes`** | Control plane / fila de jobs — não pertence ao hot path de decode. |
| **Sync-agent / cache Redis de config** | Múltiplos consumidores, versionamento — **Go API + assignment**; Rust já faz sync HTTP próprio só para câmeras assignadas. |
| **RTMP guard / auth MediaMTX** | Python leve no container MTX; sem ganho claro em Rust. |
| **MOG2 como requisito de produto** | Não portar “só porque Rust existe”; gate luma **≠** MOG2 sem validação. |
| **Timelapse batch (concat horas)** | Job assíncrono de mídia, não latência de detecção — evitar acoplar ao loop RTSP analítico. |
| **Inferência YOLO centralizada GPU** | Sidecar HTTP/Python GPU cluster pode permanecer mesmo com Rust analítico. |
| **Assignment / lease / failover** | Estado distribuído — **Go + Postgres**, não Rust processor. |

**Princípio:** Rust = **data plane analítico** (1 RTSP owner, decode, gate, YOLO, evento + capture curto). Gravação comercial, upload em massa e control plane ficam fora.

---

## 3. Ordem de fases (por dependência real)

A ordem sugerida pelo produto (Motion → Timelapse → …) **não** é a ordão técnica correta: **motion/timelapse exigem paridade inexistente no Rust** e **duplicam RTSP**. Ordem abaixo:

```text
FASE 0 → FASE 1 → FASE 2 → FASE 3 → FASE 4 → FASE 5 → FASE 6 → FASE 7
validação   analítico   capture    CP/sync    1 RTSP    sensor     motion/TL   DVR Go
```

---

### FASE 0 — Validação (atual)

| Campo | Conteúdo |
|-------|----------|
| **Objetivo** | Confirmar mapa código × função; congelar critérios de desligamento. |
| **Componentes** | Docs, piloto cam 15, métricas produção. |
| **Arquivos** | `AUDITORIA_*`, este plano; sem alteração de runtime. |
| **APIs** | Baseline: `vis_worker_ping`, eventos piloto, gravação se apps on. |
| **Riscos** | Decidir desligar motion cedo → perda gravação. |
| **Testes** | Checklist §22 auditoria; MOG2 vs luma A/B (desenho). |
| **Rollback** | N/A |
| **Sucesso** | Stakeholders aceitam matriz §1 e “NÃO migrar” §2. |
| **Desligar serviço** | Nenhum |

---

### FASE 1 — Analítico Rust em produção (substituir worker off)

| Campo | Conteúdo |
|-------|----------|
| **Objetivo** | 100% câmeras analíticas no Rust; eventos + fila estáveis. |
| **Componentes** | `confvision-rust-processor`, sidecar YOLO, Redis fila. |
| **Arquivos** | `core4-rust-pilot/confvision-rust-processor/src/**`, env EasyPanel. |
| **APIs** | `vis_camera_sync_ativas` / sync ativas, `vis_evento`, `vis_worker_ping`, stream health. |
| **Riscos** | CPU/admission; `events_published=0`; S3 ausente. |
| **Testes** | `/health`, `/ready`, `/metrics`, `/capacity-report`; evento humano por câmera. |
| **Rollback** | Reassign `worker_id`; reativar worker Python (se imagem existir). |
| **Sucesso** | Nenhuma câmera analítica órfã; DLQ vazia; latência capture OK. |
| **Desligar** | `confvision-worker` — já off; critério: paridade evento 7 dias piloto. |

---

### FASE 2 — Paridade capture analítico (snapshot/clip/S3)

| Campo | Conteúdo |
|-------|----------|
| **Objetivo** | Mídia de evento analítico 100% Rust; reduzir 2º RTSP na capture. |
| **Componentes** | `capture/process.rs`, `media/upload.rs`, ffmpeg subprocess. |
| **Arquivos** | `capture/*`, env S3 alinhado a `gravacao_storage` / evento keys. |
| **APIs** | `vis_evento`, `vis_evento_finalizar`. |
| **Riscos** | Segundo RTSP por evento sob carga; timeout clip. |
| **Testes** | Comparar snapshot/clip vs baseline Python; carga  N eventos/min. |
| **Rollback** | `CAPTURE_ENABLED=0` ou flag equivalente. |
| **Sucesso** | URLs S3 + finalizar OK; sem regressão qualidade. |
| **Desligar** | Nenhum serviço Python ainda. |

**Pré-requisito futuro (ideal):** reutilizar frame decodificado ou stream interno — hoje capture abre RTSP de novo (`capture/process.rs`).

---

### FASE 3 — Control plane: sync, assignment, cache

| Campo | Conteúdo |
|-------|----------|
| **Objetivo** | Uma fonte de verdade de câmeras por nó; eliminar dependência `sync_agent` Python. |
| **Componentes** | Go API (novo), Redis opcional; deprecar `sync_agent_main.py`. |
| **Arquivos Python ref.** | `xano_client.py` (`_cache_enabled`), `core4-rust-pilot/sync_agent.py`, `config_cache.py` (fora de `confvision/`). |
| **APIs** | `vis_camera_sync_ativas`, endpoint assignment/lease (novo), ping. |
| **Riscos** | Stale config; workers com `CONFIG_CACHE_BACKEND=redis` sem agente → import quebrado. |
| **Testes** | Failover nó; versão sync; Rust + DVR + motion leem assignment. |
| **Rollback** | Restaurar sync-agent + keys Redis legadas. |
| **Sucesso** | Nenhum worker precisa `from sync_agent import ...`. |
| **Desligar** | `confvision-sync-agent` quando Go cache + assignment OK. |

**Repo:** copiar ou remover COPY quebrado em `Dockerfile-mediamtx` / `Dockerfile.mediamtx` — decisão de build separada da migração Rust.

---

### FASE 4 — Política “1 RTSP por câmera” (fan-out)

| Campo | Conteúdo |
|-------|----------|
| **Objetivo** | Impedir Rust + motion + timelapse simultâneos na mesma câmera sem design explícito. |
| **Componentes** | Assignment `modo_gravacao` vs analítico; shard rules. |
| **Arquivos** | `sharding.py`, config câmera Postgres, docs escala. |
| **APIs** | Assignment / modos gravação. |
| **Riscos** | CPU 98% (observado piloto) se ignorado. |
| **Testes** | Inventário câmeras com >1 consumer RTSP; métricas MTX readers. |
| **Rollback** | Pausar motion/timelapse por franqueado. |
| **Sucesso** | Zero overlap não documentado em produção. |
| **Desligar** | Nenhum — preparação para FASE 5–6. |

---

### FASE 5 — Sensor (dispatch + capture no owner)

| Campo | Conteúdo |
|-------|----------|
| **Objetivo** | Latência sensor ≤ baseline; capture sem RTSP extra quando câmera já no Rust. |
| **Componentes** | Go consumer pendentes **ou** Redis stream; Rust `capture/process` com `evento_id` existente. |
| **Arquivos ref.** | `sensor_main.py`, `event_capture.py`, `capture/process.rs`. |
| **APIs** | `vis_evento_query_sensor_pendentes`, `get_camera_by_id`, `vis_evento_finalizar`. |
| **Riscos** | Evento duplicado processado; captura concorrente analítico. |
| **Testes** | Alarme real/simulado; mídia + status `pronto`. |
| **Rollback** | `sensor_main.py` poll. |
| **Sucesso** | Sensor sem Python; ou Python só fallback 30 dias. |
| **Desligar** | `confvision-sensor` |

**Nota:** Rust **não** substitui poll hoje; só capture após job.

---

### FASE 6 — Motion gravação + Timelapse

| Campo | Conteúdo |
|-------|----------|
| **Objetivo** | Clipes movimento + timelapse **sem** 2º/3º RTSP; POST segmento intacto. |
| **Componentes** | Novo módulo Rust **ou** Go consumindo frames; reutilizar `dvr_segment` contract. |
| **Arquivos ref.** | `motion_worker.py`, `timelapse_worker.py`, `dvr_segment.py`, `motion/detector.rs` (gate only). |
| **APIs** | `vis_gravacao_segmento`, gravação ativas. |
| **Riscos** | MOG2 vs luma — falsos positivos/negativos gravação comercial. |
| **Testes** | §4 auditoria MOG2; 7–14 dias A/B por franqueado; clip duração/max. |
| **Rollback** | Reativar motion/timelapse EasyPanel. |
| **Sucesso** | Paridade S3 + API; produto aceita gate **ou** MOG2 portado. |
| **Desligar** | `confvision-motion`, `confvision-timelapse` |

**Bloqueio:** código Rust **não existe** para clip longo + timelapse — FASE 6 é **implementação**, não config.

---

### FASE 7 — DVR uploader (Go)

| Campo | Conteúdo |
|-------|----------|
| **Objetivo** | Uploader idempotente; MTX inalterado. |
| **Componentes** | Go worker; MTX record API opcional no CP. |
| **Arquivos ref.** | `dvr_main.py`, `dvr_watcher.py`, `dvr_segment.py`, `mediamtx_client.py`. |
| **APIs** | `vis_camera_query_gravacao_ativas`, `vis_gravacao_segmento`, MTX v3 paths. |
| **Riscos** | Segmentos duplicados; lag upload. |
| **Testes** | Shadow mode Python+Go; compare POST. |
| **Rollback** | dvr Python. |
| **Sucesso** | Lag S3 dentro SLA; duplicate handling OK. |
| **Desligar** | `confvision-dvr` Python |

---

## 4. MOG2 × Motion Gate — critérios de teste antes de desligar Python motion

Se a decisão de produto for **substituir MOG2 pelo gate luma** na gravação:

1. **Corpus:** ≥20 câmeras (interno, externo, noturno IR, chuva, estacionamento).
2. **Métricas:** taxa clip/h, duração média, uploads vazios, `MOTION_CLIP_MAX_SEC` hits.
3. **Comparativo:** mesmo RTSP gravado; MOG2 vs luma offline + side-by-side 48h.
4. **Casos obrigatórios:** sombra longa, auto-limpeza, céu/nuvens, headlights, câmera parada 24h (sem clip).
5. **Aceite:** FP/FN dentro de tolerância comercial documentada; rollback motion app testado.

Se **não** passar: portar MOG2 (opencv rust) **ou** manter motion Python com **1 RTSP** (fan-out off).

---

## 5. sync_agent / config_cache — registro (sem correção de código)

| Pergunta | Resposta |
|----------|----------|
| Onde deveriam existir? | Pacote deploy `confvision/` (`sync_agent_main.py`, `sync_agent.py`, `config_cache.py`, `redis_client.py`) — como em `ConfVision/`, `core4-rust-pilot/`. |
| Código antigo/removido? | **Removido ou nunca mergeado** em `core4/confvision/`; referências permanecem. |
| Outro diretório? | `c:\sistemaconfmonit\core4-rust-pilot\` (raiz), `ConfVision`, `ConfVision-github`, worktrees fase0. |
| Necessário runtime? | **Só se** `CONFIG_CACHE_BACKEND=redis` (exemplos EasyPanel) **e** processo sync-agent rodando. Default código: `memory` → `get_cameras_*_direct()`. |
| Containers dependentes? | Workers com redis cache; **rtmp-guard** importa `config_cache` sempre que auth path roda — **falta arquivo = ImportError**. |
| Imports quebrados? | `xano_client` → `sync_agent` **lazy** só com redis cache. `rtmp_guard.py` → `config_cache` **direto**. Dockerfiles mediamtx COPY arquivos inexistentes. |
| Referências mortas? | `main.py` / `distributed_main.py` mensagens sync-agent com worker off. |
| Funciona sem eles? | **Sim** com `CONFIG_CACHE_BACKEND=memory` e sync-agent parado; **não** build mediamtx guard sem copiar arquivos. |
| Risco remover refs? | **Alto** para deploys redis+sync; **médio** para guard Docker; **baixo** para piloto Rust-only + memory cache. |

---

## 6. Fluxos confirmados (referência rápida)

### DVR

```text
Câmera → RTMP → MediaMTX → record fMP4 → DVR_RECORD_DIR/%path%/
  → dvr_watcher (estável) → dvr_segment.process_segment_file
  → gravacao_storage (S3) → POST vis_gravacao_segmento
```

DVR Python **não** abre RTSP.

### Sensor

```text
Receptor → POST vis_evento (API) → status pendente capture
  → sensor_main poll → get_camera_by_id → event_capture (RTSP ffmpeg/OpenCV)
  → upload → finalizar_evento
```

Estado: **Postgres via API**; retry = próximo poll (sem lease no evento).

### RTSP duplicado (mesma câmera, pior caso)

```text
MTX RTSP ─┬─ rust-processor (decode contínuo)
          ├─ motion OpenCV + ffmpeg clip (2 conexões no motion)
          ├─ timelapse OpenCV
          └─ sensor/capture evento (pontual, +1)
```

Impacto: **ALTO** se analítico + motion + timelapse coincidem.

---

## 7. Critérios globais de rollback

1. Postgres: `worker_id` / modos gravação anteriores.
2. EasyPanel: Start serviço Python pausado.
3. Rust: drain ou `MAX_CAMERAS=0`.
4. Validar RTMP/RTSP antes de reabrir analítico.

---

## Histórico

| Data | Nota |
|------|------|
| 2026-09-29 | Plano pós-validação pré-migração; sem implementação |
