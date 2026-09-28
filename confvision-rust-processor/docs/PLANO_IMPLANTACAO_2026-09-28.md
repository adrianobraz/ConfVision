# Plano de implantação — melhorias 2026-09-28

Consolidação do que foi implementado no repositório e passos operacionais pós-deploy.

## Código (rust-pilot + ConfVision Go)

| Fase | Item | Onde |
|------|------|------|
| A | RTSP 404 backoff dedicado | `STREAM_DELAY_404_SEC`, `stream_policy` PathAbsent |
| A | YOLO async + limite in-flight | `YOLO_INFER_ASYNC`, `YOLO_MAX_INFLIGHT`, `DetectionContext` |
| A | Timeout YOLO HTTP | `YOLO_HTTP_TIMEOUT_SEC` |
| A | Modo incidente (pessoa → sem YOLO, só snapshot) | `detection/coordinator.rs` |
| A | Cenário vs movimento | `MOTION_SCENE_SHIFT_PERCENT`, `motion/detector.rs` |
| B | Sidecar health + fila + workers | `yolo_sidecar.py`, env `YOLO_WORKERS=auto` |
| C | Load shedding CPU/RAM | `LOAD_SHEDDING_*`, `load/shedding.rs` |
| C | CAPTURE_WORKERS=auto | `config.rs` |
| D | Player HLS estável | `live-player.js`, `?stable=1`, retry ≥6 |
| E | d3-online-test `--quick` / full | `scripts/d3-online-test.mjs` |

## Operacional (você / EasyPanel)

1. **Opção A / RTMP:** preencher `rtsp_url_sec` ou RTMP estável; `opcao-a-run.ps1 -Apply` (CSV local gitignored).
2. **Rotacionar** `VIS_WORKER_API_KEY` se exposta.
3. **Redeploy** rust-pilot A/B + sidecar após merge.
4. **Sidecar env:** `YOLO_WORKERS=auto`, `YOLO_MAX_QUEUE=8`, `YOLO_INFER_SLOTS=2` (ajustar CPU do container).
5. **Pilot env:** copiar de `easypanel.env.fase-c.vps.example` (D3 + shedding + 404).
6. **ConfVision Go (RVA):** deploy com `live-player.js` atualizado.
7. **Validar:** `node scripts/d3-online-test.mjs --quick` e `capacity-report`.

## Testes

```bash
# Rápido (~30s) — só health pilots
node confvision-rust-processor/scripts/d3-online-test.mjs --quick

# Completo — sidecar + inferência (pode levar minutos em CPU)
node confvision-rust-processor/scripts/d3-online-test.mjs --full
```

## Referências

- [OPCAO_A_CHECKLIST_FOXPRO.md](./OPCAO_A_CHECKLIST_FOXPRO.md)
- [LOAD_ADMISSION.md](./LOAD_ADMISSION.md)
- [RUNBOOK_YOLO_SIDECAR.md](./RUNBOOK_YOLO_SIDECAR.md)
