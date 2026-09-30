# Fase 2 — Unificação do pipeline de vídeo

**Projeto:** ConfVision  
**Componente principal:** `confvision-rust-processor`  
**Serviços legados (ainda ativos):** `confvision-motion`, `confvision-timelapse`, `confvision-sensor`  
**Pré-requisito:** [FASE-1-RUST-CORE.md](./FASE-1-RUST-CORE.md)

---

## 1. Estado anterior

```text
MediaMTX (cam/{hash})
    ├── Rust Processor     → RTSP retina + decode + luma gate + YOLO + eventos
    ├── confvision-motion  → RTSP OpenCV + MOG2 + ffmpeg clip movimento
    ├── confvision-timelapse → RTSP OpenCV + MOG2 + frames + ffmpeg timelapse
    └── confvision-sensor  → poll API (sem RTSP contínuo) + RTSP pontual capture
```

**Problema:** até **3 consumidores RTSP** por câmera (Rust + motion + timelapse), decode duplicado, CPU e FDs extras.

**Meta Fase 2:** **1 decoder / 1 pipeline RTSP por câmera** no Rust, com módulos:

```text
Camera Pipeline
 ├── Motion Module      (2A)
 ├── YOLO Module        (existente)
 ├── Timelapse Module   (2B)
 ├── Sensor Module      (2C)
 ├── Snapshot Module    (existente capture)
 └── Clip Module        (ffmpeg coordenado)
```

---

## 2. Motion — auditoria `confvision-motion` (código real)

| Item | Implementação (`core4/confvision/`) |
|------|-------------------------------------|
| Entry | `motion_main.py` → sync `get_cameras_gravacao_ativas` + `filter_gravacao_cameras(motion=True)` |
| Worker/câmera | `MotionWorkerManager` / `MotionCameraWorker` (1 thread RTSP/câmera) |
| RTSP | `cv2.VideoCapture(url, CAP_FFMPEG)`, `rtsp_transport;tcp`, `BUFFERSIZE=1` |
| URL | `urls.rtsp_url_for_camera` → MediaMTX `MEDIAMTX_RTSP_BASE` |
| Decode | OpenCV (BGR frames) |
| Motion | **MOG2** `createBackgroundSubtractorMOG2(history, varThreshold, detectShadows=True)` |
| Pós-process | resize → `MOTION_ANALYSIS_WIDTH` (640), threshold 200, MORPH_OPEN, contornos |
| Área mínima | `MOTION_MIN_AREA` (default 1500) |
| FPS análise | 1 a cada `MOTION_FRAME_SKIP` (default 3) |
| REC START | MOG2 motion + não gravando → `ffmpeg` pull RTSP **segundo stream** libx264 clip |
| REC STOP | sem motion `MOTION_POST_ROLL_SEC` (5s) ou `MOTION_CLIP_MAX_SEC` (300s) |
| Upload | `dvr_segment.process_segment_file(..., tipo="movimento")` → API + S3 |
| Redis | **Não** |
| Reconnect | loop + `MOTION_RECONNECT_SEC` (5) |
| Env chave | `MOTION_*`, `MEDIAMTX_RTSP_BASE`, `XANO_BASE_URL` / Go equivalente |

**Nota:** `motion_detect.py` (MOG2 `detectShadows=False`) é gate **analítico Python worker**, não o serviço motion de gravação.

---

## 3. Luma (Rust) × MOG2 (Python)

| Aspecto | Rust (`motion/detector.rs`) | Python motion |
|---------|------------------------------|---------------|
| Entrada | Luma downscaled decode pipeline | BGR MOG2 |
| Resolução | `DECODED_LUMA_*` (decode policy) | max width 640 proporcional |
| Threshold | % pixels diff + scene shift 35% | varThreshold 25 + área contorno |
| Objetivo hoje | **Gate YOLO** (`ANALYSIS_ONLY_ON_MOTION`) | **Gravar clip** + upload segmento |
| Equivalente? | **Não** automaticamente | — |

**Decisão Fase 2A (provisória):** não portar MOG2 até shadow mode + testes de cena. MOG2 só se luma **não** atingir paridade funcional para **gravação por movimento**.

