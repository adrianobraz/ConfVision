# Rollback — MediaMTX Node V1

## Imagem

- **Anterior:** `bluenviron/mediamtx:1` + `confvision-python/mediamtx/mediamtx.yml` (`Dockerfile.mediamtx`)
- **Atual:** `confvision-mediamtx-node:1.0.0` (`Dockerfile.v1`, MTX **1.21.0**)

Rollback = redeploy imagem/tag anterior + config legada.

## Config

- `CONFIG_VERSION` legado: implícito (sem version file)
- V1: `confvision-mtx-v1.0.0` em `VERSION.json`

## Riscos conhecidos V1 → legado

| Mudança V1 | Efeito rollback |
|------------|-----------------|
| `hlsAlwaysRemux: false` | Legado `yes` — mais CPU se voltar |
| `metrics: yes` :9998 | Legado sem metrics nativo |
| MTX 1.21.0 | Pin `:1` pode ser versão menor |

## Procedimento

1. Pausar rolling update na fatia afetada.
2. Redeploy `Dockerfile.mediamtx` ou tag anterior.
3. Validar 1 câmera RTMP + RTSP + DVR.
4. Retomar rollout só após causa raiz documentada.

Secrets/env Guard **inalterados** — rollback só streaming plane.
