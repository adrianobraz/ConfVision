# Validação RTMP-GUARD

**Última execução:** 2026-09-30 (fase validação + teste real)  
**Branch de trabalho:** `consolidate/confvision-phase7` (não usar `backup/antes-validacao-rtmp-guard` para editar)  
**HEAD:** `f982038`  
**Checkpoint (somente backup):** `backup/antes-validacao-rtmp-guard`  
**Commit automático:** não realizado (aguardando autorização)

---

## Estado inicial

| Item | Valor |
|------|--------|
| WIP local | **Preservado** (painel Go + docs escala fora do escopo Guard) |
| Proibidos executados | Nenhum (`reset`, `clean`, `revert`, checkout de commit antigo) |
| Recuperação Git | Consulta `-S` / diff vs `f982038`; **nenhum arquivo/função copiado do histórico** |

### Alterações locais Guard (`git diff --stat`)

```text
 confvision/Dockerfile-mediamtx   |  3 ++-
 confvision/Dockerfile.mediamtx   |  6 +++++-
 confvision/mediamtx/mediamtx.yml |  2 ++
 confvision/rtmp_guard_main.py    | 32 ++++++++++++++++++++++++++-----
 confvision/rtmp_messages.py      | 24 +++++++++++++++++++++++
 confvision/rtmp_watch.py         | 41 ++++++++++++++++++++++++++++++++++++++++
 6 files changed, 101 insertions(+), 7 deletions(-)
```

### Untracked (Guard)

- `confvision/log_rotate.py`
- `confvision/docs/RECONSTRUCAO_RTMP_GUARD.md`
- `confvision/docs/VALIDACAO_RTMP_GUARD.md`
- `confvision/tests/test_validacao_rtmp_guard.py`

---

## Branch e commit

```text
Branch: consolidate/confvision-phase7
HEAD:   f982038 feat(rtmp/mediamtx): negar publish na pausa sistema e reduzir auto-ban
```

`start_mediamtx_guard.py` e núcleo `rtmp_guard.py` / `rtmp_ban.py` em **f982038** permanecem a base; incrementos locais são aditivos.

---

## Arquivos analisados

| Caminho | Função |
|---------|--------|
| `rtmp_guard.py` | Auth, cache, pausa sistema, path `cam/*` |
| `rtmp_guard_main.py` | HTTP :8100, `/auth`, `/ban`, `/unban`, tail log |
| `rtmp_ban.py` | BanStore, auto-ban, TTL, `unban()` |
| `rtmp_watch.py` | Parser log, falhas, DTS/track (local) |
| `rtmp_messages.py` | Motivos UX |
| `rtmp_token.py` | Hashids camera_id |
| `log_rotate.py` | Rotação copytruncate (local) |
| `start_mediamtx_guard.py` | Boot Guard → wait → MediaMTX |
| `mediamtx/mediamtx.yml` | `authHTTPAddress` → Guard |
| `Dockerfile.mediamtx`, `Dockerfile-mediamtx` | Imagem combinada |
| `requirements-guard.txt` | requests, hashids |
| `tests/test_*.py` | 3 módulos de teste |
| Go (painel) | `rtmp_auth.go`, `rtmp_watch.go`, proxies `/api/rtmp-*` |

---

## Arquivos alterados (vs f982038)

Ver stat acima. **Nenhum arquivo rastreado foi removido ou substituído por versão antiga.**

---

## Funções recuperadas

```text
NENHUMA
```

Pesquisa Git (`-S"maybe_rotate_log"`, `-S"listar_online"`, `unban` em `rtmp_ban.py`): implementações já presentes em `f982038` ou **novas** apenas no working tree (log rotate, rate-limit log, `_handle_stream_diag`).

### Registro de recuperações

*(vazio — fase não exigiu cópia histórica)*

---

## Origem histórica

| Componente | Commit referência | Uso nesta fase |
|------------|-------------------|----------------|
| Guard core | `f982038` | Base intacta |
| Boot wait | `4c761c1` lineage | Já em `start_mediamtx_guard.py` @ f982038 |
| Dockerfile combinado | `5e5a322` lineage | Já em tree |

---

## Adaptações