---

## 4. Shadow mode (2A — EM IMPLEMENTAÇÃO)

```text
Mesma câmera no MediaMTX
    ├── Rust (luma)     → MOTION_SHADOW_COMPARE=1 → rust_cam_{id}.jsonl
    └── Python (MOG2)   → MOTION_SHADOW_COMPARE=1 → python_cam_{id}.jsonl
              └── scripts/motion_shadow_compare.py
```

| Variável | Rust | Python |
|----------|------|--------|
| `MOTION_SHADOW_COMPARE` | `1` | `1` |
| `MOTION_SHADOW_LOG_DIR` | `/tmp/confvision/motion-shadow` | idem |

Rust também emite eventos de sessão (`motion_started` / `motion_ended`) alinhados a `MOTION_POST_ROLL_SEC`.

**Serviço Python:** permanece **ATIVO**; shadow não gera segundo clip por si só.

---

## 5. Motion events (Rust)

Módulo `motion/session.rs`:

- `MotionStarted` — primeira detecção após idle  
- `MotionEnded` — sem detecção por `MOTION_POST_ROLL_SEC`  
- `MotionUpdated` — reservado (clip progress futuro)

Contrato API gravação (`vis_gravacao_segmento`) **inalterado** até clip Rust implementado em `motion/recording.rs`.

---

## 6. Motion clip (pendente 2A)

Python: segundo RTSP via ffmpeg durante REC.  
Alvo Rust: **coordenar** ffmpeg no fim/início de sessão usando **mesmo** RTSP URL (sem novo decode OpenCV) — buffer pre/post: auditar `CLIP_DURACAO_SEG` (eventos) vs `MOTION_POST_ROLL_SEC` / `MOTION_CLIP_MAX_SEC` (gravação).

| Parâmetro | Default Python |
|-----------|----------------|
| Pre-buffer | implícito no ffmpeg pull contínuo |
| Post-buffer | `MOTION_POST_ROLL_SEC=5` |
| Max clip | `MOTION_CLIP_MAX_SEC=300` |

---

## 7. Timelapse — auditoria (2B — AUDITADO)

| Item | `timelapse_worker.py` |
|------|------------------------|
| Sync | `modo_gravacao=timelapse` via mesma API gravação |
| RTSP | OpenCV dedicado (duplicado) |
| MOG2 | Sim — alterna estado TIMELAPSE ↔ MOVIMENTO |
| Timelapse | JPEG a cada `TIMELAPSE_FRAME_INTERVALO_SEG` (720s), `TIMELAPSE_FRAMES_POR_SEGMENTO` (60) → ffmpeg concat 1fps |
| Movimento | Mesmo padrão ffmpeg que motion worker |
| Upload | `tipo="timelapse"` |
| Flush | `ack_flush_pedido` API |

Rust: stub `src/timelapse/mod.rs` — scheduler **não** wired ao pipeline.

---

## 8. Sensor — auditoria (2C — AUDITADO)

| Etapa | Código |
|-------|--------|
| Poll | `sensor_main.py` → `get_eventos_sensor_pendentes` (3s) |
| Evento | Já criado no receptor/API |
| Capture | `event_capture.processar_evento_sensor` → `processar_deteccao(..., for_sensor=True)` |
| RTSP | **Novo** pull ffmpeg/OpenCV por evento (`capture.py`) |
| Finaliza | upload S3 + `put_evento` status pronto |

Rust: `capture/process.rs` trata sensor no fluxo **job Redis/YOLO**, **sem** poll pendentes. Stub `src/sensor/mod.rs`.

---

## 9. Pipeline unificado (alvo)

```text
MediaMTX
    │
    ▼
Rust Processor (1× RTSP + decode)
    │
    ▼
Frame Pipeline (bounded, drop-oldest)
    ├── Motion (luma + sessão + clip ffmpeg)
    ├── YOLO
    ├── Timelapse scheduler (frames esparsos)
    └── Sensor hook (frame on demand + poll Go)
    ▼
Events / Storage (contratos atuais)
```

---

## 10. Resource management

