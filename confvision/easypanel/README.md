# EasyPanel foxpro — Environment Variables (PostgreSQL)

Cole o conteúdo do arquivo correspondente em **Environment** de cada serviço e faça **Redeploy**.

Repo GitHub workers: `adrianobraz/core4` → pasta `confvision/` (branch `main`).

API operacional (substitui Xano): **`https://vision.confmonit2.com.br`**

---

## Worker analítico — qual env usar?

| Servidor | Arquivo | Modelo | GPU | Mount |
|----------|---------|--------|-----|-------|
| **VPS atual** (foxpro, sem GPU) | [`worker.env.vps-sem-gpu`](worker.env.vps-sem-gpu) | `legacy` | Não | Nenhum |
| **GEX44 dedicado** (com RTX) | [`worker.env.gex44-gpu`](worker.env.gex44-gpu) | `distributed` | `cuda:0` | `/dev/shm/confvision` |

Templates sem senhas (GitHub): [`worker.env.vps-sem-gpu.example`](worker.env.vps-sem-gpu.example) · [`worker.env.gex44-gpu.example`](worker.env.gex44-gpu.example)

> **`worker.env`** espelha a VPS sem GPU enquanto o GEX44 não estiver pronto.

### VPS sem GPU (legacy, ~50 câmeras)

- `YOLO_ARCH=legacy` · `YOLO_DEVICE=` (vazio, CPU)
- `MAX_CAMERAS=50` · `WORKER_SHARD_TOTAL=2`
- **Sem** `SCHEDULER_BACKEND`, **sem** mount `/dev/shm`
- Pode manter **confvision-worker STOPPED** se não houver analítico neste nó

Log esperado:
```text
[START] ConfVision worker | ... yolo_arch=legacy
[YOLO] model=yolov8n.pt device=cpu
```

### GEX44 com GPU (distributed, 600 câmeras)

| Variável | Valor |
|----------|-------|
| `YOLO_ARCH` | `distributed` |
| `YOLO_DEVICE` | `cuda:0` |
| `SCHEDULER_BACKEND` | `redis` |
| `MAX_CAMERAS` | `600` |
| `FRAME_STORE_DIR` | `/dev/shm/confvision/frames` |

Mount: `/dev/shm/confvision` → `/dev/shm/confvision`

Log esperado:
```text
[START] ConfVision distributed | arch=distributed ...
[SCHEDULER] backend=redis key=confvision:yolo:queue
[YOLO-GPU] device=cuda:0
```

---

## Qual arquivo usar em cada serviço

| Serviço EasyPanel | Arquivo local | Entry point |
|-------------------|---------------|-------------|
| **confvision** (MediaMTX + Guard) | `mediamtx.env` | `Dockerfile-mediamtx` |
| **confvision-worker** | `worker.env.vps-sem-gpu` ou `worker.env.gex44-gpu` | `python -u main.py` |
| **confvision-sync-agent** | `sync-agent.env` | `python -u sync_agent_main.py` |
| **confvision-dvr** | `dvr.env` | `python -u dvr_main.py` |
| **confvision-motion** | `motion.env` | `python -u motion_main.py` |
| **confvision-sensor** | `sensor.env` | `python -u sensor_main.py` |
| **confvision-timelapse** | `timelapse.env` | `python -u timelapse_main.py` |

> **Parar:** `confvision-rtmp-guard` separado (legado) se já usa `Dockerfile-mediamtx`.

---

## Variáveis críticas (todos os serviços com API)

| Variável | Valor | Onde |
|----------|-------|------|
| `XANO_BASE_URL` | `https://vision.confmonit2.com.br` | todos os serviços acima |
| `VIS_WORKER_API_KEY` | **igual** `VIS_WORKER_API_KEY` do Go ConfVision (Proxmox) | todos os serviços acima |
| `RTMP_PUBLISH_SECRET` | **igual** app Go ConfVision | mediamtx, worker, dvr, motion, timelapse, sensor |
| `MEDIAMTX_NODE_ID` | `1` (id em `vis_mediamtx_node`) | worker, sync-agent, dvr, motion, timelapse |

## Redis central (Proxmox / Memurai)

| Variável | Valor |
|----------|-------|
| `REDIS_URL` | `redis://default:2008.03.28.Dri.zin@185.130.61.5:6379/0` |
| `CONFIG_CACHE_BACKEND` | `redis` |

> Substitui `foxpro/visionredis` na VPS. Postgres central fica no Proxmox (`10.2.2.120`) — **desligue** `foxpro/visionpsql` após redeploy.

## Só no worker / motion

| Variável | Valor |
|----------|-------|
| `EVENT_STORE` | `postgres` |
| `TERMINAL_NOTIFY_ENABLED` | `false` |

Terminal CV01 fica no **Go central** (`TERMINAL_NOTIFY_ENABLED=true` no Proxmox).

## Só no MediaMTX + Guard

| Variável | Valor |
|----------|-------|
| `RTMP_GUARD_ADMIN_KEY` | igual Go `.env` |
| `MEDIAMTX_API_USER` / `MEDIAMTX_API_PASS` | credenciais API MediaMTX |
| `MEDIAMTX_API_BASE` | `http://127.0.0.1:9997` |

---

## Mount obrigatório

| Serviço | Host | Container |
|---------|------|-----------|
| confvision (MediaMTX) | `/opt/confvision/recordings` | `/recordings` |
| confvision-dvr | `/opt/confvision/recordings` | `/recordings` |
| **confvision-worker** (só GEX44 / distributed) | `/dev/shm/confvision` | `/dev/shm/confvision` |

---

## Log esperado após redeploy

**Worker:**
```text
[CONFIG] OK | xano=https://vision.confmonit2.com.br | event_store=postgres
```

**Guard (MediaMTX):**
```text
[RTMP-GUARD] START | auth+ban+watch
[RTMP-GUARD] OK publish ... path=cam/{hash12}
```

---

## Teste API (Go central)

```bash
curl https://vision.confmonit2.com.br/vis_health
# {"status":"ok","enabled":true,"postgres":"ok"}
```

Ver também: [`../VARIAVEIS_VPS.md`](../VARIAVEIS_VPS.md)
