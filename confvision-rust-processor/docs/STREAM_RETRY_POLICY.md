# Política automática de retry RTSP (Rust + API Go)

## Objetivo

Evitar CPU infinita em câmeras **sem publisher** (RTSP 404) em fleet **100% Rust**, sem SQL manual.

## Intervalos (falhas consecutivas — path ausente / 404)

| Falhas | Espera até próxima tentativa |
|--------|------------------------------|
| 1–5 | 1 min |
| 6–9 | 2 min |
| 10–19 | 10 min |
| 20–29 | 20 min |
| 30–59 | 30 min |
| ≥ 60 | 60 min (+ contador horário) |

Após **6 tentativas** no modo ≥60 falhas → `analitico_pausado=true`, `stream_motivo_pausa=sistema_stream_*`.

## Reset pelo usuário

`ativo: false` → `ativo: true` incrementa `stream_policy_generation` no Postgres; o Rust zera contadores no sync.

## Nunca conectou

Sem `ultimo_stream_ok_em` e fora do grace (`STREAM_NEVER_OK_GRACE_HOURS`, default 72) → pausa analítico automática.

## Env (rust-pilot)

```env
STREAM_RETRY_ENABLED=1
SYNC_FILTER_WORKER_ID=true
STREAM_NEVER_OK_GRACE_HOURS=72
STREAM_HOURLY_FAIL_THRESHOLD=60
STREAM_HOURLY_MAX_ATTEMPTS=6
```

## API

- Ping: `POST /vis_worker_ping` com `camera_stream_health[]` (`stream_ok`, `stream_failure`, `stream_incident`, `pause_analytic`; campos opcionais `last_error`, `error_class`).
- Reativação (webhook/ops): `POST /vis_camera_stream_reactivate?camera_id=`.

Ver também: [TROUBLESHOOTING_RTSP_ERRORS.md](./TROUBLESHOOTING_RTSP_ERRORS.md), [PAINEL_ATIVO_VS_STREAM_POLICY.md](./PAINEL_ATIVO_VS_STREAM_POLICY.md).

## Migration Postgres

- `sql/migrations/20260326_vis_camera_stream_policy.sql` — contadores e generation.
- `sql/migrations/20260326_vis_camera_stream_error_diag.sql` — `stream_ultimo_erro`, `stream_erro_classe`, `stream_ultimo_erro_em`.
