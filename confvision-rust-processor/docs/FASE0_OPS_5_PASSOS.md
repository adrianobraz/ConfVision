# Fase 0 — cinco passos operacionais

## 1. CPU / admission

1. Nos apps **rust-pilot A e B**, aplicar variáveis de `easypanel.env.fase0-reduce-load.example`:
   - `LOAD_ADMISSION_ENABLED=0`
   - `LOAD_SHEDDING_ENABLED=0`
   - `YOLO_FRAME_STRIDE=10`, `DECODE_FRAME_STRIDE=8`
2. Redeploy A e B.
3. Confirmar `/health`: `capacity_state` abaixo de `critical`, `allow_new_camera=true` em `/capacity-report`.

## 2. Sidecar YOLO

1. App **rust-yolo-sidecar**: env de `easypanel.env.yolo-sidecar.example` (`YOLO_INFER_SLOTS=2`, fila maior).
2. Redeploy sidecar (branch `rust-pilot` — handler trata BrokenPipe e JPEG vazio).
3. Teste: `POST /v1/detect` com JPEG válido → 200; body vazio → 400.
4. URL pública health: `https://foxpro-rust-yolo-sidecar.rkr351.easypanel.host/health`

## 3. Câmera 4 (RTSP 404)

- Path esperado: `rtsp://srv1.dnsid.com.br:8554/cam/boezyjmydlk4`
- **404** = stream **não publicado** no MediaMTX (DVR/RTMP), não bug do Rust.
- Republicar RTMP com hash da câmera 4 ou corrigir DVR.
- Runbook: `docs/RUNBOOK_CAMERAS_404_ATIVO.md`

## 4. Evento Fase 0 (câmera 15)

1. Postgres: `sql/fase0_pilot_ops.sql` ou `go run` em `core4/.../scripts/fase0-apply-ops`
2. Garantir: `worker_id=rust-processor-pilot-b-02`, `analitico_pausado=false`
3. Após sync (~60s): cam 15 nas métricas do B; movimento + pessoa na cena ou teste controlado
4. Validar: `events_published>0` e linha em `vis_evento` para id 15

## 5. Testes repo

```powershell
cd confvision-rust-processor
cargo test
cd ..\..\core4\home\confmonit\v4.0\confvision
go test ./src/modulos/confvision/... ./src/modulos/visdata/...
cd scripts
.\fase0-verify.ps1 -WorkerKey '...'
```