Incrementos locais adaptados à arquitetura atual (Go `rtmp_auth`, env EasyPanel, volume `/recordings`):

- Rotação de log sem reiniciar MediaMTX (copytruncate).
- Rate-limit **somente** de linhas de log para `ip_banido`.
- Classificação watch para HLS/RTMP diagnóstico (sem alterar timestamps).

---

## Ambiente

| Recurso | Disponível |
|---------|------------|
| Host Windows dev | Sim |
| Python (`python` / `py`) | **Não** (Store alias / `py` ausente) |
| Docker CLI | Sim (29.8.0) |
| Docker daemon | **Não** (`dockerDesktopLinuxEngine` indisponível) |
| Tenant RTMP E2E | **Não** configurado nesta sessão |
| HLS player test | **Não** |

---

## Python

```text
STATUS: SKIP
Motivo: Python runtime não disponível no host (python/py/python3 não executáveis)
```

Comando preparado (CI / container com daemon ativo):

```bash
cd confvision
pip install -r requirements-guard.txt
python -m compileall -q .
python -m unittest discover -s tests -p "test_*.py" -q
```

**Suítes esperadas:** `test_rtmp_messages` (3), `test_rtmp_guard_stream_policy` (3), `test_validacao_rtmp_guard` (9) → **~15 testes**.

---

## Docker

```text
STATUS: SKIP (daemon)
Motivo: failed to connect to docker API at npipe:////./pipe/dockerDesktopLinuxEngine
```

---

## Build

```text
DOCKER BUILD: SKIP
Imagem alvo: confvision-rtmp-guard:validacao-local
Dockerfile: confvision/Dockerfile.mediamtx
```

Revisão estática Dockerfile: inclui `log_rotate.py`, env `RTMP_LOG_*`, `RTMP_BAN_DENY_LOG_INTERVAL_SEC` → **OK no papel**.

---

## Startup

Análise de `start_mediamtx_guard.py`:

1. `Popen(rtmp_guard_main.py)`
2. `_wait_guard_http` → URL `http://127.0.0.1:8100/health`
3. Log `[START] Guard HTTP pronto`
4. `Popen(mediamtx)`

```text
STATUS: PASS (análise estática)
Execução container: SKIP (sem Docker daemon)
```

---

## Authentication

- Fluxo: MediaMTX POST `/auth` → `RtmpGuard.authorize()` → GET `{XANO_BASE_URL}/vis_camera/rtmp_auth/{id}`.
- Testes unitários em repo (`test_rtmp_guard_stream_policy.py`) cobrem pausa sistema / inativa sem ban.

```text
STATUS: PASS COM RESSALVAS (código + testes no repo; execução SKIP)
E2E staging: SKIP
```

---

## IP Ban

- Checagem **antes** de path (`authorize` L83–84).
- Resposta HTTP = `code` em `do_POST` **antes** de rate-limit de log.

```text
STATUS: PASS COM RESSALVAS (revisão + testes escritos; E2E SKIP)
```

---

## Unban

- `BanStore.unban()` em `rtmp_ban.py`.
- `POST /unban` em `rtmp_guard_main.py`.
- Painel: `ProxyRtmpBansFranqueado`, `/api/rtmp-bans/unban` (Go).

```text
STATUS: PASS COM RESSALVAS (implementação existe; E2E ban→unban→publish SKIP)
```

---

## Rate Limit

- `_should_log_ban_deny` — afeta **apenas** `print` após `send_response(code)`.
- Env: `RTMP_BAN_DENY_LOG_INTERVAL_SEC` (default 60).

```text
STATUS: PASS (revisão de código + testes em test_validacao_rtmp_guard.py — execução SKIP)
```

---

## Log Rotation

- `maybe_rotate_log`: limite `RTMP_LOG_MAX_BYTES`, retenção `RTMP_LOG_KEEP_ROTATED`, copytruncate.
- Invocado no loop `follow_file` (~120s).

```text
STATUS: PASS COM RESSALVAS (lógica + testes unitários escritos; execução SKIP)
```

---

## Reconnect

```text
STATUS: SKIP — sem ambiente RTMP
```

Sem alteração de backoff/retry (regra da fase).

---

## RTSP

MediaMTX expõe `:8554`; validação funcional depende de publish RTMP OK.

