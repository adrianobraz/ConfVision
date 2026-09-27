# D3 — YOLO, eventos e captura (Rust)

## Objetivo

Paridade mínima com o worker Python (`detector.py` → `event_queue` → `capture_workers` → `processar_deteccao`):

1. **Motion gate** (já Fase 6.3) + **YOLO** por câmera.
2. **Regras** — zonas (`areas`), `modo_deteccao`, `cooldown_seg`, conf mínima, classe pessoa.
3. **Snapshot** local + **`EventJob`** na fila Redis (`confvision:eventos`, D2).
4. **Pool de captura** — `BRPOP` na fila, `POST /vis_evento`, clip ffmpeg, `POST /vis_evento_finalizar`, upload S3 (quando configurado).

Sem Xano novo; API **Go** (`CONFVISION_API_URL`).

## Backends YOLO

| `YOLO_BACKEND` | Uso |
|----------------|-----|
| `off` | Só decode+motion (default). |
| `http` | Sidecar HTTP (`YOLO_HTTP_URL`) — ver `scripts/yolo_sidecar.py`. |
| `onnx` | Feature Cargo `yolo-onnx` + `YOLO_MODEL_PATH` (CPU/GPU D4). |

```env
YOLO_ENABLED=1
YOLO_BACKEND=http
YOLO_HTTP_URL=http://127.0.0.1:8091
YOLO_CONF_DEFAULT=0.5
YOLO_FRAME_STRIDE=5
```

Motion gate: reutiliza `ANALYSIS_ONLY_ON_MOTION` / `YOLO_ONLY_ON_MOTION` (alias).

## Captura (substitui `CAPTURE_WORKERS` Python)

```env
CAPTURE_ENABLED=1
CAPTURE_WORKERS=4
CAPTURE_DIR=/tmp/confvision
CLIP_DURACAO_SEG=20
SNAPSHOT_JPEG_QUALITY=85
# S3 (opcional — upload após captura local)
# S3_ENDPOINT=...
# S3_BUCKET=...
# S3_ACCESS_KEY=...
# S3_SECRET_KEY=...
```

Com `QUEUE_BACKEND=redis` + `CAPTURE_ENABLED=1`, o processor consome a **mesma** fila que publica (compatível com worker off).

## EasyPanel (pilots A/B)

Manter D2 (`REDIS_URL`, `QUEUE_BACKEND=redis`) e acrescentar D3:

```env
YOLO_ENABLED=1
YOLO_BACKEND=http
YOLO_HTTP_URL=http://foxpro_rust-yolo-sidecar:8091
CAPTURE_ENABLED=1
CAPTURE_WORKERS=2
CAPTURE_DIR=/tmp/confvision
```

App **foxpro/rust-yolo-sidecar** (EasyPanel): branch `rust-pilot`, caminho de build `/`, Dockerfile `Dockerfile.yolo-sidecar` na raiz do repo; domínio → porta **8091**.

Redeploy dos pilots com commit D3+ ou feature `yolo-onnx` + modelo em `/app/models/yolov8n.onnx`.

## Verificar

```bash
bash confvision-rust-processor/scripts/d3-event-verify.sh \
  "https://foxpro-rust-pilot.rkr351.easypanel.host" \
  "https://foxpro-rust-pilot-b.rkr351.easypanel.host"
```

Esperado no `/health` (D3):

- `yolo_enabled`, `yolo_backend`, `yolo_device`
- `capture_enabled`, `capture_workers`
- `events_published`, `events_queue_full` (contadores)

Teste funcional: movimento + pessoa na cena → `LLEN confvision:eventos` sobe e desce; evento em Postgres (`vis_evento`).

## Rollback

```env
YOLO_ENABLED=0
CAPTURE_ENABLED=0
```

Fila pode permanecer `redis` (sem publicação).

## Referências código Python

- `confvision/detection_handler.py`, `event_capture.py`, `capture_workers.py`
- `confvision/area_utils.py`
