# Diagnostico Fase 0 — atualizado 2026-09-29

## Linha do tempo (logs operador)

| Fase | O que aconteceu |
|------|------------------|
| 23:46–23:53 | Shedding derrubou workers A (20,22) e B (21, **15**) |
| 00:08 / 00:05 | IDs elegiveis para reabrir (shed list) |
| **00:36** | Restart A/B — **4 workers** cada piloto |
| 00:36:55 | Shutdown rapido B (deploy/restart) + **erros h264 NAL** |
| A cam 4 | **RTSP 404** no DESCRIBE (`boezyjmydlk4`) — path pode divergir do publish RTMP |

## Estado apos restart (testes ~00:37 UTC)

| Check | Resultado |
|-------|-----------|
| Rust B `cameras_online` | **4/4**, fps ~21 |
| Rust A `cameras_online` | **3/4** (404 na 4) |
| `events_published` | **0** |
| Cam 15 Rust metrics | **online**, ~9 fps, **37+ frames decoded**, **motion_detected=0** |
| Global B metrics | **decode_errors=45**, motion_detected=0 |
| YOLO sidecar | OK |
| `vis_camera.status` 15 | **offline** (`ultimo_ping_em` null — ping e por **vis_worker** + `stream_ok`, nao RTMP) |

## Causa raiz (camadas)

1. **Resolvido (parcial):** shedding + admission zeravam workers — apos restart, RTSP OK.
2. **Ativo — eventos:** pipeline roda YOLO (frames decoded/enqueued) mas **`pessoas_match=0`** ou regras area/modo — nenhum `evento publicado na fila` nos logs.
3. **Ativo — motion:** `motion_detected=0` em cena estatica; com `ANALYSIS_ONLY_ON_MOTION` YOLO so apos movimento/arm gate (MotionGatedSession).
4. **Ativo — decode:** erros **Invalid NAL unit size** corrompem parte dos frames (cameras/streams instaveis).
5. **Cadastro UI:** `somente_armado` na 15 e plano armado — confirmar **dispositivo armado** no produto (Rust usa gate de **movimento**, nao API armado ainda).
6. **Cam 4 pilot-a:** RTSP 404 — corrigir hash/URL vs MediaMTX.

## Fechar Fase 0 (checklist)

- [x] Rust online na piloto 15 (processor B)
- [ ] `events_published > 0` ou linha em `vis_evento` camera 15
- [ ] Baseline anotado
- [ ] Manter `LOAD_SHEDDING_ENABLED=0` ate aceite
- [ ] Gerar movimento/pessoa na 15 ou teste gerador eventos

## Comandos

```powershell
$env:POSTGRES_URL='postgres://...'
.\fase0-diagnose.ps1 -WorkerKey '...'
.\fase0-verify.ps1 -WorkerKey '...'
cd pg-audit; go run .
```
