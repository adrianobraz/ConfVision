# Fase 4.1A — Alinhamento workers (fechamento)

**Objetivo:** Rust usa a **mesma semântica** que `confvision/sharding.py` e o ping **`vis_worker`**, para coexistir com workers Python (outros `worker_tipo`) sem disputar câmera.

---

## Mapa env (Rust ↔ Python)

| Variável | Python (`config.py`) | Rust (`config.rs`) | Uso |
|----------|----------------------|---------------------|-----|
| `WORKER_ID` | ✅ | ✅ | Sync query (modo worker_id), ping, assign Postgres |
| `WORKER_TIPO` | default `analitico` | default `rust_processor` | Chave `vis_worker` com Python |
| `SHARD_MODE` | auto / worker_id / hash | idem | Query API + filtro local |
| `WORKER_SHARD_INDEX` | ✅ | ✅ | hash: `camera_id % total == index` |
| `WORKER_SHARD_TOTAL` | ✅ | ✅ | hash |
| `MAX_CAMERAS` | truncagem pós-filtro | idem (`effective_max_cameras`) | Hard limit local |
| `MEDIAMTX_NODE_ID` | query `vis_mediamtx_node_id` | idem | Fleet MediaMTX |
| `PROCESSOR_ID` | — | ✅ | **Só** métricas/`/health`; **não** entra no sync |

Implementação Rust: `src/sharding/mod.rs` (comentário: espelho de `sharding.py`).

---

## Produção foxpro (A/B)

| App | `WORKER_ID` típico | `SHARD_MODE` | `SYNC_FILTER_WORKER_ID` |
|-----|-------------------|--------------|---------------------------|
| rust-pilot A | `rust-processor-pilot-a-01` | `worker_id` | `true` |
| rust-pilot B | `rust-processor-pilot-b-02` | `worker_id` | `true` |

Go filtra `vis_camera.worker_id` na sync; Rust **não** aplica hash local nesse modo.

### Modo `hash` (600 cams / 1 GPU, vários processos)

- Python: **não** envia `worker_id` na query; filtra hash **localmente**.
- Rust: defina **`SYNC_FILTER_WORKER_ID=false`** para igualar a query; hash continua em `filter_analytic_cameras`.

---

## Fluxo validado

```text
Postgres vis_camera.worker_id
        ↓
GET /vis_camera_sync_ativas?worker_id=…  (Go)
        ↓
filter_analytic_cameras (Rust: ativo, deteccao_humano, !pausado, shard, MAX_CAMERAS)
        ↓
CameraManager
        ↓
POST /vis_worker_ping → vis_worker (shard_*, max_cameras, cameras_ativas, …)
```

---

## Verificação

### Monitor (CT111 — **sem** Rust/cargo)

```bash
export CONFVISION_API_URL=https://vision.confmonit2.com.br
export VIS_WORKER_API_KEY='…'
bash confvision-rust-processor/scripts/phase-4.1a-verify.sh
```

### Build host (fmt + testes + FFmpeg)

```bash
bash confvision-rust-processor/scripts/phase-4.1-verify.sh
```

---

## Coexistência Rust + Python

- **Mesmo `worker_id` + mesmo `worker_tipo`:** colisão — não usar.
- **Mesmo `worker_id`, tipos diferentes:** linhas distintas em `vis_worker`; assign de câmera é **um** `worker_id` no Postgres — só um processor deve receber sync daquela câmera.
- Ingest Python **off** na foxpro: apenas Rust `rust_processor` nos pings.

---

## Próxima fase

**Fase 5** — [FASE_5.md](./FASE_5.md) · `phase-f5-verify.sh` · `phase-f5-ramp-run.sh`

Ver: [FASE_4_1_FECHAMENTO.md](./FASE_4_1_FECHAMENTO.md), [FASE_D_FECHAMENTO.md](./FASE_D_FECHAMENTO.md).
