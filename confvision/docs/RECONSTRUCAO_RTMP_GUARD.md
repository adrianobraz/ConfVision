# Reconstrução RTMP-GUARD — relatório

**Data:** 2026-09-30  
**Branch:** `consolidate/confvision-phase7`  
**Commit de referência (HEAD remoto):** `f982038` — `feat(rtmp/mediamtx): negar publish na pausa sistema e reduzir auto-ban`  
**Checkpoint local:** branch `backup/antes-reconstrucao-rtmp-guard` (aponta para o mesmo HEAD; **não** descarta WIP local)

---

## 1. Estado encontrado (fonte de verdade = repo atual)

A implementação **não estava ausente do Git** na pasta `core4/confvision/`. O módulo é **flat** (vários `.py` na raiz de `confvision/`), não um diretório `rtmp-guard/` separado.

| Artefato | Caminho | Situação |
|----------|---------|----------|
| Auth HTTP MediaMTX | `rtmp_guard.py`, `POST /auth` em `rtmp_guard_main.py` | **Presente** |
| Ban / unban | `rtmp_ban.py`, `/bans`, `/ban`, `/unban` | **Presente** |
| Boot Guard + MediaMTX | `start_mediamtx_guard.py` | **Presente** |
| Config MediaMTX | `mediamtx/mediamtx.yml` | **Presente** |
| Imagem combinada | `Dockerfile.mediamtx`, `Dockerfile-mediamtx` | **Presente** |
| Watch / falhas | `rtmp_watch.py`, `rtmp_messages.py` | **Presente** |
| Testes | `tests/test_rtmp_guard_stream_policy.py`, `test_rtmp_messages.py` | **Presente** |
| Auth cadastro (Go) | `home/.../visdata/rtmp_auth.go` | **Presente** (WIP local) |
| Proxy painel | `rtmp_watch.go`, `rtmp_franqueado.go`, rotas `/api/rtmp-*` | **Presente** |
| UI IPs banidos | `recursos/.../ips-banidos.html/js` | **No disco**; **não rastreado** no Git (`git ls-files` vazio) |

**Conclusão:** perda provável foi **no servidor** (volume/pasta apagada) ou confusão com layout antigo; no repositório o Guard **já estava reconstruído/consolidado** até `f982038`.

---

## 2. O que foi “perdido” vs histórico Git

Pesquisa `git log -- confvision/rtmp_*` — últimos commits relevantes:

| Commit | Conteúdo |
|--------|----------|
| `f982038` | `stream_pausado_sistema`, `MOTIVOS_NO_BAN`, testes stream policy |
| `4c761c1` | Boot race Guard, bans falsos em auth refused |
| `980a4da` | Deploy path `confvision/`, wait Guard |
| `5cbad98` | Guard HTTP antes MediaMTX |
| `5e5a322` | Dockerfile combinado MediaMTX+Guard |

**Nenhum commit recente apagou** o conjunto `rtmp_guard*.py` do tree `confvision/`.  
**Não foi feito** checkout/reset/revert de branch inteira nesta tarefa.

---

## 3. Tabela comparativa (tarefa §4)

| Componente | Existe atualmente | Existia no Git | Última versão | Situação |
|------------|------------------:|---------------:|---------------|----------|
| RTMP-GUARD (Python) | Sim | Sim | `f982038` | **OK** — evoluído |
| Config MediaMTX | Sim | Sim | `f982038` | **OK** |
| Auth → Go/Xano | Sim | Sim | WIP + `rtmp_auth.go` | **OK** (deploy depende env `XANO_BASE_URL`) |
| IP Ban | Sim | Sim | `f982038` | **OK** |
| API Guard (:8100) | Sim | Sim | `f982038` | **OK** |
| Logs + watch | Sim | Sim | `f982038` | **Melhorado** (rotação + DTS/vídeo) |
| UI ips-banidos | Sim (disco) | Parcial | backup zip local | **Pendente:** `git add` páginas |

---

## 4. O que foi recuperado / reforçado nesta tarefa

Sem substituir o tree por commit antigo. **Diff incremental:**

| Arquivo | Mudança |
|---------|---------|
| `log_rotate.py` | **Novo** — copytruncate `mediamtx.log` (`RTMP_LOG_MAX_BYTES`, `RTMP_LOG_KEEP_ROTATED`) |
| `rtmp_watch.py` | Rotação no tail; parse **DTS** e **video track not set up** / só áudio |
| `rtmp_messages.py` | Motivos amigáveis `hls_dts_nao_monotono`, `rtmp_video_track_*`, `ip_banido` |
| `rtmp_guard_main.py` | Rate-limit log `403 ip_banido` (`RTMP_BAN_DENY_LOG_INTERVAL_SEC`) |
| `Dockerfile.mediamtx` (+ alias) | Env de log/ban log; `COPY log_rotate.py` |
| `mediamtx/mediamtx.yml` | Comentário política de log |

**Preservado:** Rust processor, sharding, stream policy Go, commit `f982038` Guard rules, arquitetura Fase 7.

