# ConfVision — Deploy EasyPanel (foxpro)

Documentação dos serviços Python/MediaMTX do ConfVision no EasyPanel.

- **Projeto EasyPanel:** foxpro
- **Repositório workers:** [adrianobraz/ConfVision](https://github.com/adrianobraz/ConfVision) (branch `main`)
- **Pasta local:** `core4/confvision/`

> ConfVision é um sistema **solo**. Cada worker roda a partir desta pasta.
> A app web Go fica em `home/confmonit/v4.0/confvision/` (deploy separado).

---

## Visão geral dos serviços

| Serviço EasyPanel | Função | Imagem / origem | Entry point |
|---|---|---|---|
| `confvision` | MediaMTX (RTSP, HLS, API, gravação local) | `bluenviron/mediamtx:1` | (imagem oficial) |
| `confvision-worker` | Detecção de pessoas (YOLO) + eventos | GitHub `ConfVision` | `python -u main.py` |
| `confvision-dvr` | Gravação contínua (DVR) | GitHub `ConfVision` | `python -u dvr_main.py` |
| `confvision-motion` | Gravação por movimento | GitHub `ConfVision` | `python -u motion_main.py` |
| `confvision-timelapse` | Timelapse inteligente | GitHub `ConfVision` | `python -u timelapse_main.py` |
| `confvision-rtmp-watch` | Monitor de falhas RTMP (legado) | GitHub `ConfVision` | `python -u rtmp_watch_main.py` |
| `confvision-rtmp-guard` | **Auth RTMP + ban/desban + falhas** | GitHub `ConfVision` | `python -u rtmp_guard_main.py` |

### Regra de plano por câmera

Uma câmera usa **um** modo de gravação por vez:

| Plano (licença) | Worker responsável |
|---|---|
| Contínuo (`gravacao_7d`, etc.) | `confvision-dvr` |
| Movimento (`gravacao_movimento_*`) | `confvision-motion` |
| Timelapse (`gravacao_timelapse_*`) | `confvision-timelapse` |

Não ativar DVR + motion + timelapse na mesma câmera.

---

## 1. confvision (MediaMTX)

Servidor de mídia. Câmeras publicam e consomem stream via este container.

### Configuração

- **Tipo:** Docker Image
- **Imagem:** `bluenviron/mediamtx:1`

### Variáveis principais (Environment)

| Variável | Valor (foxpro) | Descrição |
|---|---|---|
| `MTX_RTSPTRANSPORTS` | `tcp` | Transporte RTSP |
| `MTX_API` | `yes` | Habilita API REST |
| `MTX_APIADDRESS` | `:9997` | Porta da API |
| `MTX_PATHDEFAULTS_RECORD` | `no` | Gravação desligada por padrão |
| `MTX_PATHS_LIVE_1_RECORD` | `yes` | Exemplo de path com record |
| `MTX_PATHS_LIVE_1_RECORDPATH` | `/recordings/%path/%Y-%m-%d_%H-%M-%S` | Pasta de gravação |
| `MTX_PATHS_LIVE_1_RECORDFORMAT` | `fmp4` | Formato do segmento |
| `MTX_PATHS_LIVE_1_RECORDSEGMENTDURATION` | `5m` | Duração do segmento |
| `MTX_AUTHMETHOD` | `http` | Auth via confvision-rtmp-guard |
| `MTX_AUTHHTTPADDRESS` | `http://foxpro_confvision-rtmp-guard:8100/auth` | Webhook de auth |
| ~~`MTX_AUTHINTERNALUSERS_…`~~ | — | Removido: publish aberto (`any`) não deve ser usado |

> Preferir montar `mediamtx/mediamtx.yml` (já com `authMethod: http`).

### Variáveis principais (Environment) — legado (NÃO usar em produção)

| Variável | Valor (foxpro) | Descrição |
|---|---|---|
| `MTX_RTSPTRANSPORTS` | `tcp` | Transporte RTSP |
| `MTX_API` | `yes` | Habilita API REST |
| `MTX_APIADDRESS` | `:9997` | Porta da API |
| `MTX_PATHDEFAULTS_RECORD` | `no` | Gravação desligada por padrão |
| `MTX_PATHS_LIVE_1_RECORD` | `yes` | Exemplo de path com record |
| `MTX_PATHS_LIVE_1_RECORDPATH` | `/recordings/%path/%Y-%m-%d_%H-%M-%S` | Pasta de gravação |
| `MTX_PATHS_LIVE_1_RECORDFORMAT` | `fmp4` | Formato do segmento |
| `MTX_PATHS_LIVE_1_RECORDSEGMENTDURATION` | `5m` | Duração do segmento |
| `MTX_AUTHMETHOD` | `internal` | (legado) |

### Log em arquivo (obrigatório para Falhas RTMP)

O MediaMTX precisa gravar log em arquivo no volume compartilhado `/recordings`, para o `confvision-rtmp-watch` ler.

**Opção A — arquivo `mediamtx.yml` montado** (recomendado; ver `mediamtx/mediamtx.yml`):

```yaml
logDestinations: [stdout, file]
logFile: /recordings/mediamtx.log
```

**Opção B — variáveis de ambiente:**

| Variável | Valor |
|---|---|
| `MTX_LOGDESTINATIONS` | `stdout,file` |
| `MTX_LOGFILE` | `/recordings/mediamtx.log` |

### Mount (Bind)

| Host | Container |
|---|---|
| `/opt/confvision/recordings` | `/recordings` |

Mesmo volume usado por DVR e `confvision-rtmp-watch`.

### Rede interna (nome do container)

Outros serviços referenciam este container como:

- RTSP: `rtsp://foxpro_confvision:8554`
- API: `http://foxpro_confvision:9997`
- RTMP: `rtmp://rtmp.confmonit.com.br:1935` (público)

---

## 2. confvision-worker (detecção YOLO)

Processa câmeras com detecção de pessoas ativa. **Não** grava timelapse/DVR.

### Configuração

- **Source:** GitHub → `adrianobraz/ConfVision` / `main` / `/`
- **Comando:** *(vazio — usa CMD do Dockerfile)*
- **Dockerfile CMD:** `python -u main.py`

### Variáveis de ambiente

| Variável | Valor (foxpro) |
|---|---|
| `XANO_BASE_URL` | `https://xpcy-oyme-lno7.b2.xano.io/api:AC7rgWwW` |
| `MEDIAMTX_RTSP_BASE` | `rtsp://foxpro_confvision:8554` |
| `WORKER_ID` | `worker-docker-01` |
| `WORKER_VERSION` | `0.3.0` |
| `SYNC_INTERVAL_SEC` | `30` |
| `FRAME_SKIP` | `5` |
| `YOLO_MODEL` | `yolov8n.pt` |
| `YOLO_CONF_DEFAULT` | `0.5` |
| `CLIP_DURACAO_SEG` | `20` |
| `CAPTURE_DIR` | `/tmp/confvision` |
| `UPLOAD_WORKERS` | `4` |
| `CAPTURE_WORKERS` | `8` |
| `MAX_CAMERAS` | `50` |
| `SHARD_MODE` | `auto` |
| `EVENT_QUEUE_BACKEND` | `memory` |
| `CONTABO_S3_*` | Credenciais S3 para eventos/clips |

### Métricas típicas

CPU alta (~100–300%) é esperado — inferência YOLO em várias câmeras.

---

## 3. confvision-dvr (gravação contínua)

Sincroniza câmeras com plano **contínuo**, lê segmentos gravados pelo MediaMTX e envia ao S3.

### Configuração

- **Source:** GitHub → `adrianobraz/ConfVision` / `main`
- **Comando:** `python`
- **Arguments:** `-u dvr_main.py`
- **Replicas:** 1

### Mount (Bind)

| Host | Container |
|---|---|
| `/opt/confvision/recordings` | `/recordings` |

Deve ser o **mesmo volume** onde o MediaMTX grava.

### Variáveis de ambiente

| Variável | Valor (foxpro) |
|---|---|
| `XANO_BASE_URL` | `https://xpcy-oyme-lno7.b2.xano.io/api:AC7rgWwW` |
| `MEDIAMTX_API_BASE` | `http://foxpro_confvision:9997` |
| `MEDIAMTX_RTSP_BASE` | `rtsp://foxpro_confvision:8554` |
| `DVR_RECORD_DIR` | `/recordings` |
| `DVR_SYNC_INTERVAL_SEC` | `30` |
| `WORKER_ID` | `worker-docker-01` |
| `DVR_WORKER_VERSION` | `0.2.0` |
| `MEDIAMTX_API_USER` | `dvr` |
| `MEDIAMTX_API_PASS` | *(mesma senha do MediaMTX)* |
| `DVR_MTX_SYNC_API` | `true` |

> Credenciais S3 de gravação vêm do Xano por franqueado (API `vis_gravacao_storage_credenciais_by_franqueado`).
> Não é necessário `CONTABO_S3_*` no env do DVR.

---

## 4. confvision-motion (gravação por movimento)

Grava clipes quando detecta movimento (MOG2). Plano **movimento** — não timelapse.

### Configuração

- **Source:** GitHub → `adrianobraz/ConfVision` / `main`
- **Comando:** `python`
- **Arguments:** `-u motion_main.py`

### Variáveis de ambiente

| Variável | Valor (foxpro) |
|---|---|
| `XANO_BASE_URL` | `https://xpcy-oyme-lno7.b2.xano.io/api:AC7rgWwW` |
| `MEDIAMTX_RTSP_BASE` | `rtsp://foxpro_confvision:8554` |
| `WORKER_ID` | `worker-docker-01` |
| `MOTION_WORKER_VERSION` | `0.1.0` |
| `MOTION_SYNC_INTERVAL_SEC` | `30` |
| `MOTION_RECORD_DIR` | `/tmp/confvision/motion` |
| `MOTION_CLIP_MAX_SEC` | `300` |
| `MOTION_POST_ROLL_SEC` | `5` |
| `MOTION_FRAME_SKIP` | `3` |
| `MOTION_MIN_AREA` | `1500` |
| `MAX_CAMERAS` | `50` |

---

## 5. confvision-timelapse (Timelapse Inteligente)

Alterna automaticamente entre dois modos:

- **Sem movimento:** captura 1 frame a cada `TIMELAPSE_FRAME_INTERVALO_SEG` segundos, acumula frames e monta MP4 timelapse → S3
- **Com movimento:** grava vídeo normal via ffmpeg → S3
- **Movimento para:** volta ao modo timelapse

Referência funcional: documento interno *Timelapse Inteligente com IA* (ConfVision).

### Configuração

- **Source:** GitHub → `adrianobraz/ConfVision` / `main`
- **Comando:** `python`
- **Arguments:** `-u timelapse_main.py`
- **Replicas:** 1
- **Mount:** não obrigatório (grava em `/tmp/confvision/motion` internamente e envia ao S3)

### Variáveis de ambiente

Copiar as do **confvision-motion**, mais as duas do timelapse:

| Variável | Descrição |
|---|---|
| `TIMELAPSE_FRAME_INTERVALO_SEG` | Segundos entre cada foto (modo parado) |
| `TIMELAPSE_FRAMES_POR_SEGMENTO` | Fotos acumuladas antes de montar/enviar 1 vídeo |

#### Perfis recomendados

| Perfil | `INTERVALO_SEG` | `FRAMES` | Tempo real | Vídeo gerado |
|---|---|---|---|---|
| **Teste** (EasyPanel) | `20` | `24` | 8 min | 24 s |
| **Produção** (sugerido) | `720` | `60` | 12 h | 1 min |
| **Produção** (máx. storage) | `720` | `120` | 24 h | 2 min |
| **1 h → 3 min** | `20` | `180` | 1 h | 3 min |

Fórmula: `tempo_real = FRAMES × INTERVALO_SEG` | `vídeo = FRAMES` segundos (1 fps).

Exemplo **teste** no Environment:

```env
TIMELAPSE_FRAME_INTERVALO_SEG=20
TIMELAPSE_FRAMES_POR_SEGMENTO=24
```

Exemplo **produção**:

```env
TIMELAPSE_FRAME_INTERVALO_SEG=720
TIMELAPSE_FRAMES_POR_SEGMENTO=60
```

Demais variáveis compartilhadas com motion: `MOTION_SYNC_INTERVAL_SEC`, `MOTION_RECORD_DIR`, `MOTION_CLIP_MAX_SEC`, `MOTION_POST_ROLL_SEC`, `MOTION_FRAME_SKIP`, `MOTION_MIN_AREA`, etc.

### Logs esperados após deploy

```
[TIMELAPSE] START ConfVision timelapse worker
[TIMELAPSE] OK | xano=... | rtsp=rtsp://foxpro_confvision:8554 ...
[TIMELAPSE] SYNC 1 camera(s) timelapse ids=[123]
[TIMELAPSE] camera=123 stream OK
[TIMELAPSE] camera=123 frame 1/5 capturado
```

Se aparecer `SYNC 0 camera(s)`: verificar licença timelapse ativa na câmera e storage Contabo configurado no franqueado.

---

## 6. confvision-rtmp-guard (Auth + Ban + Falhas RTMP)

Substitui o `rtmp-watch`. Faz:

1. **Auth HTTP** do MediaMTX (`POST /auth`) — publish só com **chave 24** no path
2. **Auto-ban** de IP por taxa de falha (auth negada ou EOF no log)
3. **Desban / ban manual** (`POST /unban`, `POST /ban`)
4. **Lista de falhas** para a UI (`GET /falhas`)
5. **Câmeras online** (`GET /online`) — paths prontos + câmera/franqueado/IP

### Chave RTMP (Hashids)

```text
Hashids(vis_camera.id)
salt = RTMP_PUBLISH_SECRET
alphabet = 0-9a-z
min_length = 12
```

URL no aparelho (sem `?pass=` / `?user=`, **sem** `/live/`):

```text
rtmp://rtmp.dnsid.com.br:1935/{hash12}
```

- **WIFI:** sem `/` no fim  
- **DVR:** com `/` no campo do aparelho (ele remove ao enviar)  
- Guard: decode → `vis_camera` existe + `ativo=true` + `bloqueado=false`

HLS/ao vivo usam o mesmo path: `…/{hash12}/index.m3u8`

### Liberar o guard (ordem segura)

1. Garantir `RTMP_PUBLISH_SECRET` **igual** no guard e na app Go (e nos workers que leem RTSP)
2. Redeploy / start `confvision-rtmp-guard` (instalar `hashids` via requirements)
3. MediaMTX (`mediamtx.yml`):
   - `authMethod: http`
   - `authHTTPAddress: http://foxpro_confvision-rtmp-guard:8100/auth`
   - `authHTTPExclude: [read, playback]`
   - `paths: all_others:`
4. Redeploy app Go ConfVision
5. No painel, copiar a **nova** URL Hashids para cada câmera (corta chave 24 antiga)
6. Log do guard: `OK publish … path={hash12}`

### Configuração EasyPanel

- **Source:** GitHub → `adrianobraz/ConfVision` / `main`
- **Comando:** `python`
- **Arguments:** `-u rtmp_guard_main.py`
- **Replicas:** 1
- **Porta:** `8100` — expor **somente** para o IP do servidor da app Go (firewall)

### Mount (Bind) — mesmo volume do MediaMTX

| Host | Container |
|---|---|
| `/opt/confvision/recordings` | `/recordings` |

### Variáveis de ambiente

| Variável | Valor | Descrição |
|---|---|---|
| `XANO_BASE_URL` | `https://…/api:AC7rgWwW` | Grupo confVision |
| `RTMP_PUBLISH_SECRET` | *(segredo longo)* | **Igual** ao da app Go |
| `RTMP_GUARD_ADMIN_KEY` | *(chave)* | Header `X-RTMP-Guard-Key` |
| `RTMP_GUARD_HTTP_PORT` | `8100` | Porta HTTP |
| `MTX_LOG_FILE` | `/recordings/mediamtx.log` | Log MediaMTX |
| `RTMP_WATCH_JSON` | `/recordings/rtmp_falhas.json` | Persistência falhas |
| `RTMP_BAN_JSON` | `/recordings/rtmp_bans.json` | Persistência bans |
| `RTMP_BAN_MAX_FAILS` | `20` | Falhas para auto-ban |
| `RTMP_BAN_WINDOW_SEC` | `60` | Janela de contagem |
| `RTMP_BAN_TTL_SEC` | `3600` | Duração do ban (1h) |
| `MEDIAMTX_API_BASE` | `http://foxpro_confvision:9997` | API Control (lista online) |
| `MEDIAMTX_API_USER` | `dvr` | Auth da API MediaMTX |
| `MEDIAMTX_API_PASS` | *(senha)* | Igual ao worker DVR |
| `RTMP_ALLOW_READ_OPEN` | `1` | HLS/read sem token (fase 1) |

### MediaMTX

Montar `mediamtx/mediamtx.yml` (já com):

```yaml
authMethod: http
authHTTPAddress: http://foxpro_confvision-rtmp-guard:8100/auth
```

### App Go ConfVision (outro servidor — `.env`)

```env
RTMP_GUARD_URL=http://IP_DA_VPS:8100
RTMP_WATCH_URL=http://IP_DA_VPS:8100
RTMP_GUARD_ADMIN_KEY=igual-ao-do-guard
RTMP_PUBLISH_SECRET=igual-ao-do-guard
```

### Endpoints

| Método | Path | Descrição |
|---|---|---|
| POST | `/auth` | MediaMTX auth webhook |
| GET | `/falhas` | Falhas RTMP |
| GET | `/bans` | IPs banidos |
| POST | `/ban` | Ban manual `{ip, motivo?, ttl_sec?}` |
| POST | `/unban` | Desban `{ip}` |
| GET | `/health` | Healthcheck |

Proxy na app: `/api/rtmp-falhas`, `/api/rtmp-bans`, `/api/rtmp-bans/unban`, `/api/cameras/{id}/rtmp-publish`, `/api/cameras/{id}/bloquear`.

### Xano (push necessário)

- Coluna `vis_camera.bloqueado` (bool)
- `GET vis_camera/rtmp_auth/{id}`
- `POST vis_camera/bloquear/{id}`

### Checklist

- [ ] Push table/APIs Xano (`bloqueado`, rtmp_auth, bloquear)
- [ ] Serviço `confvision-rtmp-guard` no ar; `GET /health` ok
- [ ] MediaMTX com `authMethod: http` apontando ao guard
- [ ] Mesmo `RTMP_PUBLISH_SECRET` no guard e na app Go
- [ ] Porta 8100 liberada só para o IP da app Go
- [ ] Câmeras com URL nova (`{hash12}`, sem `/live/`, sem query)
- [ ] Tela `/rtmp-falhas` lista falhas + Desbanir

---

## Diagrama de arquitetura

```
Câmera IP / DVR
    │ RTMP  rtmp://…:1935/{hash12}
    ▼
confvision (MediaMTX)  ← auth HTTP → confvision-rtmp-guard:8100/auth
    │ log → /recordings/mediamtx.log
    │
    ├── confvision-rtmp-guard (auth + ban + falhas)
    ├── confvision-worker / dvr / motion / …
    └── App Go ConfVision (outro host)  RTMP_GUARD_URL → /rtmp-falhas
```

---

## 6b. confvision-rtmp-watch (legado)

Use apenas se ainda não migrou para o guard. Entry: `python -u rtmp_watch_main.py`, porta `8099`.
O guard já inclui `/falhas` — preferir o guard.

---

## Diagrama legado (workers)

```
Câmera IP / DVR
    │ RTMP publish  rtmp://…:1935/{hash12}
    ▼
confvision (MediaMTX)  ← foxpro_confvision:8554 / :9997 / :1935
    │ log → /recordings/mediamtx.log
    │
    ├── confvision-worker     → main.py              (detecção YOLO / eventos)
    ├── confvision-dvr        → dvr_main.py          (gravação contínua)
    ├── confvision-motion     → motion_main.py       (gravação por movimento)
    ├── confvision-timelapse  → timelapse_main.py    (timelapse inteligente)
    └── confvision-rtmp-watch → rtmp_watch_main.py   (falhas RTMP → :8099)
            │
            ▼
        App Go ConfVision  RTMP_WATCH_URL → /rtmp-falhas
            │
            ▼
        Xano API / Contabo S3
```

---

## Scripts de deploy (build local)

Todos usam a mesma imagem Docker `confvision-worker:TAG`; só muda o **Arguments** no EasyPanel.

| Script | Serviço EasyPanel |
|---|---|
| `./deploy-vps.sh` | `confvision-worker` (`main.py`) |
| `./deploy-dvr-vps.sh` | `confvision-dvr` |
| `./deploy-motion-vps.sh` | `confvision-motion` |
| `./deploy-timelapse-vps.sh` | `confvision-timelapse` |
| `./deploy-rtmp-watch-vps.sh` | `confvision-rtmp-watch` |

Build manual:

```bash
cd confvision
docker build -t confvision-worker:1 .
```

---

## Checklist pós-deploy timelapse

- [ ] Serviço `confvision-timelapse` criado e em execução
- [ ] Logs com `[TIMELAPSE] SYNC N` onde N > 0
- [ ] Câmera com licença `gravacao_timelapse_7d/15d/30d` ativa
- [ ] Storage Contabo ativo no franqueado (tela Gravações)
- [ ] Câmera **não** está em plano motion ou contínuo simultaneamente
- [ ] Para teste: `TIMELAPSE_FRAME_INTERVALO_SEG=20` e `TIMELAPSE_FRAMES_POR_SEGMENTO=24`
- [ ] Produção: `TIMELAPSE_FRAME_INTERVALO_SEG=720` e `TIMELAPSE_FRAMES_POR_SEGMENTO=60`

---

## Referências no código

| Arquivo | Descrição |
|---|---|
| `main.py` | Worker detecção de pessoas |
| `dvr_main.py` | Worker DVR (gravação contínua) |
| `motion_main.py` | Worker gravação por movimento |
| `timelapse_main.py` | Entry point timelapse |
| `timelapse_worker.py` | Lógica timelapse ↔ movimento |
| `rtmp_guard_main.py` | Entry point auth + ban + falhas RTMP |
| `rtmp_guard.py` | Regras auth (Hashids + Xano ativo/bloqueado) |
| `rtmp_ban.py` | Store de IPs banidos |
| `rtmp_token.py` | Hashids publish path |
| `rtmp_watch_main.py` | Entry point monitor falhas RTMP (legado) |
| `rtmp_watch.py` | Parser do log MediaMTX + store |
| `rtmp_messages.py` | Mensagens amigáveis (PT) |
| `mediamtx/mediamtx.yml` | Config MediaMTX (auth HTTP + logFile) |
| `.env.example` | Todas as variáveis documentadas |
| `config.py` | Leitura das variáveis de ambiente |