| Hoje | Meta Fase 2 |
|------|-------------|
| 1–3 RTSP/câmera | 1 RTSP/câmera no nó Rust owner |
| OpenCV threads | eliminar onde Rust owner |
| ffmpeg | subprocess pontual (clip/timelapse), não pull paralelo OpenCV |

---

## 11. Backpressure

Reutilizar Fase 1: `FRAME_BUFFER_MAX`, drop-oldest, admission/load shedding.  
Timelapse/sensor: prioridade **abaixo** de RTSP/eventos/YOLO (implementar na 2B/2C).

---

## 12. Métricas (roadmap)

| Grupo | Status |
|-------|--------|
| Pipeline frames | ✅ Fase 1 `/metrics` |
| Motion luma counts | ✅ `motion_detected`, etc. |
| `motion_started` / `motion_ended` | ⚠️ logs + shadow JSONL; Prometheus nativo pendente |
| Timelapse / Sensor | ❌ pós-integração |

---

## 13–16. Testes

| Tipo | Status |
|------|--------|
| Unit Rust motion session | ✅ `motion/session.rs` |
| Shadow compare script | ✅ `scripts/motion_shadow_compare.py` |
| Integração MediaMTX | ⚠️ manual |
| Falha / stress multi-câmera | ⚠️ pendente |
| Luma vs MOG2 cenas | ⚠️ pendente coleta shadow |

---

## 17. Rollback

1. Desligar `MOTION_SHADOW_COMPARE` nos processors.  
2. Manter `confvision-motion` / timelapse / sensor **EasyPanel ativos**.  
3. Rust: flags de gravação futuras default **off** até VALIDADO.  
4. Não remover código Python até período de estabilidade.

---

## 18. Serviços antigos desativados

| Serviço | Status |
|---------|--------|
| confvision-motion | **ATIVO** (produção) |
| confvision-timelapse | **ATIVO** |
| confvision-sensor | **ATIVO** |

---

## 19. Matriz de migração

| Serviço | Função | Antes | Depois | Status |
|---------|--------|-------|--------|--------|
| confvision-motion | Motion + clip gravação | Python MOG2 + RTSP | Rust pipeline + ffmpeg | **SHADOW** (compare JSONL) |
| confvision-timelapse | Timelapse + movimento | Serviço separado RTSP | Rust scheduler + ffmpeg | **AUDITADO** |
| confvision-sensor | Capture pós-evento | Poll + RTSP evento | Rust frame + poll Go | **AUDITADO** |

---

## 20. Problemas conhecidos

1. Rust luma ≠ MOG2 — decisão pendente testes.  
2. Gravação movimento **não existe** no Rust (só gate analítico).  
3. Timelapse e sensor ainda abrem RTSP próprio.  
4. Sensor poll no Go/Xano — integração Rust precisa contrato estável.

---

## 21. Próxima fase

**FASE 3 — Processamento e eventos** (fora deste documento): event bus, assignment, consolidação eventos — **somente após** 2A validada (motion clip + shadow).

---

## 22. Alterações desta entrega (código)

| Área | Arquivo | Mudança |
|------|---------|---------|
| Shadow Rust | `src/motion/shadow.rs`, `hooks.rs`, `session.rs` | JSONL + sessão |
| Pipeline | `pipeline/frame_pipeline.rs`, `worker/camera_worker.rs` | hooks pós-motion |
| Config | `MOTION_SHADOW_*`, `MOTION_POST_ROLL_SEC`, `MOTION_CLIP_MAX_SEC` | env |
| Stubs | `timelapse/mod.rs`, `sensor/mod.rs`, `motion/recording.rs` | 2B/2C |
| Python | `core4/confvision/motion_worker.py` | shadow JSONL |
| Tool | `scripts/motion_shadow_compare.py` | comparação |

---

## Referências

- [ARCHITECTURE_DECISIONS.md](./ARCHITECTURE_DECISIONS.md) (ADR-007)
- [AUDITORIA_MIGRACAO_SERVICOS_RUST.md](../../../core4/confvision/docs/AUDITORIA_MIGRACAO_SERVICOS_RUST.md)