---

## 5. Versões históricas usadas (somente consulta `git show`)

| Componente | Referência histórica | Uso |
|------------|---------------------|-----|
| Stream policy + no-ban inativa | `f982038` | Já no HEAD — **não recopiado** |
| Boot wait | `4c761c1` / `start_mediamtx_guard.py` atual | Já no HEAD |
| Dockerfile combinado | `5e5a322` lineage | Já no HEAD |

---

## 6. Problemas de log — investigação

### Problema 1 — `403 ip_banido` + retry

| Pergunta | Resposta |
|----------|----------|
| Quem retenta? | **Quase sempre o DVR/NVR/app RTMP** — abre TCP de novo na `:1935`; o Guard **não** controla o cliente. |
| O que o Guard faz? | Responde 403 rápido; ban persiste em `/recordings/rtmp_bans.json` (TTL ~1h). |
| Correção sistema | **Log:** rate-limit 1 linha/min por IP (`RTMP_BAN_DENY_LOG_INTERVAL_SEC`). **Ops:** `/unban` ou painel IPs banidos; corrigir path (`cam/{hash}`). |
| Backoff no Guard? | **Não aplicável** ao protocolo RTMP publish; backoff seria no equipamento. |

### Problema 2 — DTS not monotonically increasing

| Item | Detalhe |
|------|---------|
| Origem | **Publisher RTMP** (timestamps) → MediaMTX mux **HLS** |
| Impacto | Player HLS instável/crash; RTSP para Rust pode continuar parcialmente |
| Frequência | Por câmera/encode — ver `[RTMP-WATCH]` / motivo `hls_dts_nao_monotono` |
| Alteração timestamps | **Não feita** — só classificação/diagnóstico |

### Problema 3 — video track not set up

| Item | Detalhe |
|------|---------|
| Origem | Sessão RTMP com **áudio antes do vídeo** ou reconnect mid-handshake |
| Camada | MediaMTX demux RTMP (não Rust) |
| Ação | Motivo `rtmp_video_track_nao_configurado` no watch; corrigir encode no NVR |

### Problema 4 — Tamanho de logs

| Política | Config |
|----------|--------|
| Arquivo | `/recordings/mediamtx.log` |
| Rotação | `RTMP_LOG_MAX_BYTES=52428800` (50 MiB), `RTMP_LOG_KEEP_ROTATED=3` |
| Falhas JSON | `RTMP_WATCH_JSON` — snapshot deduplicado (max 500 itens config) |
| stdout | EasyPanel — retenção do painel |

---

## 7. Testes

| Teste | Resultado |
|-------|-----------|
| `pytest tests/test_rtmp_*.py` | **Não executado** — Python ausente no host dev Windows |
| Regressão manual recomendada | 1 publish OK; IP ban → 403; unban; path inválido; pausa sistema |
| Carga 10–1000 câmeras | **Pendente** staging (§20 do plano) |

Checklist segurança (staging):

- IP autorizado → publish 200  
- IP banido → 403 `ip_banido`  
- Credencial inválida → 403/401  
- Câmera inexistente → 403  

Integração: Camera → RTMP → Guard → MediaMTX → RTSP → Rust (validar `worker_id` + path `cam/*`).

---

## 8. Deploy / reconstrução no servidor

Se a **pasta/container** foi apagada no EasyPanel:

1. Rebuild imagem: Dockerfile path **`confvision/Dockerfile.mediamtx`**
2. Volume **`/recordings`** (bans, log, falhas JSON)
3. Env: `RTMP_PUBLISH_SECRET`, `XANO_BASE_URL` (API Go), `MEDIAMTX_API_USER/PASS`
4. Portas: 1935, 8554, 8888, 8100, 9997
5. Confirmar log: `[START] Guard HTTP pronto` antes de linhas MediaMTX

---

## 9. Pendências

1. **Versionar** `ips-banidos.html/js` no Git (existem, untracked).  
2. **Deploy** imagem com commit que inclui `log_rotate.py` + rate-limit logs.  
3. **Merge** `rust_processor_d5.go` no Go canônico (fora escopo Guard, mas ops assign).  
4. Testes pytest/ carga em staging.  
5. Evolução opcional: checar `stream_pausado_sistema` **antes** de `ip_banido` para IPs legítimos com ban antigo (mudança de ordem auth — requer validação).

---

## 10. Mapa rápido de arquivos (módulo Guard)

```text
confvision/
  start_mediamtx_guard.py   # entrypoint container
  rtmp_guard_main.py        # HTTP :8100
  rtmp_guard.py             # regras auth
  rtmp_ban.py               # bans JSON
  rtmp_watch.py             # tail log + falhas
  rtmp_messages.py          # UX motivos
  log_rotate.py             # rotação log
  mediamtx/mediamtx.yml
  Dockerfile.mediamtx
```

**Critério de sucesso:** estado atual ConfVision **+** Guard operacional documentado **+** melhorias de log/DTS **sem** rollback de arquitetura — **atendido** no repositório; falta rebuild/deploy no tenant afetado.