```text
STATUS: SKIP
```

---

## HLS

`:8888`, `hlsAlwaysRemux: yes` em `mediamtx.yml`.

```text
STATUS: SKIP — sem player/tenant nesta sessão
```

---

## DTS

| Campo | Valor |
|-------|--------|
| Detecção | `rtmp_watch.feed()` → `hls_dts_nao_monotono` |
| Origem típica | Encode RTMP / mux HLS MediaMTX |
| Impacto | HLS instável; RTSP/Rust podem seguir parcialmente |
| Correção artificial | **Não aplicada** |

```text
DTS WATCH: PASS COM RESSALVAS (parser; execução unittest SKIP)
```

---

## Video Track

| Campo | Valor |
|-------|--------|
| Detecção | `rtmp_video_track_nao_configurado`, `rtmp_so_audio_sem_video` |
| Origem típica | RTMP reconnect ou só-áudio |
| Temporário vs persistente | **Requer logs produção** — não classificado E2E aqui |

```text
TRACK WATCH: PASS COM RESSALVAS (parser; execução unittest SKIP)
```

---

## Tabela funcional (§2)

| Funcionalidade | Estado atual | Histórico (f982038) | Situação |
|----------------|--------------|---------------------|----------|
| Authentication | Código + testes repo | Igual base | EXISTE; E2E SKIP |
| IP Ban | Completo | Igual | EXISTE; E2E SKIP |
| IP Unban | `/unban` + Go proxy | Igual | EXISTE; E2E SKIP |
| Watch | + DTS/track local | Base | EXISTE PARCIALMENTE VALIDADO |
| Log Rotation | `log_rotate.py` novo | Ausente | EXISTE; testes SKIP |
| RTMP Publish | `cam/{hash12}` | Igual | EXISTE; E2E SKIP |
| Camera ID | Hashids + API | Igual | EXISTE |
| RTSP | MediaMTX config | Igual | NÃO EXECUTADO |
| HLS | MediaMTX config | Igual | SKIP |
| DTS | Watch local | Ausente | NOVO (diagnóstico) |
| Video Track | Watch local | Ausente | NOVO (diagnóstico) |
| Reconnect | — | — | SKIP |

---

## IPS-BANIDOS (html/js)

| Pergunta | Resposta |
|----------|----------|
| Usados pelo Guard Python? | **Não** — Guard expõe `GET /bans`; UI usa **Go** |
| Referências no código? | **Sim** — `paginas.go`, `rotas.go`, menu, `urls.js`, `dashboard.js` |
| UI oficial franqueado? | **Sim** — rota `/ips-banidos`, API `/api/rtmp-bans/franqueado` |
| Duplicados? | Cópia em `confvision-20260930T181656Z-1-001/` (backup local) |
| Completos? | html + js presentes em `recursos/modulos/confvision/` |
| Versionados no Git? | **`git ls-files *ips-banidos*` → vazio** |

**Recomendação:** **NECESSÁRIO VERSIONAR** `ips-banidos.html` e `ips-banidos.js` no repositório Go (`home/confmonit/v4.0/confvision/`) antes de deploy limpo — **não versionado automaticamente nesta fase**.

---

## Problemas encontrados

1. Docker daemon parado → compileall, unittest, build **SKIP**.
2. Python ausente no host dev.
3. UI IPs banidos operacionalmente necessária mas **untracked** no Git.
4. IP banido ainda avaliado antes de motivos cadastro (gap conhecido; fora escopo correção).
5. Retry RTMP permanece no **equipamento**, não no Guard.

---

## Pendências

1. Iniciar Docker Desktop → rodar unittest + `docker build -f Dockerfile.mediamtx`.
2. Smoke tenant: `[START] Guard HTTP pronto`, 1 câmera RTMP→RTSP.
3. E2E ban / unban / reconnect.
4. Versionar `ips-banidos.*` (com autorização).
5. Commit Guard + docs + tests (com autorização).

---

## Tratamento automático de falhas de câmera

### Erro detectado

