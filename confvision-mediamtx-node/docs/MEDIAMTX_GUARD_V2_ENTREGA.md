# MediaMTX + Guard V2 — entrega operacional

Camada **Streaming Plane** do ConfVision: NVR → **Guard** → **MediaMTX** → **RTSP** → Rust (inalterado nesta etapa).

---

## A. Inventário

### Utilizados (código existente — sem reescrita)

| Área | Caminho |
|------|---------|
| Auth / regras | `confvision-python/rtmp_guard.py` (`RtmpGuard`, `CameraCache`) |
| HTTP Guard | `confvision-python/rtmp_guard_main.py` |
| Ban | `confvision-python/rtmp_ban.py` |
| Path/hash | `confvision-python/rtmp_token.py` (`cam/{hash12+}`) |
| Falhas / tail log | `confvision-python/rtmp_watch.py` |
| Online / API MTX | `confvision-python/rtmp_online.py` |
| Publish health | `confvision-python/rtmp_publish_health.py` |
| Stream health → CP | `confvision-python/stream_health_client.py` |
| Mensagens | `confvision-python/rtmp_messages.py` |
| Startup | `confvision-python/start_mediamtx_guard.py` |
| Config legado | `confvision-python/mediamtx/mediamtx.yml` |
| Docker legado EasyPanel | `confvision-python/Dockerfile.mediamtx` |
| Env foxpro | `confvision-python/easypanel/mediamtx.env` |
| Docs operação | `confvision-python/DEPLOY_EASYPANEL.md`, `docs/VALIDACAO_RTMP_GUARD.md` |

### Novos (V2)

| Caminho | Função |
|---------|--------|
| `confvision-mediamtx-node/deploy/Dockerfile.v2` | Imagem canônica V2 (MTX 1.21.0 + config organizada) |
| `confvision-mediamtx-node/config/mediamtx.confvision.yml` | MediaMTX V2 (RTMP/RTSP/HLS/API/metrics/playback) |
| `confvision-mediamtx-node/config/overlays/production-legacy-hls.yaml` | Paridade HLS legado (`hlsAlwaysRemux: true`) |
| `confvision-mediamtx-node/scripts/mtx-node-health.ps1` / `.sh` | Health Guard + API |
| `confvision-mediamtx-node/scripts/test-guard-auth.ps1` | Testes POST `/auth` (path inválido, etc.) |
| Este documento | Entrega A–H |

### Alterados

| Caminho | Mudança |
|---------|---------|
| `start_mediamtx_guard.py` | Sequência Guard → MTX → API; preflight porta 8100; `NODE READY` |
| `rtmp_guard_main.py` | `GET /health/ready` |
| `rtmp_online.py` | `mediamtx_api_reachable()` |
| `Dockerfile.mediamtx` / `Dockerfile-mediamtx` | Pin `mediamtx:1.21.0` |
| `confvision-python/mediamtx/mediamtx.yml` | Comentário apontando V2 |

### Mantidos sem alteração de lógica

- Rust (`confvision-rust-processor/`)
- Go Control Plane (`home/confmonit/v4.0/confcam/`)
- XanoScript / Postgres schema
- YOLO / workers Python analíticos

---

## B. Arquitetura

```text
CAMERA / NVR
     │ RTMP :1935  path cam/{hash12+}
     ▼
┌─────────────────────────────────────┐
│  Container confvision (V2)          │
│  ┌─────────────┐   authHTTP         │
│  │ RTMP-GUARD  │◄──────────────────┐ │
│  │ :8100       │                   │ │
│  └──────┬──────┘                   │ │
│         │ GET /vis_camera/rtmp_auth│ │
│         ▼ (cache local TTL)        │ │
│  Control Plane (Go) ───────────────┘ │
│  ┌─────────────┐                     │
│  │ MediaMTX    │ RTSP :8554          │
│  │ API :9997   │ HLS :8888           │
│  │ metrics:9998│                     │
│  └─────────────┘                     │
└─────────────────────────────────────┘
     │ RTSP rtsp://foxpro_confvision:8554/cam/{hash}
     ▼
Rust processor (existente)
```

**Redis:** usado por workers/sync/event queue — **não** é source of truth do Guard. Auth usa `CameraCache` in-process + API Go.

---

## C. Configuração

### Portas (projeto real)

| Porta | Serviço | Exposição típica |
|-------|---------|------------------|
| **1935** | RTMP ingest | Pública (NVR) |
| **8554** | RTSP | Rede Docker / interna (Rust) |
| **8888** | HLS | Conforme painel / firewall |
| **9997** | MediaMTX API | Interna + credencial `dvr` via Guard |
| **9998** | Métricas Prometheus | Interna (V2 config) |
| **9996** | Playback gravações | Interna (V2) |
| **8100** | Guard admin + `/auth` | **Restrita** (Go + ops); MTX usa `127.0.0.1:8100/auth` no mesmo container |

### Variáveis obrigatórias (secrets fora do Git)

Copiar de `confvision-python/easypanel/mediamtx.env` ou `confvision-mediamtx-node/deploy/easypanel.env.example`:

- `RTMP_PUBLISH_SECRET` — Hashids (path `cam/…`)
- `VIS_WORKER_API_KEY` — `GET /vis_camera/{id}/rtmp_auth`
- `CONFVISION_API_URL` — base Go (stream health / pause)
- `MEDIAMTX_API_USER` / `MEDIAMTX_API_PASS`
- `RTMP_GUARD_ADMIN_KEY` — `/falhas`, `/bans`, `/online`

