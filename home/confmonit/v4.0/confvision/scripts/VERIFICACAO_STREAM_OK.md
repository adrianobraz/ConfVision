# Verificacao: stream_ok e status offline no Postgres

## Fluxo esperado

1. Rust `run_ping_loop` (cada `PING_INTERVAL_SEC`, default 30s) chama `POST /vis_worker_ping` em `vision.confmonit2.com.br`.
2. Rota `visapi.handleDispatch` -> `visdata.UpsertWorkerPing`.
3. Se o body incluir `camera_stream_health[]`, `ApplyCameraStreamHealthBatch` roda:
   - `event=stream_ok` -> `UPDATE vis_camera SET ultimo_stream_ok_em = NOW(), stream_falhas_consecutivas = 0 ...`
   - `event=stream_failure` -> incrementa contadores (nao zera stream ok)

## Bug encontrado (antes do fix)

`stream_ok` so era enfileirado quando a **sessao RTSP terminava com Ok** (`rtsp session ended — reconnecting`), nao durante stream estavel.

Efeito: câmera **online no /metrics** por minutos, mas `ultimo_stream_ok_em` **null** e `stream_falhas_consecutivas` > 0 de tentativas antigas.

## Campo `vis_camera.status = offline`

O ping do worker **nao** atualiza `vis_camera.status` nem `vis_camera.ultimo_ping_em` (esses campos eram do worker Python / scan).

O checklist Fase 0 usava `status offline` como proxy errado. Corrigido para validar **`vis_worker`** (`cameras_ativas=4`) e, apos deploy, **`ultimo_stream_ok_em`**.

## Fix Rust (branch local)

`CameraManager::queue_stream_ok_for_online_cameras()` antes de cada drain no ping — uma entrada `stream_ok` por câmera com `CameraStatus::Online`.

**Deploy necessario** nos containers pilot A/B para refletir em producao.

## Como validar apos deploy

```powershell
$env:POSTGRES_URL='...'
cd scripts/pg-audit; go run .
# Secao "Stream OK piloto-b" deve mostrar ultimo_stream_ok_em recente na 15
```

```powershell
.\fase0-verify.ps1 -WorkerKey '...'
```
