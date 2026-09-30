# ConfVision — Grafo de dependências (código auditado)

Atualizado: 2026-09-30. Adaptado ao deploy típico **FoxPro / EasyPanel** (Control Plane + Node com MediaMTX + workers).

```text
                    ┌─────────────────────────┐
                    │   Clientes / Portal     │
                    └───────────┬─────────────┘
                                │ HTTPS
                                ▼
                    ┌─────────────────────────┐
                    │  confvision Go (visapi) │  ← Control Plane API
                    │  + visdata (Postgres)   │
                    └───────────┬─────────────┘
                                │
              ┌─────────────────┼─────────────────┐
              │                 │                 │
              ▼                 ▼                 ▼
       ┌────────────┐    ┌────────────┐   ┌──────────────┐
       │ PostgreSQL │    │   Redis    │   │ Contabo S3   │
       │  metadata  │    │ fila/cache │   │    media     │
       └────────────┘    └─────┬──────┘   └──────▲───────┘
                               │                  │
     ┌─────────────────────────┼──────────────────┼────────────────────┐
     │                         │                  │                    │
     ▼                         ▼                  │                    ▼
┌─────────────┐        ┌──────────────┐          │            ┌───────────────┐
│ sync-agent  │───────►│ config_cache │          │            │ DVR worker    │
│ (Python)    │  write │ (Redis keys) │          │            │ dvr_main.py   │
└──────┬──────┘        └──────────────┘          │            └───────┬───────┘
       │ GET API                                  │                    │
       ▼                                          │                    │
┌──────────────────────────────────────────────────────────────────────────────┐
│                         NODE (mesmo host ou VPS dedicado)                     │
│  ┌─────────────┐     RTSP      ┌──────────────┐                              │
│  │  MediaMTX   │◄──────────────│ Câmeras RTMP │                              │
│  │ record path │               └──────────────┘                              │
│  └──────┬──────┘                                                             │
│         │ rtsp://…/cam/{hash}                                                │
│         ▼                                                                    │
│  ┌──────────────────────┐   LPUSH/BRPOP   ┌─────────┐   PUT objects         │
│  │ confvision-rust-     │◄───────────────►│ Redis   │───────────────────────┘
│  │ processor            │                 └─────────┘                        │
│  │  RTSP ingest         │   POST /vis_evento, finalizar                       │
│  │  YOLO / motion       │──────────────────────────────► Go API ──► Postgres  │
│  │  capture workers     │   CAPTURE_DIR (temp)                                │
│  └──────────────────────┘                                                    │
│  ┌──────────────────────┐   POST segmento                                    │
│  │ motion / timelapse / │────────────────────────────────► Go + S3           │
│  │ sensor (Python)      │   MOTION_RECORD_DIR local                          │
│  └──────────────────────┘                                                    │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Dependências por componente

| Componente | Depende de | Se cair… (resumo) |
|------------|------------|-------------------|
| **Go visapi** | PostgreSQL | API indisponível; workers não persistem eventos novos |
| **Rust processor** | MediaMTX RTSP, Go API (sync/ping/evento), Redis (se fila redis) | RTSP continua tentando; sem PG eventos não persistem; sem Redis fila capture para (RTSP/YOLO podem continuar) |
| **Capture Rust** | Redis ou memory queue, Go API, S3 opcional | S3 falha → evento ainda `finalizar_evento` sem URL |
| **DVR Python** | MediaMTX API, Go API, S3 por franqueado, disco `DVR_RECORD_DIR` | Grava local MTX pode continuar; upload/metadata falham isolados |
| **sync-agent** | Go API, Redis opcional | Workers Python podem usar API direta ou cache stale |
| **Portal** | Go API | Leitura histórica indisponível |

## Fluxos críticos

1. **Evento analítico:** Câmera → Rust RTSP → detecção → Redis `confvision:eventos` → capture → Postgres `vis_evento` + S3.
2. **DVR contínuo:** Câmera → RTMP/MediaMTX record → `DVR_RECORD_DIR` → watcher → S3 → `vis_gravacao_segmento`.
3. **Desired state câmera:** Postgres `vis_camera.worker_id` → Rust pull `GET /vis_camera_sync_ativas` → reconciliação local.
