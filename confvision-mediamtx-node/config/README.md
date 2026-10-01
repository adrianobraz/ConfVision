# Configuração — seções (sem lógica de negócio)

| Seção | Arquivo | Conteúdo |
|-------|---------|----------|
| GLOBAL | `mediamtx.confvision.yml` | timeouts, filas UDP |
| LOGGING | idem | stdout + `/recordings/mediamtx.log` |
| AUTH | idem | HTTP → RTMP-GUARD |
| API | idem | :9997 |
| METRICS | idem | :9998 Prometheus |
| RTSP / RTMP | idem | ingest + fan-out interno |
| HLS | idem | on-demand |
| WEBRTC / SRT / MoQ | baseline off, `overlays/*` |
| RECORD | `pathDefaults` | off; DVR patch por path |
| PATH DEFAULTS | idem | publisher, fmp4 path |
| PATHS | idem | `~^cam/...` + `all_others` |

Overlays (`overlays/`) ajustam ambiente; para produção preferir editar um único YAML renderizado (`scripts/render-config.ps1`) e validar (`scripts/validate-config.ps1`).
