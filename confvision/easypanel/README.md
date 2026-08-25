# EasyPanel foxpro — Environment Variables (PostgreSQL)

Cole o conteúdo do arquivo correspondente em **Environment** de cada serviço e faça **Redeploy**.

Repo GitHub workers: `adrianobraz/core4` → pasta `confvision/` (branch `main`).

API operacional (substitui Xano): **`https://vision.confmonit2.com.br`**

---

## Worker analítico — modo distributed (600 câmeras)

| Variável | Valor | Por quê |
|----------|-------|---------|
| `YOLO_ARCH` | `distributed` | capture MOG2 + batch GPU (sem `_infer_lock`) |
| `SCHEDULER_BACKEND` | `redis` | fila YOLO P1–P4 compartilhada entre processos |
| `EVENT_QUEUE_BACKEND` | `redis` | fila detecção → clip (não usar `memory` em produção) |
| `REDIS_URL` | Memurai Proxmox `185.130.61.5:6379` | **mesmo Redis** para as 3 filas acima |
| `MAX_CAMERAS` | `600` | limite por container (1 GEX44 = 1 worker) |
| `YOLO_DEVICE` | `cuda:0` | GPU dedicada |
| `FRAME_STORE_DIR` | `/dev/shm/confvision/frames` | frames evidência + latest em RAM |

**Mount obrigatório no confvision-worker:**

| Host | Container |
|------|-----------|
| `/dev/shm/confvision` | `/dev/shm/confvision` |

**Redis — o que fazer:** não instalar nada novo. Só garantir `REDIS_URL` + `SCHEDULER_BACKEND=redis` + `EVENT_QUEUE_BACKEND=redis` no Environment do EasyPanel. Log esperado: `[SCHEDULER] backend=redis key=confvision:yolo:queue`.

Template completo: [`worker.env.example`](worker.env.example) (sem senhas). Produção: copie de `worker.env` local (gitignored).

**Log esperado (distributed):**
```text
[START] ConfVision distributed | arch=distributed ...
[SCHEDULER] backend=redis key=confvision:yolo:queue
[YOLO-GPU] model=yolov8n.pt device=cuda:0 batch=sim
[GPU-SCHED] iniciado batch_size=16 timeout_ms=10
[EVENTO] ... MOTION_TO_EVENT_MS=180 ...
```

Para voltar ao modo antigo: `YOLO_ARCH=legacy`.

---

## Qual arquivo usar em cada serviço

| Serviço EasyPanel | Arquivo local | Entry point |
|-------------------|---------------|-------------|
| **confvision** (MediaMTX + Guard) | `mediamtx.env` | `Dockerfile-mediamtx` |
| **confvision-worker** | `worker.env` | `python -u main.py` |
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
| **confvision-worker** (distributed) | `/dev/shm/confvision` | `/dev/shm/confvision` |

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
