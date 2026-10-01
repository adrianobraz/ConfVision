# Inventário de protocolos — ConfVision MediaMTX Node V1

| Protocolo | Baseline V1 | Status | Notas |
|-----------|-------------|--------|-------|
| RTMP ingest | `:1935` on | **ATIVO** | Fluxo principal NVR → `cam/{hash}` |
| RTSP | `:8554` TCP | **ATIVO** | Rust / rede interna |
| HLS | `:8888` | **ATIVO** | `hlsAlwaysRemux: false` — mux on-demand |
| Playback | `:9996` | **ATIVO** | Segmentos `record` |
| HTTP API | `:9997` | **ATIVO** | DVR `v3/config/paths/*` |
| Prometheus metrics | `:9998` | **ATIVO** | Scrape via credencial Guard |
| WebRTC | — | **NÃO UTILIZADO** | Overlay `production-webrtc.yaml` |
| SRT | — | **NÃO UTILIZADO** | Overlay `production-srt.yaml` |
| MoQ (Media over QUIC) | — | **FUTURO** | `moq: false` baseline |
| RTMPS / RTSPS / TLS | — | **FUTURO** | Terminação TLS no LB/reverse proxy |

Auth: sempre **HTTP externo** → RTMP-GUARD (`authMethod: http`).
