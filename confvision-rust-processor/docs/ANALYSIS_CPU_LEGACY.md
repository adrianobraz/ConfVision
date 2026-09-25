# Consumo de CPU — análise (decode + motion)

O piloto Rust decodifica H.264 na CPU e roda detecção de movimento por câmera. Sem throttling, o custo acompanha quase todos os AUs de vídeo do RTSP (~15–30 FPS), o que satura VPS compartilhadas.

## Variáveis

| Variável | Default (normal) | Com `ANALYSIS_LEGACY_VPS=1` |
|----------|------------------|-----------------------------|
| `MOTION_ANALYSIS_MAX_FPS` | 5 | 2 |
| `MOTION_FRAME_STRIDE` | 1 | 3 |
| `DECODE_FRAME_STRIDE` | 1 | 5 |

- **`MOTION_ANALYSIS_MAX_FPS`**: teto de decode+motion por câmera (0 = sem limite).
- **`MOTION_FRAME_STRIDE`**: só considera 1 AU a cada N (análogo a `MOTION_FRAME_SKIP` no worker Python).
- **`DECODE_FRAME_STRIDE`**: exige que o índice do AU seja múltiplo de N **e** de `MOTION_FRAME_STRIDE` (análogo a `FRAME_SKIP`).
- **`ANALYSIS_LEGACY_VPS=1`**: ativa os defaults da coluna “legacy” acima; variáveis explícitas ainda podem sobrescrever.

O gate roda **no produtor RTSP**: AUs descartados não entram na fila (menos cópia e menos decode). Após descarte por **FPS**, o próximo enqueue exige **keyframe** com reset do decoder H.264. Keyframes **não** ignoram o teto de FPS.

## EasyPanel / Hostinger (piloto)

Para 2–5 câmeras em VPS já usada pelo stack Python:

```env
ANALYSIS_LEGACY_VPS=1
MAX_CAMERAS=3
MEDIAMTX_NODE_ID=1
```

Ajuste fino: `MOTION_ANALYSIS_MAX_FPS=2`, `MOTION_FRAME_STRIDE=3`, `DECODE_FRAME_STRIDE=5`.

Monitore `/health` e métricas (`limiting_resource`, FPS estimado, CPU do container). Suba `MAX_CAMERAS` só depois de CPU estável.
