# Consumo de CPU — análise (decode + motion)

O piloto Rust decodifica H.264 na CPU e roda detecção de movimento por câmera. Sem throttling, o custo acompanha quase todos os AUs de vídeo do RTSP (~15–30 FPS), o que satura VPS compartilhadas.

## Variáveis

| Variável | Default (normal) | Com `ANALYSIS_LEGACY_VPS=1` |
|----------|------------------|-----------------------------|
| `MOTION_ANALYSIS_MAX_FPS` | 5 | 2 |
| `MOTION_FRAME_STRIDE` | 1 | 3 |
| `DECODE_FRAME_STRIDE` | 1 | 5 |
| `ANALYSIS_ONLY_ON_MOTION` | 0 | 1 (≈ `YOLO_ONLY_ON_MOTION`) |
| `MOTION_GATE_PROBE_MAX_FPS` | 1 | 0.5 |
| `MOTION_GATE_MISS_FRAMES` | 10 | `YOLO_MOTION_MISS_FRAMES×5` (default 10) |

- **`MOTION_ANALYSIS_MAX_FPS`**: teto de decode+motion por câmera (0 = sem limite).
- **`ANALYSIS_ONLY_ON_MOTION`**: cena parada → só **probes** baratos (`MOTION_GATE_PROBE_MAX_FPS`); após movimento → taxa **armed** (max FPS + strides). Desarma após `MOTION_GATE_MISS_FRAMES` análises sem movimento.
- **`MOTION_FRAME_STRIDE`**: só considera 1 AU a cada N (análogo a `MOTION_FRAME_SKIP` no worker Python).
- **`DECODE_FRAME_STRIDE`**: exige que o índice do AU seja múltiplo de N **e** de `MOTION_FRAME_STRIDE` (análogo a `FRAME_SKIP`).
- **`ANALYSIS_LEGACY_VPS=1`**: ativa os defaults da coluna “legacy” acima; variáveis explícitas ainda podem sobrescrever.

O gate roda **no produtor RTSP**: AUs descartados não entram na fila (menos cópia e menos decode). Após descarte por **FPS**, o próximo enqueue exige **keyframe** com reset do decoder H.264. Keyframes **não** ignoram o teto de FPS.

## EasyPanel / Hostinger (piloto)

Para 2–5 câmeras em VPS já usada pelo stack Python:

```env
ANALYSIS_LEGACY_VPS=1
ANALYSIS_ONLY_ON_MOTION=1
MAX_CAMERAS=3
MEDIAMTX_NODE_ID=1
```

Paridade Python (piloto Rust ainda **sem YOLO**): MOG2/YOLO no Python ≈ **probe decode + luma diff** idle, **decode frequente** só com movimento.

Com `ANALYSIS_LEGACY_VPS=1` ou `ANALYSIS_ONLY_ON_MOTION=1` (defaults):

| Variável | Efeito |
|----------|--------|
| `MOTION_PROBE_KEYFRAME_ONLY=1` | Probe idle só em IDR |
| `RTSP_IDLE_SUSPEND=1` | Desconecta RTSP entre probes |
| `MOTION_PIXEL_DIFF_THRESHOLD` / `MOTION_PERCENT_THRESHOLD` | Menos falso “armed” |

Ajuste fino: `MOTION_ANALYSIS_MAX_FPS=2`, `MOTION_FRAME_STRIDE=3`, `DECODE_FRAME_STRIDE=5`.

Monitore `/health` e métricas (`limiting_resource`, FPS estimado, CPU do container). Suba `MAX_CAMERAS` só depois de CPU estável.
