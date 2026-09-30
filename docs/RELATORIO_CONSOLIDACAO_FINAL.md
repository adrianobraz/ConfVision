# Relatório de consolidação final — ConfVision

**Data:** 2026-09-30  
**Branch:** `consolidate/confvision-rust`  
**Base:** `main` @ `bc22286` (backup: `backup/consolidation-preserve-2026-09-30`)

---

## O que foi incorporado

- **`confvision-rust-processor/`** completo a partir de `core4-rust-pilot` (src, docs Fase 1–6, SQL, scripts, Docker, deploy interno).
- **`deploy/tenant-stack/`** (env examples, docs ops, scripts host/tenant).
- **Python:** `sync_agent.py`, `sync_agent_main.py`, `config_cache.py` em `core4/confvision/`.
- **Go:** WIP visdata/confvision (coleta, RTMP, integração, relatório operacional, D5 em `cameras.go`), scripts auxiliares.
- **Documentação:** arquitetura, serviços, deploy, auditorias, plano consolidação.

---

## O que foi preservado

- Worktrees **`core4-rust-pilot`**, **`core4-fase0-push`**, **`core4-push-wt`**, clone **`ConfVision`** — **não removidos**.
- Branch **`backup/consolidation-preserve-2026-09-30`** (ponto antes da consolidação).
- Python legado em **`core4/confvision/`** (motion, dvr, timelapse, sensor, etc.) — **não substituído**.
- Aplicação Go em **`home/confmonit/v4.0/confvision/`** — **mantida**, com WIP commitado.

---

## O que continua legado

| Serviço | Estado |
|---------|--------|
| confvision-motion | **Ativo** (Python) |
| confvision-sensor | **Ativo** |
| confvision-sync-agent | **Ativo** (+ arquivos incorporados) |
| confvision-timelapse | **Ativo** |
| confvision-dvr | **Ativo** |
| confvision-worker | **DESATIVADO** — não reativar com Rust |

Detalhe: `confvision/docs/SERVICOS_PROCESSAMENTO.md`.

---

## O que foi migrado para Rust (código no repo)

Processamento **analítico principal**: RTSP, decode, motion gate, YOLO, eventos, Redis, health, storage, control plane, capacity — em **`confvision-rust-processor/`**.

---

## O que ainda não foi migrado

- Clips motion MOG2 dedicados (**motion** Python).
- Timelapse, sensor terminal, DVR contínuo, sync cache config (**Python**).
- Lease anti-duplicata formal entre nós.
- Merge histórico Git `main` ↔ `rust-pilot` (integração foi **por cópia de árvore**, não merge de branches).

---

## PostgreSQL

Migrations catalogadas em **`docs/DEPLOY_ATUALIZACAO.md`**.  
**Nenhuma migration aplicada nesta consolidação local.** Aplicar no ambiente conforme checklist (**APLICAR NO AMBIENTE**).

---

## Redis

Sem alteração de infraestrutura. Rust e Python continuam usando filas/cache conforme env examples.

---

## MediaMTX

Config Python permanece em `confvision/mediamtx/`; deploy MediaMTX+Guard inalterado em essência.

---

## Serviços

Ver tabela em **`confvision/docs/SERVICOS_PROCESSAMENTO.md`**.  
**Rust** = analítico principal; **worker** = off.

---

## Testes executados

| Teste | Resultado |
|-------|-----------|
| `cargo test` (`confvision-rust-processor`) | **136 passed** |
| `go build` (ConfVision app) | **OK** |
| `go test ./...` (ConfVision app) | **OK** (incl. `visdata`, `confvision`) |

---

## Testes que falharam

- Nenhum após correção `APIFUNCTION_URL` em `config.go`.

---

## Problemas encontrados

1. **`src/apifunction`** dependia de `config.ApiFunctionURL` — **corrigido** (`APIFUNCTION_URL`).
2. Clone **`ConfVision`** WIP decode **não** mergeado (piloto já à frente em `a2caf47`); clone preservado.
3. Histórico Git **`main` vs `rust-pilot`** ainda divergente — consolidação física no `core4`, não merge Git entre branches.

---

## Git

| Item | Valor |
|------|--------|
| Branch criada | `consolidate/confvision-rust` |
| Backup branch | `backup/consolidation-preserve-2026-09-30` |
| Commits | `0ad1bc7` rust processor + deploy · `d4bc1a4` python sync · `5688f6c` Go WIP · `b66b4a3` docs · `e672254` config fix |

---

## Push realizado

Push para **`origin/consolidate/confvision-rust`** (ver saída do comando no ambiente; se falhar por rede/auth, repetir `git push -u origin consolidate/confvision-rust`).

---

## Como atualizar o servidor

**`docs/DEPLOY_ATUALIZACAO.md`**

---

## Como fazer rollback

1. Checkout **`backup/consolidation-preserve-2026-09-30`** ou commit **`bc22286`**.  
2. Redeploy imagens/binários anteriores.  
3. Restaurar envs e PG se migrations tiverem sido aplicadas.  
4. Manter **confvision-worker STOPPED** se já estava.

---

## RTSP / duplicidade

- **Proibido** em operação normal: **Rust + confvision-worker** na mesma câmera.  
- Motion/timelapse/DVR podem consumir MediaMTX/RTSP por **responsabilidades distintas** — revisar shard/worker_id no runbook.

---

## Artefatos não commitados (proposital)

- `confvision-rust-processor/target/`, `tmp/`
- `home/confmonit.rar`, binários `confvision`, `.env`
- `home/confmonit/v4.0/franqueadopromais/docs/` (fora do escopo)