- Linha MediaMTX (tail `MTX_LOG_FILE`): `received a packet for video track 0, but track is not set up` (ou equivalente com `video track` + `not set up`).
- Parser: `rtmp_watch.RtmpLogParser.feed` → código local `rtmp_video_track_nao_configurado` (`classificar_motivo` / `_handle_stream_diag`).
- **Não** dispara pausa automática: `hls_dts_nao_monotono` (DTS) — apenas diagnóstico local nesta fase.

### Regra de decisão

1. Primeira ocorrência deduplicada no `RtmpWatchStore` (`vezes == 1`).
2. Path `cam/{hash12}` → `camera_id` via `parse_chave_rtmp`.
3. POST `CONFVISION_API_URL/vis_worker_ping` com `camera_stream_health[]`:
   - `event`: `pause_analytic`
   - `pause_reason`: `sistema_stream_video_track_not_set_up`
   - `error_class`: `VIDEO_TRACK_NOT_SET_UP`
   - `last_error`: trecho do log (até 500 chars).
4. Auth: `VIS_WORKER_API_KEY` (header `Authorization: Bearer` + `X-Vis-Worker-Key`).
5. Após pausa OK: `GUARD.cache.invalidate(camera_id)` para negar publish com `stream_pausado_sistema` sem cache stale.

Implementação: `stream_health_client.py`, hook em `rtmp_guard_main._WatchWithBan._handle_stream_diag`.

### Estado aplicado (Postgres / painel)

Reutiliza mecanismo existente `ApplyCameraStreamHealthBatch` → `pause_analytic`:

| Campo | Valor |
|-------|--------|
| `analitico_pausado` | `true` |
| `stream_motivo_pausa` | `sistema_stream_video_track_not_set_up` |
| `stream_erro_classe` | `VIDEO_TRACK_NOT_SET_UP` |
| `stream_ultimo_erro` | trecho do log |

Publish RTMP: `rtmp_guard._stream_pausado_pelo_sistema` → HTTP **403** `stream_pausado_sistema` (sem auto-ban de IP).

### Evento gerado

`notifyCameraStreamPaused` (Go): `vis_stream_relatorio` (dedupe 30 min por `referencia_dedupe`) + `vis_evento` tipo `sistema_stream` (dedupe 24 h por câmera).

Título/dica específicos para `sistema_stream_video_track_not_set_up` em `stream_health.go`.

### Mensagem no painel

`armado.js` → `labelMotivoPausaSistema`: texto para video track; badge **Stream com problema** quando `stream_motivo_pausa` começa com `sistema_stream`.

### Recuperação

Usuário corrige encoder/DVR → **Reativar stream** (`POST /api/cameras/{id}/stream/reativar` → `ReactivateCameraStream`: limpa motivo `sistema_stream%`, incrementa `stream_policy_generation`).

Sem retry automático agressivo após pausa.

### Deduplicação

| Camada | Mecanismo |
|--------|-----------|
| Log watch local | `RtmpWatchStore` (mesmo IP/código/path, mesmo minuto) |
| POST pause Guard → Go | `RTMP_STREAM_PAUSE_DEDUPE_SEC` (default 3600 s) por `camera_id` |
| Relatório / evento | Dedupe Postgres em `notifyCameraStreamPaused` |

### Comportamento do IP banido

1. `RtmpGuard.authorize`: `BanStore.is_banned(ip)` **antes** de Xano, path ou cache → **403** `ip_banido`.
2. Sem sessão MediaMTX autorizada → sem publish/HLS/RTSP downstream no ConfVision.
3. Log HTTP: bloco rate-limited `IP BANIDO` / `MOTIVO: ip_banido` / `AÇÃO: conexão rejeitada` (`RTMP_BAN_DENY_LOG_INTERVAL_SEC`, default 60 s). **403 sempre**; throttle só no print.
4. DVR pode repetir TCP/RTMP; Guard continua 403 — firmware não controlável.

Variáveis Guard (stream health): `CONFVISION_API_URL`, `VIS_WORKER_API_KEY`, opcional `RTMP_STREAM_PAUSE_DEDUPE_SEC`, `RTMP_GUARD_WORKER_ID`.

---

## Riscos

- Deploy só de `f982038` sem untracked → sem log rotate/rate-limit/diag novos.
- Deploy Go sem ips-banidos versionados → 404 página franqueado.
- Volume `/recordings/rtmp_bans.json` desincronizado de expectativa operador.