Opcionais úteis: `RTMP_BAN_MAX_FAILS=12`, `RTMP_AUTH_CACHE_SEC=10`, identidade nó `CONFVISION_MEDIAMTX_NODE_ID`.

### Escolha de imagem

| Cenário | Dockerfile | Config YAML |
|---------|------------|-------------|
| **Foxpro redeploy mínimo** | `confvision-python/Dockerfile.mediamtx` | `mediamtx/mediamtx.yml` (HLS remux on) |
| **V2 recomendado (novo build)** | `confvision-mediamtx-node/deploy/Dockerfile.v2` | `mediamtx.confvision.yml` (+ overlay HLS legado se necessário) |

Build V2 (raiz `core4`):

```bash
docker build -f confvision-mediamtx-node/deploy/Dockerfile.v2 -t confvision-mediamtx-node:2.0.0 .
```

Volume: host `/opt/confvision/recordings` → `/recordings` (bans, logs, falhas, segmentos).

---

## D. Startup

### Sequência (container)

1. Preflight: porta **8100** livre (evita `confvision-rtmp-guard` legado duplicado)
2. **Guard** (`rtmp_guard_main.py`) → `GET /health` OK
3. **MediaMTX** (`/mediamtx /mediamtx.yml`)
4. API `GET /v3/config/global/get` OK
5. Log: **`NODE READY`**

### EasyPanel foxpro

1. Serviço **`confvision`**: **Start** (ou Redeploy após mudar Dockerfile)
2. **`confvision-rtmp-guard`**: manter **STOPPED** se usa imagem combinada
3. Env: colar `easypanel/mediamtx.env`
4. Dockerfile path:
   - Legado: `Dockerfile.mediamtx`
   - V2: `confvision-mediamtx-node/deploy/Dockerfile.v2` (ajustar path no painel conforme repo `core4`)

Log esperado:

```text
[START] Guard RTMP ...
[START] Guard HTTP pronto (http://127.0.0.1:8100/health)
[START] MediaMTX /mediamtx /mediamtx.yml ...
[START] MediaMTX API pronta ...
[START] NODE READY — Guard + MediaMTX ...
```

---

## E. Health

| Check | Comando |
|-------|---------|
| Guard vivo | `curl -s http://127.0.0.1:8100/health` |
| Nó pronto (Guard + MTX API) | `curl -s http://127.0.0.1:8100/health/ready` → HTTP 200, `"status":"ready"` |
| MediaMTX paths | `curl -s -u dvr:*** http://127.0.0.1:9997/v3/paths/list?itemsPerPage=5` |
| Script | `confvision-mediamtx-node/scripts/mtx-node-health.ps1 -GuardBase http://127.0.0.1:8100` |

RTMP/RTSP listening: na VPS, `ss -lntp | egrep '1935|8554'` ou publish de teste.

Redis: necessário para **workers**, não para subir Guard/MTX.

---

## F. Teste (ordem)

### 1. Infra (sem câmera)

- Guard `/health` → `secret_configured: true`
- `/health/ready` → `mediamtx_api: ok`
- API paths list → 200

### 2. Rejeição Guard (`test-guard-auth.ps1`)

- Path inválido → 403 `path_invalido`
- Hash curto → 403
- IP banido / câmera inativa → teste com env + API real (manual)

### 3. Uma câmera

1. Path RTMP: `rtmp://<host>:1935/cam/{hash12}` (hash da câmera no cadastro, não id numérico UI)
2. Guard log: `publish_ok` ou motivo claro
3. `GET /online` (admin key) ou API paths → path online
4. RTSP: `rtsp://foxpro_confvision:8554/cam/{hash}` (Rust depois, etapa seguinte)

---

## G. Rollback

Ver `docs/ROLLBACK.md`. Resumo:

- Redeploy imagem/tag anterior + `Dockerfile.mediamtx` + `mediamtx/mediamtx.yml`
- Volume `/recordings` preservado (bans, logs)
- Secrets Guard **iguais** — só troca streaming plane

---

## H. Problemas encontrados / correções

| Arquivo | Função / área | Problema | Correção |
|---------|----------------|----------|----------|
| VPS foxpro | Serviço parado | Camada MTX+Guard down → todas câmeras paradas | **Start** serviço `confvision`; doc D acima |
| `start_mediamtx_guard.py` | startup | MTX podia subir antes do Guard estável; sem sinal NODE READY | Sequência + wait API + preflight 8100 |
| `rtmp_guard_main.py` | health | Só `/health` não indicava MTX | `/health/ready` |
| `Dockerfile.mediamtx` | imagem | Tag `mediamtx:1` flutuante | Pin **1.21.0** |
| Legado vs V2 HLS | config | `hlsAlwaysRemux: yes` vs `false` | Overlay `production-legacy-hls.yaml` |
| `easypanel/mediamtx.env` | deploy | `RTMP_BAN_MAX_FAILS=3` agressivo | Ajustar **12** + redeploy (sessão anterior) |
| Guard duplicado | ops | Dois containers na **8100** | Stop `confvision-rtmp-guard` legado |

**Interface Rust (documentação apenas):** consumir `MEDIAMTX_RTSP_BASE` + path `cam/{hash}`; cadastro `rtsp_url_sec` deve usar hostname Docker `foxpro_confvision:8554`, não DNS público.

---

## Próximo passo (fora desta etapa)

Com Guard + MTX **RUNNING**: teste 1 câmera RTMP → RTSP; depois **Start** Rust pilot (`WORKER_ID` alinhado ao Postgres).
