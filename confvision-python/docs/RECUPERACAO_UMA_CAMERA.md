# Recuperar uma câmera — fluxo básico (agora)

Ordem de verificação (pare no **primeiro** `[FALHA]`):

```text
NVR RTMP → Guard auth → MediaMTX publish → RTSP path → Rust sync → decode
```

## Diagnóstico local (API Go)

```powershell
cd C:\sistemaconfmonit\core4\confvision-python
$env:VIS_WORKER_API_KEY = "<chave EasyPanel>"
$env:RUST_HEALTH_URL = "https://<seu-rust-pilot>/health"
.\scripts\diagnostico-fluxo-camera.ps1 -CameraId 6 -WorkerId "rust-processor-pilot-b-04"
```

## Causas frequentes (evidência foxpro 2026-10)

| Camada | Sintoma | Ação mínima |
|--------|---------|-------------|
| Guard | `ip_banido` | Unban no Guard / `rtmp_bans.json`; `RTMP_BAN_MAX_FAILS=12` |
| Guard | `camera_inativa` | `ativo=true` ou plano online; path `cam/{hash}` ≠ número UI |
| Guard | `stream_pausado_sistema` | Corrigir encode; Reativar stream no painel |
| MediaMTX | HLS “no stream” | NVR não publicou ou auth 403 |
| Cadastro | `rtsp_url_sec` → `srv1.dnsid.com.br` | **PUT** RTSP interno `rtsp://foxpro_confvision:8554/cam/{hash}` |
| Rust | HTTP 503 `/health` | **Start** app `foxpro-rust-pilot*` no EasyPanel |
| Assignment | sync vazio | `worker_id` no Postgres = `WORKER_ID` do container Rust |

## Fluxo RTMP → MTX → RTSP (câmera 6 exemplo)

- Hash: `m6e4ywjpda5o` → path `cam/m6e4ywjpda5o`
- RTSP interno: `rtsp://foxpro_confvision:8554/cam/m6e4ywjpda5o`
- Worker cadastro: `rust-processor-pilot-b-04`

## Corrigir RTSP para rede Docker (API)

```powershell
$body = '{"rtsp_url_sec":"rtsp://foxpro_confvision:8554/cam/m6e4ywjpda5o","analitico_pausado":false}'
Invoke-RestMethod -Method Put -Uri "https://vision.confmonit2.com.br/vis_camera/6" `
  -Headers @{ "X-Vis-Worker-Key" = $env:VIS_WORKER_API_KEY; "Content-Type" = "application/json" } `
  -Body $body
```

## EasyPanel — ordem Start

1. `confvision` (MediaMTX + Guard)
2. `foxpro-rust-yolo-sidecar` (se YOLO HTTP)
3. `foxpro-rust-pilot` / `-b` / `-b-04` com `WORKER_ID` = Postgres

Não depende de `confvision-worker` se o **NVR** publica RTMP direto no MediaMTX.

Ver também: [confvision-rust-processor/docs/RECUPERACAO_CAMERAS.md](../../confvision-rust-processor/docs/RECUPERACAO_CAMERAS.md).
