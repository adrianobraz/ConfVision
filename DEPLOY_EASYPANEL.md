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
| `confvision-rtmp-watch` | Monitor de falhas RTMP (log → UI) | GitHub `ConfVision` | `python -u rtmp_watch_main.py` |

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
| `MTX_AUTHMETHOD` | `internal` | Autenticação interna |
| `MTX_AUTHINTERNALUSERS_0_USER` | `dvr` | Usuário API/DVR |
| `MTX_AUTHINTERNALUSERS_0_PASS` | *(senha forte)* | Senha — trocar em produção |
| `MTX_AUTHINTERNALUSERS_0_PERMISSIONS` | `api,read,playback` | Permissões |
| `MTX_AUTHINTERNALUSERS_1_USER` | `any` | Publicador genérico |
| `MTX_AUTHINTERNALUSERS_1_PERMISSIONS` | `publish,read,playback` | Permissões |

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

## 6. confvision-rtmp-watch (Falhas RTMP)

Lê o log do MediaMTX (`/recordings/mediamtx.log`), detecta conexões que abrem e fecham sem publicar, e expõe uma API HTTP interna. A app Go do ConfVision faz proxy e mostra a tela **Configurar → Falhas RTMP** (`/rtmp-falhas`).

### Configuração

- **Source:** GitHub → `adrianobraz/ConfVision` / `main` (ou imagem `confvision-worker:TAG`)
- **Comando:** `python`
- **Arguments:** `-u rtmp_watch_main.py`
- **Replicas:** 1
- **Porta interna:** `8099` (não precisa expor na internet)

### Mount (Bind) — obrigatório o mesmo volume do MediaMTX

| Host | Container |
|---|---|
| `/opt/confvision/recordings` | `/recordings` |

### Variáveis de ambiente

| Variável | Valor |
|---|---|
| `MTX_LOG_FILE` | `/recordings/mediamtx.log` |
| `RTMP_WATCH_JSON` | `/recordings/rtmp_falhas.json` |
| `RTMP_WATCH_HTTP_PORT` | `8099` |
| `RTMP_WATCH_MAX` | `500` |
| `RTMP_WATCH_DEDUPE_SEC` | `60` |

### App Go ConfVision (`.env`)

No serviço web ConfVision (não no worker):

```env
RTMP_WATCH_URL=http://foxpro_confvision-rtmp-watch:8099
```

Nome do host = projeto EasyPanel + nome do serviço (`foxpro` + `confvision-rtmp-watch`).

### Endpoints internos do watch

| Método | Path | Descrição |
|---|---|---|
| GET | `/health` | Status |
| GET | `/falhas?limit=100` | Lista de falhas (PT) |
| GET | `/resumo` | Totais / IPs únicos |

Proxy na app: `/api/rtmp-falhas` e `/api/rtmp-falhas/resumo`.

### Logs esperados

```
[RTMP-WATCH] START | log=/recordings/mediamtx.log | json=/recordings/rtmp_falhas.json | http=:8099
[RTMP-WATCH] HTTP em 0.0.0.0:8099  GET /falhas /resumo /health
[RTMP-WATCH] FALHA ip=187.x.x.x path=live/2 codigo=eof ...
```

### Checklist

- [ ] MediaMTX com `logFile: /recordings/mediamtx.log` e volume `/recordings`
- [ ] Serviço `confvision-rtmp-watch` no ar; `GET /health` ok
- [ ] Arquivo `/recordings/mediamtx.log` existe e cresce
- [ ] `RTMP_WATCH_URL` no `.env` da app Go
- [ ] Tela `/rtmp-falhas` lista falhas após tentativa ruim de publish

---

## Diagrama de arquitetura

```
Câmera IP / DVR
    │ RTMP publish  rtmp://…:1935/live/{id}
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
| `rtmp_watch_main.py` | Entry point monitor falhas RTMP |
| `rtmp_watch.py` | Parser do log MediaMTX + store |
| `rtmp_messages.py` | Mensagens amigáveis (PT) |
| `mediamtx/mediamtx.yml` | Config MediaMTX (inclui logFile) |
| `.env.example` | Todas as variáveis documentadas |
| `config.py` | Leitura das variáveis de ambiente |
