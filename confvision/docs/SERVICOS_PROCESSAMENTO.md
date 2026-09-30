# Serviços de processamento ConfVision (pós-consolidação)

| Serviço EasyPanel | Responsabilidade atual | Quem executa | Rust substitui? |
|-------------------|------------------------|--------------|-----------------|
| **confvision-rust-processor** | Analítico principal: RTSP, decode, motion gate, YOLO, eventos, Redis, snapshots/clips | Rust (`core4/confvision-rust-processor`) | **N/A (principal)** |
| **confvision-worker** | Analítico Python legado (`main.py`) | Python | **Sim** (desativado — não reativar em paralelo) |
| **confvision-motion** | Clips MOG2 / motion worker legado | Python `motion_main.py` | **Parcial** (Rust faz gate analítico; motion Python mantido) |
| **confvision-sensor** | Terminal / sensor | Python `sensor_main.py` | **Não** (legado ativo) |
| **confvision-sync-agent** | Cache config Redis para workers Python | Python `sync_agent_main.py` | **Não** (legado ativo) |
| **confvision-timelapse** | Timelapse | Python `timelapse_main.py` | **Não** (legado ativo) |
| **confvision-dvr** | Segmentos gravação + API Go | Python `dvr_main.py` + MediaMTX | **Não** (legado ativo) |
| **confvision** (MediaMTX) | Ingest RTSP/RTMP/HLS | MediaMTX + Guard | **Não** |

## RTSP — evitar duplicidade analítica

- **Uma câmera** não deve ter **Rust processor + confvision-worker** ativos no mesmo `WORKER_ID`/shard.
- **confvision-worker:** manter **STOPPED**.
- Motion/timelapse/DVR podem usar RTSP/MediaMTX conforme config existente; documentar no runbook se a mesma câmera abre múltiplas sessões **por design** (ex.: DVR grava path vs Rust analítico).

Código Python canônico: `core4/confvision/`. Rust: `core4/confvision-rust-processor/`.
