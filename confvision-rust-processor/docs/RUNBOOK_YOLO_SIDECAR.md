# Runbook — sidecar YOLO (foxpro / rust-yolo-sidecar)

## Sintomas

- `/health` dos pilots: `yolo_enabled=true` mas eventos não sobem; logs Rust `yolo http timeout`
- EasyPanel: sidecar **502** ou CPU/mem 0%
- Log sidecar: `Corrupt JPEG`, falha download `yolov8n.pt`

## Verificação rápida

```bash
node confvision-rust-processor/scripts/d3-online-test.mjs
bash confvision-rust-processor/scripts/phase-d6-verify.sh
```

Esperado sidecar: `POST /v1/detect` → **400** (jpeg vazio) ou **200** (inferência), não **502**.

## Pilots — env

```env
YOLO_ENABLED=1
YOLO_BACKEND=http
YOLO_HTTP_URL=http://foxpro_rust-yolo-sidecar:8091
```

Sem `/v1/detect` na URL (Rust concatena).

## Sidecar — env

```env
YOLO_MODEL=yolov8n.pt
YOLO_DEVICE=cpu
YOLO_HTTP_PORT=8091
YOLO_CONFIG_DIR=/tmp/Ultralytics
```

## Build EasyPanel (pip / hash)

O sidecar usa `requirements-yolo-sidecar.txt` + `ultralytics==8.3.0 --no-deps` (evita `opencv-python` duplicado e a cadeia 8.4+ com `matplotlib`/`polars` que falha com hash mismatch em rede lenta). Se o build falhar, **Redeploy** de novo; timeout pip: `PIP_DEFAULT_TIMEOUT=300` no Dockerfile.

## Ações

1. **Redeploy** app `rust-yolo-sidecar` (branch `rust-pilot`, `Dockerfile.yolo-sidecar`).
2. Logs: modelo baixado 100%, linha `[sidecar] model=... port=8091`.
3. **RAM VPS:** sidecar + 2 processors no mesmo host → `capacity_state=critical`; reduzir câmeras ou subir RAM.
4. **Timeout:** processor usa **60s** HTTP YOLO; inferência CPU cold pode levar ~15–30s aquecido.
5. **JPEG inválido:** revisar snapshot no Rust se Huffman repetido em toda requisição.

## Rollback analítico (sem derrubar processors)

```env
YOLO_ENABLED=0
CAPTURE_ENABLED=0
```

Redis D2 pode permanecer ativo.
