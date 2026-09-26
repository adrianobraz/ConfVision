# Runbook — câmeras ativas com RTSP 404 (MediaMTX)

## Sintoma

- No Rust (`/metrics` ou `/capacity-report`): `status=offline` ou `error`, `last_error` com **404**, **DESCRIBE failed** ou **Not Found**.
- RTSP típico: `rtsp://…:8554/cam/{hash}` — path inexistente no MediaMTX (DVR não publicou ou path errado).

## Causas comuns

1. **Câmera ativa no ConfVision** (`ativo=true`) mas **sem publish RTMP** (wifi/DVR desligado, hash errado).
2. **Barra final no path** (`cam/{hash}/`) — UI/DVR com `dvr=true` pode enviar trailing slash; o `rtmp_guard` normaliza com `rstrip("/")` (deploy confvision atualizado).
3. **Câmera no worker Rust** que deveria estar no **Python** (piloto ou carga).

## Diagnóstico rápido

```bash
BASE=https://foxpro-rust-pilot.rkr351.easypanel.host
curl -fsS "$BASE/capacity-report" | jq '{summary, recommended_actions, cameras: [.cameras[] | select(.issue=="rtsp_404")]}'
```

Ou script: `confvision-rust-processor/scripts/capacity-report.sh "$BASE"`.

## Comportamento automático (Rust + API)

Com `STREAM_RETRY_ENABLED=1`, câmeras 404 entram em backoff (1–5 min → … → 1 h) e podem pausar analítico sozinhas (`analitico_pausado`). Ver **`docs/STREAM_RETRY_POLICY.md`**.

Reativar após corrigir DVR: `POST /vis_camera_stream_reactivate?camera_id=` ou usuário desliga/liga `ativo` (zera contadores).

## Ações manuais (emergência)

| Objetivo | Ação |
|----------|------|
| Corrigir stream | Garantir publish RTMP no path `cam/{hash}` (sem barra extra após deploy do guard) |
| Forçar pausa | `analitico_pausado=true` via painel/API |

SQL pronto (revise IDs e `PYTHON_WORKER_ID`):  
`confvision-rust-processor/sql/cameras_404_worker_python.sql`

## Piloto conhecido (referência)

Em testes com `MAX_CAMERAS=10`, ids **2, 4, 8, 9** apareceram com RTSP 404 enquanto **6** câmeras ficaram online — mover 404 para Python reduz carga sem perder monitoramento no worker legado.

## Após mudança no Postgres

- Aguardar `SYNC_INTERVAL_SEC` (default 60s) ou reiniciar o serviço rust-pilot.
- Validar: `./scripts/validate-piloto-fase-a.sh "$BASE"` e `/capacity-report`.

## Escala (2º processor)

Só criar **2º `PROCESSOR_ID`** quando `/capacity-report` indicar **`split_second_processor`** (sem headroom). Primeiro elimine 404/offline desnecessários no Rust.
