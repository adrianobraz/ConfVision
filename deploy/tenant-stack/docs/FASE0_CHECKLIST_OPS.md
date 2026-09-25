# Fase 0 — Checklist ops (piloto estável)

Objetivo: **1 câmera online** no Rust com métricas confiáveis, **sem mudar arquitetura** do binário.  
Sem Xano — Postgres + API Go + EasyPanel/Proxmox.

Marque na ordem.

---

## A. Pré-requisitos globais (uma vez)

- [ ] API Go acessível: `CONFVISION_API_URL` (Rust) / `XANO_BASE_URL` (worker Python) = mesma base (ex.: `https://vision.confmonit2.com.br`).
- [ ] `VIS_WORKER_API_KEY` e `RTMP_PUBLISH_SECRET` **iguais** em MTX, worker e Rust.
- [ ] `vis_mediamtx_node_id` das câmeras alinhado a `MEDIAMTX_NODE_ID` (foxpro costuma usar `1`).
- [ ] Redis/Postgres centrais operacionais (worker produção; Rust piloto pode `QUEUE_BACKEND=none`).

---

## B. Stack mínima ligada (tenant / foxpro)

Ordem de start:

1. [ ] **MediaMTX** (`confvision`, `Dockerfile-mediamtx`) — Running  
2. [ ] **Publisher** (`confvision-worker`) — Running  
3. [ ] **Rust** (`rust-pilot` / `confvision-rust-processor`) — Running  

Rede:

- [ ] Rust e worker resolvem o host RTSP (`getent hosts` / ping interno).
- [ ] `MEDIAMTX_RTSP_BASE` usa hostname **interno** da rede Docker (ex.: `rtsp://foxpro_confvision:8554` no EasyPanel, ou `rtsp://mtx:8554` no Compose local).

Referência: [RECUPERACAO_CAMERAS.md](../../../confvision-rust-processor/docs/RECUPERACAO_CAMERAS.md).

---

## C. Banco (1 câmera piloto)

- [ ] Câmera de teste: ativa, analítico conforme regras do Go (`vis_camera_sync_ativas`).
- [ ] Anotar `worker_id` **anterior**.
- [ ] `UPDATE vis_camera SET worker_id = '<PROCESSOR_ID>' WHERE id = <ID>;`  
  Ex.: `rust-processor-pilot-01` ou `ct_cli_<cliente>`.
- [ ] Aguardar 1–2× `SYNC_INTERVAL_SEC` ou reiniciar Rust.

---

## D. Smoke automatizado (opcional)

No host/CT com Docker:

```bash
cd deploy/tenant-stack
export TENANT_ID=ct_cli_rva
export RUST_HEALTH_URL=http://127.0.0.1:8090/health
export MEDIAMTX_RTSP_HOST=mtx   # ou foxpro_confvision
./scripts/tenant-stack-smoke.sh
```

---

## E. Critérios de sucesso (GO)

Logs Rust:

- [ ] `worker started camera_id=... url=rtsp://.../cam/...`
- [ ] `rtsp connected`
- [ ] `frame received` (periódico)

HTTP:

- [ ] `/ready` → `"ready": true` (após warm-up)
- [ ] `/health` → `cameras_online: 1`, `fps_total > 0`
- [ ] `/metrics` → `frames_received` crescente; `rtsp_hotpath.session_next_calls` > 0

Baseline CPU:

- [ ] Anotar `capacity.cpu.percent` com 1 câmera online (referência para sizing GPU).

---

## F. FAIL — ações

| Sintoma | Ação |
|---------|------|
| `Name or service not known` | MTX off ou Rust fora da rede; corrigir hostname |
| RTSP 404 DESCRIBE | Path sem RTMP; worker off ou secret/hash errado |
| `frames_received: 0`, reconnects | Ver B + C |
| Health não responde | Container Rust parado |
| CPU alta, 0 frames | Infra RTSP; não otimizar código antes de B |

---

## G. Baseline host (referência)

- [ ] Com **Rust parado**, anotar CPU global do host (~5–10% típico).
- [ ] Com **1 câmera online**, anotar delta CPU container Rust + `%` em `/metrics`.

Esses números alimentam Fase 2 (GPU, ~400 câmeras/servidor).

---

## Próximo passo

Novo cliente / container Proxmox: [FASE1_PROVISIONAMENTO_TENANT.md](./FASE1_PROVISIONAMENTO_TENANT.md).