---

## Próxima fase

1. CI: pytest em container.
2. E2E mínimo 1 câmera + unban.
3. Opcional: ordem auth ban vs `stream_pausado_sistema` (correção mínima, fase dedicada).
4. Ownership/lease/D5 — **fora de escopo**.

---

## Alterações de código feitas durante fases Guard (não commitadas)

| Arquivo | Natureza |
|---------|----------|
| `log_rotate.py` | Novo |
| `rtmp_guard_main.py` | Rate-limit log |
| `rtmp_watch.py` | DTS/track + rotate hook |
| `rtmp_messages.py` | Motivos |
| `Dockerfile.mediamtx`, `Dockerfile-mediamtx` | COPY + env |
| `mediamtx/mediamtx.yml` | Comentários |
| `tests/test_validacao_rtmp_guard.py` | Novo |
| `docs/RECONSTRUCAO_RTMP_GUARD.md`, este doc | Documentação |

**Preservação:** funções e comportamento de `f982038` mantidos; nenhum módulo substituído por versão antiga.

---

## Execução dos testes bloqueados (2026-09-30)

Fase dedicada a rodar compileall/unittest/Docker **sem commit** e **sem alterar código** (salvo este relatório).

### 1. Ambiente confirmado

| Check | Resultado |
|-------|-----------|
| `git branch --show-current` | `consolidate/confvision-phase7` |
| `git log -1 --oneline` | `f982038` |
| WIP confvision/ | Preservado (6 modified + 4 untracked Guard) |
| `com.docker.service` (Windows) | **Running** |
| Docker Engine API (`docker info` Server) | **FAIL** — `dockerDesktopLinuxEngine` pipe ausente após ~3 min |
| Ação tentada | Iniciar `Docker Desktop.exe` (sem alterar repo) — engine **não** ficou operacional |

**Decisão (regra da fase):** daemon parado → **PARAR** testes Docker; não modificar o projeto para “consertar” Docker.

### 2. Python no host

| Comando | Resultado |
|---------|-----------|
| `python --version` | Não encontrado (Store alias) |
| `py --version` | Comando inexistente |
| `python3 --version` | Não encontrado |

→ **SKIP** (não instalar Python no host nesta fase).

### TESTE EXECUTADO vs ANÁLISE ESTÁTICA

| Item | Tipo | Resultado |
|------|------|-----------|
| compileall | TESTE EXECUTADO | **SKIP** (sem Python/Docker) |
| unittest discover | TESTE EXECUTADO | **SKIP** |
| pip install requirements-guard | TESTE EXECUTADO | **SKIP** |
| docker build Dockerfile.mediamtx | TESTE EXECUTADO | **SKIP** (daemon) |
| container startup `[START] Guard HTTP pronto` | TESTE EXECUTADO | **SKIP** |
| Auth / IP ban / unban E2E HTTP | TESTE EXECUTADO | **SKIP** |
| log_rotate / DTS / track (unittest) | TESTE EXECUTADO | **SKIP** |
| Startup ordem Guard→MTX | ANÁLISE ESTÁTICA | PASS (código) |
| Rate limit só log | ANÁLISE ESTÁTICA | PASS (código) |
| test_rtmp_guard_stream_policy (repo) | ANÁLISE ESTÁTICA | Presente; não reexecutado |

### Comandos prontos (quando Docker Engine estiver OK)

```powershell
cd c:\sistemaconfmonit\core4\confvision
docker build -f Dockerfile.mediamtx -t confvision-rtmp-guard:validacao-local .
docker run --rm -v "${PWD}:/app" -w /app python:3.11-slim bash -c "pip install -q -r requirements-guard.txt && python -m compileall -q . && python -m unittest discover -s tests -p 'test_*.py' -v"
docker run --rm -p 8100:8100 -p 1935:1935 -e RTMP_PUBLISH_SECRET=test -e XANO_BASE_URL=http://127.0.0.1:9 confvision-rtmp-guard:validacao-local
# (startup smoke — exige env real para auth completa)
```

### Correções nesta fase

- **Código Guard:** nenhuma.
- **Relatório:** esta seção adicionada.
