# EasyPanel foxpro — Environment Variables (PostgreSQL)

Cole o conteúdo do arquivo correspondente em **Environment** de cada serviço e faça **Redeploy**.

Repo GitHub workers: `adrianobraz/ConfVision` (branch `main`).

API operacional (substitui Xano): **`https://vision.confmonit2.com.br`**

---

## Qual arquivo usar em cada serviço

| Serviço EasyPanel | Arquivo local | Entry point |
|-------------------|---------------|-------------|
| **confvision** (MediaMTX + Guard) | `mediamtx.env` | `Dockerfile-mediamtx` |
| **confvision-worker** | `worker.env` | `python -u main.py` |
| **confvision-dvr** | `dvr.env` | `python -u dvr_main.py` |
| **confvision-motion** | `motion.env` | `python -u motion_main.py` |
| **confvision-timelapse** | `timelapse.env` | `python -u timelapse_main.py` |

> **Parar:** `confvision-rtmp-guard` separado (legado) se já usa `Dockerfile-mediamtx`.

---

## Variáveis críticas (todos os serviços com API)

| Variável | Valor | Onde |
|----------|-------|------|
| `XANO_BASE_URL` | `https://vision.confmonit2.com.br` | worker, dvr, motion, timelapse, mediamtx |
| `RTMP_PUBLISH_SECRET` | **igual** app Go ConfVision | todos acima |
| `MEDIAMTX_NODE_ID` | `1` (id em `vis_mediamtx_node`) | worker, dvr, motion, timelapse |

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
