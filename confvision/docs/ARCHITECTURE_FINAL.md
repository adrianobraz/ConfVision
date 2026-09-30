# ConfVision — Arquitetura final (consolidada, estado real)

Data: 2026-09-30. Reflete **implementação atual**, não o spec completo das 6 fases.

---

## Diagrama

```text
                         CLIENTES / PORTAL
                                │
                                ▼
              ┌─────────────────────────────────────┐
              │  CONTROL PLANE                       │
              │  Go confvision (visapi + visdata)      │
              │  PostgreSQL (metadata, config)       │
              └──────────────┬──────────────────────┘
                             │
       ┌─────────────────────┼─────────────────────┐
       │                     │                     │
       ▼                     ▼                     ▼
  D5 Scheduler          vis_worker           vis_mediamtx_node
  (capacity-report)     (heartbeat)          (ingest node)
       │                     │                     │
       └──────────┬──────────┴──────────┬──────────┘
                  │                     │
    ┌─────────────▼─────────────┐       │
    │  DATA PLANE (foxpro/VPS)  │       │
    │                           │       │
    │  confvision-rust-processor│◄──────┘  WORKER_ID + sync pull
    │    RTSP → decode → motion gate     │
    │    → YOLO → Redis → capture        │
    │    → API → vis_evento + S3         │
    │                           │       │
    │  MediaMTX (RTSP/RTMP/record)        │
    │                           │       │
    │  Python (paralelo, legado):         │
    │    motion_main / timelapse_main     │
    │    sensor_main / dvr_main           │
    │    sync_agent_main (+ Redis cache)  │
    └─────────────┬─────────────┘       │
                  │                     │
                  ▼                     ▼
            Redis (filas/cache)   S3 Contabo (mídia)
```

**Node Agent (binário):** **não existe** — funções em Rust (`sync`, `ping`, `CameraManager`).

---

## Quem chama quem

| De | Para | Protocolo | Dados |
|----|------|-----------|-------|
| Rust | Go API | HTTPS | sync câmeras, create/finalizar evento, ping worker |
| Rust | MediaMTX | RTSP | frames |
| Rust | Redis | LIST | `confvision:eventos` (+ DLQ) |
| Rust | S3 | HTTP PUT | snapshots/clips |
| Capture (Rust) | Go | HTTPS | `vis_evento` |
| DVR Python | MediaMTX | API + filesystem | segmentos |
| DVR Python | Go | POST | `vis_gravacao_segmento` |
| sync-agent | Go + Redis | HTTP + cache | config workers Python |
| Portal | Go | HTTPS | leitura/CRUD |
| D5 Go | Rust | GET `/capacity-report` | assign `worker_id` |

---

## Quem grava onde

| Dado | Store | Dono metadata |
|------|-------|---------------|
| Config câmera | PostgreSQL `vis_camera` | Control Plane |
| Assignment | `vis_camera.worker_id` | Control Plane (D5/manual) |
| Runtime heartbeat | `vis_worker` | Data plane escreve via API |
| Eventos | `vis_evento` (+ clips) | Go via API Rust |
| Mídia evento | S3 + URLs em PG | Rust capture |
| DVR | S3 + `vis_gravacao_segmento` | Python dvr |
| Fila analítica | Redis | Ephemeral |
| Cache config | Redis `confvision:sync:*` | sync-agent |

---

## Quem processa vídeo

| Função | Processador |
|--------|-------------|
| Analítico humano / YOLO | **Rust** (piloto) |
| Motion clip MOG2 (legado) | **Python motion** |
| Timelapse | **Python timelapse** |
| Sensor | **Python sensor** |
| Gravação contínua | **MediaMTX + Python DVR** |
| Snapshot/clip evento | **Rust** (FFmpeg pontual) |

---

## Estado e consistência

| Estado | Forte / fraco | Onde |
|--------|-----------------|------|
| Desired assignment | Forte (PG) | `worker_id` |
| Actual running cameras | Fraco (memória Rust) | Reconcile após restart via sync |
| Lease anti-duplicata | **Ausente** | — |
| Stream health | PG + Rust policy | `stream_*` columns |

---

## Repositórios de código (consolidação)

| Caminho | Papel | Status |
|---------|-------|--------|
| `core4/home/confmonit/v4.0/confvision/` | Aplicação Go (control plane) | **IMPLEMENTADO** |
| `core4/confvision/` | Processamento Python + EasyPanel env | **IMPLEMENTADO / LEGADO** |
| `core4/confvision-rust-processor/` | Processador analítico Rust (Fases 1–6) | **IMPLEMENTADO** |
| `core4/deploy/tenant-stack/` | Stack tenant/host (env examples) | **IMPLEMENTADO** |
| `core4-rust-pilot` | Worktree histórico `rust-pilot` | **PRESERVADO** (fonte já copiada) |

Branch de integração Git: `consolidate/confvision-rust` (a partir de `main` local).

### Legenda

| Tag | Significado |
|-----|-------------|
| **IMPLEMENTADO** | Código no `core4` consolidado |
| **LEGADO** | Python motion/dvr/timelapse/sensor/sync — mantido até substituição |
| **EM MIGRAÇÃO** | Responsabilidades movendo para Rust sem desligar legado |
| **FUTURO** | Lease anti-duplicata, Node Agent binário, desligar motion Python |

Publicação: push de `consolidate/confvision-rust` → PR para `main` / branch de deploy acordada.
