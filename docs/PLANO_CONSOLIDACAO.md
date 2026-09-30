# Plano de Consolidação do ConfVision

**Etapa:** 2 — análise e planejamento (sem execução Git/destructive).  
**Data:** 2026-09-30  
**Referências:** [AUDITORIA_ARVORES_GIT.md](./AUDITORIA_ARVORES_GIT.md), [ESTRUTURA_ECOSISTEMA.md](./ESTRUTURA_ECOSISTEMA.md).

---

## 1. Estado atual

| Árvore | Branch | HEAD | Tracking | Ahead/Behind upstream | WIP (status) |
|--------|--------|------|----------|------------------------|--------------|
| `core4` | `main` | `bc22286` | `origin/fix/confvision-rtmp-postgres-nal` | **0 / 0** vs tracking | ~62 linhas (4 modified + muitos untracked) |
| `core4-rust-pilot` | `rust-pilot` | `a2caf47` | `origin/rust-pilot` | **0 / 0** | ~101 (75 modified + 26 untracked) |
| `core4-fase0-push` | `chore/fase0-apply-ops` | `264c8c2` | `origin/chore/fase0-apply-ops` | **0 / 0** | limpo |
| `core4-push-wt` | `confvision/fase0-scripts` | `57e7554` | `origin/confvision/fase0-scripts` | **0 / 0** | limpo |
| `ConfVision` (clone) | `rust-pilot` | `7282bc7` | `origin/rust-pilot` | **0 / 0** vs `origin/rust-pilot` | ~61 modified |

**Remotes:** todos `https://github.com/adrianobraz/ConfVision.git`.

**Divergência local `main` vs `origin/main`:** `main` está **18 commits à frente** e **62 commits atrás** de `origin/main` (histórico local de produção/ecossistema não alinhado ao remoto `main`).

**Divergência remota `origin/main` vs `origin/rust-pilot`:** **1 / 81** commits (eixos de produto distintos no GitHub).

---

## 2. Árvores

Um repositório object store (`core4\.git`) com **4 worktrees** + **1 clone** independente (`ConfVision\.git`). Layout:

| Layout | Onde | Conteúdo |
|--------|------|----------|
| **Ecossistema `main`** | `core4\` | `home/confmonit/v4.0/`, `apis/`, `tables/`, Python em **`confvision/`** |
| **Layout `rust-pilot`** | `core4-rust-pilot\` | Python na **raiz**, `confvision-rust-processor/`, subset `home/...` no Git |
| **Fase 0 branches** | worktrees mínimos | ~94–99 paths no HEAD (Python raiz + scripts) |
| **Clone legado** | `ConfVision\` | layout `rust-pilot`, sem `home/` |

---

## 3. Branches — linha do tempo

### Tips atuais (paralelos recentes — D5 / deploy)

| Branch | Tip | Parent imediato | Tema |
|--------|-----|-----------------|------|
| `main` | `bc22286` | `8e303c1` | D5 multi-host, deploy Go, escala |
| `rust-pilot` | `a2caf47` | `6c807a7` | tenant-stack, D5 `servidor_id` Go (outro commit) |

Os tips **não são fast-forward** um do outro: **`git merge-base main rust-pilot` retorna vazio (exit 1)** neste repositório — o Git **não reporta um merge-base único** entre as refs `main` e `rust-pilot`, apesar de compartilharem histórico antigo (ex.: `Initial commit` na linha `rust-pilot`).

### Contagem de commits exclusivos (`rev-list --left-right`)

| Comparação | Só esquerda | Só direita |
|------------|-------------|------------|
| `main` … `rust-pilot` | **18** | **142** |
| `main` … `chore/fase0-apply-ops` | **18** | **63** |
| `main` … `confvision/fase0-scripts` | **18** | **62** |
| `chore/fase0-apply-ops` … `confvision/fase0-scripts` | **1** | **0** |
| `7282bc7` … `a2caf47` (piloto vs clone) | **0** | **47** |

### Fase 0 no histórico de `main`

`main` já contém commits equivalentes em tema:

- `8c9af69` — chore(fase0) apply-ops  
- `339b14c` — scripts Fase 0 verify/diagnose  

Branches dedicadas têm tips **diferentes** (`264c8c2`, `57e7554`) — provável **recommit/rebase** do mesmo trabalho no layout `rust-pilot`.

---

## 4. Commits

| Ref | Mensagem resumida |
|-----|-------------------|
| `bc22286` | feat(confvision): D5 multi-host, deploy Go… |
| `a2caf47` | docs(deploy): tenant-stack escala… D5 servidor_id Go |
| `264c8c2` | chore(fase0): apply-ops Postgres… |
| `57e7554` | chore(confvision): scripts Fase 0 verify… |
| `7282bc7` | fix(docker): FFmpeg libs rust-pilot 502 |

**Diff estatístico `main` vs `rust-pilot` (Git):** ~**4113** paths tocados (grande parte **renomeação** `confvision/*` ↔ raiz + remoção massiva de paths `home/` na visão `rust-pilot`).

**Diff focado** (`confvision-rust-processor/` + `home/.../confvision/`): **682** arquivos, +25k / −59k linhas.

---

## 5. WIP

### `core4` (`main`) — preservar

- **Go (untracked):** dezenas de `.go` em `visdata/`, `confvision/`, scripts migrations/coleta, `apifunction/`, recursos JS/HTML.
- **Docs (untracked):** `confvision/docs/*` (consolidação, plano Rust, índices Fase 5/6).
- **Docs repo (untracked):** `docs/AUDITORIA_ARVORES_GIT.md`, `ESTRUTURA_ECOSISTEMA.md`, este plano.
- **Modified tracked:** `motion_worker.py`, `paginas.go`, binário `confvision` (Go build).
- **Artefatos (untracked):** `confvision-rust-processor/target/` (~1,6 GB), `tmp/`, `*.exe`, `home/confmonit.rar`.

### `core4-rust-pilot` — preservar

- **Modified (75):** Rust core, SQL, scripts ops, `cameras.go`, `motion_worker.py`, tenant-stack docs, etc.
- **Untracked (26):** **FASE-1…FASE-6**, ADR `ARCHITECTURE_DECISIONS.md`, Fase 5/6 satellite docs, módulos **`control_plane/`**, **`events/`** (catalog, dedup, engine), **`media/storage.rs`**, motion shadow/recording/session, **`sensor/`**, **`timelapse/`**, `motion_shadow_compare.py`.

### `ConfVision` clone

- **61 modified**, concentrados em **`decode/`**, **`capacity/`**, **`load/`**, tenant-stack, Docker profiling — **sobre base `7282bc7`**, não sobre `a2caf47` + WIP do piloto.

---

## 6. Conteúdo exclusivo

| Origem | Exclusivo |
|--------|-----------|
| **main (committed + WIP)** | Ecossistema Go completo, Xano, layout `confvision/`, D5/RTMP/coleta commits recentes |
| **rust-pilot (committed + WIP)** | Histórico Rust completo, layout raiz, deploy/tenant-stack, **Fases 1–6 WIP untracked** |
| **Fase 0 branches** | Delta fino vs `main` em scripts (`fase0-diagnose.ps1`, pacote `fase0-apply-ops` — ver diff) |
| **ConfVision clone** | WIP **decode/GPU/profiling** possivelmente **não refletido** no piloto atual |

**Somente no `main` (commits recentes, exemplos):** `bc22286`, `8e303c1`, `e106de1`, `be3e72b`, `01c530d`, integração coleta/404 UI.

**Somente no `rust-pilot` (142 commits):** cadeia Rust phase 1 → Fase D/C, sidecar YOLO, políticas stream, etc. (desde ~`ffde1a6` Rust phase 1).

---

## 7. Conflitos potenciais

| Componente | Main | Rust pilot | Fase 0 | Clone | Risco | Motivo |
|------------|------|------------|--------|-------|-------|--------|
| Layout paths (`confvision/` vs raiz) | subpasta | raiz | raiz | raiz | **CRÍTICO** | merge Git massivo; deploy/EasyPanel paths |
| `main` vs `rust-pilot` merge-base | — | — | — | — | **CRÍTICO** | `merge-base` falha; tips paralelos D5 |
| `confvision-rust-processor/` | quase vazio no `main` | completo | — | versão antiga | **ALTO** | introdução + WIP |
| Python workers | `confvision/*.py` | `*.py` raiz | raiz | raiz | **ALTO** | renomeação + diffs funcionais (`motion_worker.py` WIP em ambos) |
| Go `visdata/cameras.go` | committed | modified (D5 assign) | — | — | **MÉDIO** | mesmo arquivo, deltas pequenos |
| Go não rastreado em `main` | muito WIP | — | — | — | **ALTO** | perda se merge antes de commit/stash |
| Docs Fase 1–6 | índices em `core4/confvision/docs` | corpo em piloto untracked | — | — | **ALTO** | só disco |
| MediaMTX / Docker | ambos | ambos | — | divergente | **MÉDIO** | Dockerfiles modificados no clone |
| PostgreSQL SQL | `home/.../sql/migrations` | `rust-processor/sql` + WIP | scripts ops | — | **ALTO** | duplicatas/overlap stream_policy |
| `origin/main` vs local `main` | 62 behind | — | — | — | **ALTO** | push/merge sem reconciliar remoto |
| Decode/profiling WIP | — | parcial | — | extenso | **CRÍTICO** | clone vs piloto — reconciliar manual |

---

## 8. Fases 1–6

| Fase | Arquivos principais | Branch / local | No commit? | Untracked? | Destino proposto |
|------|---------------------|----------------|------------|------------|------------------|
| **1** | `FASE-1-RUST-CORE.md`, core Rust `src/` | `rust-pilot` | base sim; extensões WIP | docs **??** | `confvision-rust-processor/docs/` + código em `main` após merge |
| **2** | `FASE-2-UNIFICACAO-VIDEO.md`, motion/pipeline | piloto | parcial | docs **??**, motion modules **??** | idem |
| **3** | `FASE-3-…`, `events/*` | piloto | parcial | **??** engine/catalog/dedup | idem |
| **4** | `FASE-4-…`, `control_plane/` | piloto | parcial | **??** | idem |
| **5** | `FASE-5-…`, `STORAGE_*`, `media/storage.rs` | piloto | parcial | **??** | idem + índice em `core4/confvision/docs/` |
| **6** | `FASE-6-…`, runbooks capacity/DR | piloto | parcial | **??** | idem |

**Consolidação markdown (untracked em `main`):** `AUDITORIA_CONSOLIDACAO_6_FASES.md`, `ARCHITECTURE_FINAL.md`, `PLANO_MIGRACAO_RUST.md` — **preservar** e referenciar paths finais pós-layout.

**Testes Rust:** WIP reportado anteriormente **136** testes OK no piloto — revalidar após integração.

---

## 9. Fase 0

| Branch | Tip | vs `main` | Integração |
|--------|-----|-----------|------------|
| `chore/fase0-apply-ops` | `264c8c2` | 63 commits só na branch (layout); conteúdo **em grande parte já cherry-picked** em `main` (`8c9af69`) | **REVISAR** diff em `home/.../scripts/fase0-*` — exclusivo detectado: `fase0-diagnose.ps1` vs main |
| `confvision/fase0-scripts` | `57e7554` | 62 commits só na branch | **INTEGRAR** somente deltas vs `264c8c2` (7 paths exclusivos apply-ops) |

**Commits exclusivos entre Fase 0 branches:** apply-ops inclui `fase0-apply-ops/main.go`, testes `imagem_token_test.go`, `integracao_dispatch*.go`.

**Proposta:** comparar arquivo a arquivo `main` vs `264c8c2` em `home/confmonit/v4.0/confvision/scripts/` antes de merge; **não** re-mergear layout raiz inteiro.

---

## 10. PostgreSQL

| Artefato | Onde | Classificação |
|----------|------|---------------|
| `20260326_vis_camera_stream_policy.sql` | `main` `home/.../sql/migrations` + piloto `rust-processor/sql/migrations` | **Duplicado** — alinhar versão aplicada |
| `20260326_vis_camera_stream_error_diag.sql` | piloto (modified WIP) + possível espelho `home/...` | **Revisar** — piloto WIP |
| `20260928_vis_coleta_operacional.sql` | `main` tree | **main** — coleta operacional |
| `coleta_operacional_migration.sql` | visdata (path local) | **main / WIP** |
| `fase0_pilot_ops.sql`, `piloto_fase_a_isolamento.sql`, phase C/D SQL | piloto `rust-processor/sql/` | **rust-pilot / ops** — catalogar antes de apply |
| Scripts apply-ops Go | `main` tracked + branch Fase 0 | **main + delta Fase 0** |

**Não aplicado em banco nesta etapa.** Classificar em ambiente real quais scripts já rodaram na VPS.

---

## 11. Estratégia proposta

### Base recomendada: **`main` em `core4`**

**Justificativa:**

1. Único lugar com **ecossistema Go + Xano + layout produção** (`confvision/` sob `core4`).
2. Commits recentes de **produção** (RTMP/Postgres, coleta, 404 UI, D5 dispatch) estão no tip **`main`**, não no tip **`rust-pilot`**.
3. **`rust-pilot`** permanece como **fonte de importação** (Rust, ops, Python histórico), não como branch de deploy final sem re-mapear paths.

**Não usar `rust-pilot` como base** — perde-se o layout `home/` e o estado ecossistema.

### Conteúdo a preservar (por origem)

| Origem | Preservar |
|--------|-----------|
| **main** | Todo WIP Go untracked, docs `confvision/docs`, binários **fora do Git** |
| **rust-pilot** | Todo WIP (modified + untracked Fases 1–6), `cargo test` baseline |
| **Fase 0** | Deltas de scripts/testes não presentes em `main` |
| **ConfVision clone** | Diff WIP `decode/` vs piloto **antes** de descartar clone |

### Conteúdo a integrar

1. **`confvision-rust-processor/`** completo → permanecer em **`core4/confvision-rust-processor/`** ou **`core4/confvision/rust-processor/`** (decisão de path — **BLOQUEIO**: ver critérios).
2. Python: alinhar **`rust-pilot/*.py`** → **`core4/confvision/`** (não raiz).
3. Commits **`rust-pilot`** relevantes → cherry-pick ou merge orientado por **subdiretórios**, não merge bruto 4113 paths.
4. Docs Fase 1–6 untracked → commit no processor docs.
5. Fase 0: patches pontuais em `home/.../scripts/`.

### Permanecer separado (temporário)

- Worktrees Fase 0 até merge confirmado.
- XanoScript (`apis/`, `tables/`) — mesmo repo, domínio diferente; merges ConfVision não devem misturar sem revisão.
- **`receptor-4`**, **`dialyze`** — fora do repo ConfVision.

### Descartável (decisão posterior, **não apagar agora**)

- `target/` (~6,4 GB piloto, ~1,6 GB main untracked)
- `home/confmonit.rar`, `*.exe` test build
- Clone `ConfVision` **após** export WIP decode
- Worktrees vazios pós-`git worktree remove`

---

## 12. Ordem de execução (futura — **não executar na Etapa 2**)

1. **Backup lógico:** tarball/listagem de WIP (`main` untracked + piloto untracked + clone); opcional `git bundle` / tag anotada.
2. **Congelar baseline:** registrar hashes HEAD + `cargo test` + smoke Go.
3. **Reconciliar `main` com remoto:** decidir `origin/main` vs `origin/fix/confvision-rtmp-postgres-nal` (**62 behind** — **BLOQUEIO** até estratégia definida).
4. **Exportar WIP piloto:** `git add -N` / patch series / branch `wip/rust-phases-1-6-snapshot` **só** com arquivos preserváveis.
5. **Reconciliar clone `ConfVision`:** diff WIP decode vs piloto; incorporar ou arquivar.
6. **Criar branch `consolidate/confvision-main`** a partir de `main` atualizado.
7. **Integrar Fase 0:** diff `main` vs `264c8c2` / `57e7554` — cherry-pick commits de scripts se necessário.
8. **Integrar Rust:** trazer `confvision-rust-processor` + commits por tema (não merge único se `merge-base` continuar falhando — usar **`ort` merge com estratégia `-X`**, ou cherry-pick ranges, ou `git read-tree` por path).
9. **Re-mapear Python:** garantir paridade `confvision/` vs raiz `rust-pilot`.
10. **Mesclar Go:** `cameras.go` e commits D5 de ambos tips (`bc22286` vs `a2caf47`).
11. **`.gitignore`:** root + processor (ver §13).
12. **Validação:** `go test` ConfVision, Python lint/smoke, **`cargo test`**, scripts Fase 0 verify.
13. **PostgreSQL:** plano de apply migrations (sem execução automática).
14. **Documentação:** unificar índices Fase 5/6.
15. **Commit(s) atômicos** + review.
16. **Push** para branch de integração (não `main` direto até CI).
17. **Remover worktrees** somente após merge confirmado.

---

## 13. Riscos

| Risco | Impacto |
|-------|---------|
| Merge automático `main` + `rust-pilot` | Perda/mistura de **4113 paths**, quebra layout deploy |
| Perda WIP untracked | **Irreversível** sem backup |
| Tips D5 paralelos | Duplicação ou regressão D5/assign |
| SQL duplicado stream/coleta | Banco inconsistente se apply errado |
| Clone decode WIP | Perda otimizações GPU/profiling |
| `main` 62 behind `origin/main` | Conflitos em massa ao publicar |
| Artefatos `target/` commitados | Repo gigante |

---

## 14. Pré-requisitos

- [ ] Backup WIP verificado (lista de arquivos + cópia externa).
- [ ] Decisão explícita: tracking branch remota alvo (`main` vs `fix/confvision-rtmp-postgres-nal`).
- [ ] Decisão path canônico do processor (`confvision-rust-processor/` na raiz do repo vs sob `confvision/`).
- [ ] Inventário SQL aplicado na VPS vs arquivos no repo.
- [ ] Reconciliação clone `ConfVision` vs piloto (decode).
- [ ] Estratégia para **`git merge-base` ausente** (consulta git experiente ou merge por subtree/path).

---

## 15. Critérios para iniciar a consolidação (Etapa 3+)

1. Todos os itens **BLOQUEIO** abaixo resolvidos ou aceitos por escrito.
2. Backup WIP concluído.
3. Plano de reconciliação com `origin/main` aprovado.
4. Matriz de serviços (§ abaixo) validada pelo responsável deploy.
5. Janela de teste VPS disponível.

---

## BLOQUEIO — REVISÃO MANUAL NECESSÁRIA

1. **`git merge-base main rust-pilot` falha** — definir se merge será por **path** (subtree), **cherry-pick** serializado, ou **reescrita** de histórico (rebase interativo — alto risco).
2. **`main` local 62 commits atrás de `origin/main`** — integrar remoto antes ou deliberately fork; evitar push cego.
3. **Path canônico pós-consolidação** — raiz vs `confvision/` para Python e Rust (impacto EasyPanel/Docker).
4. **WIP decode no clone `ConfVision`** vs piloto @ `a2caf47` — merge manual obrigatório antes de arquivar clone.
5. **Estado real das migrations PostgreSQL** na VPS — o que já foi aplicado.

---

## Anexo A — Serviços (proposta)

| Serviço | Main (`core4/confvision/`) | Rust pilot (raiz / processor) | Diferença | Estratégia proposta |
|---------|----------------------------|----------------------------------|-----------|---------------------|
| **Rust Processor** | não no commit (WIP `target/` only) | `confvision-rust-processor/` completo + WIP Fases 1–6 | só piloto | **INTEGRAR** em `main`; **PRESERVAR** WIP untracked antes |
| **Worker** | `main.py`, `capture_worker.py`, distributed | `main.py` raiz | path + possíveis deltas | **INTEGRAR** → `confvision/`; diff `main.py` |
| **Motion** | `motion_main.py`, `motion_worker.py` (WIP) | `motion_main.py`, `motion_worker.py` (WIP) | ambos WIP | **REVISAR** diff; unificar em `confvision/` |
| **Sensor** | `sensor_main.py` | `sensor_main.py` | layout | **PRESERVAR** main; port deltas piloto |
| **Sync Agent** | referências; **sem** `sync_agent.py` no tree main Git | `sync_agent.py`, `sync_agent_main.py` | ausente main | **INTEGRAR** do piloto → `confvision/` |
| **Timelapse** | `timelapse_main.py` | idem raiz | layout | **PRESERVAR** + port |
| **DVR** | `dvr_main.py` | idem raiz | layout | **PRESERVAR** + port |

**MediaMTX:** `start_mediamtx_guard.py`, Dockerfiles em ambos; WIP em piloto e clone.

---

## Anexo B — Untracked críticos (piloto)

### A. Preservar (commit futuro)

- `confvision-rust-processor/docs/FASE-1` … `FASE-6` (+ satellite docs)
- `src/control_plane/`, `events/*.rs`, `media/storage.rs`, motion/sensor/timelapse modules
- Scripts ops necessários (`motion_shadow_compare.py`)

### B. Ignorar (proposta `.gitignore`)

- `confvision-rust-processor/target/` (~6,42 GB piloto)
- `confvision-rust-processor/tmp/`, `profile-output/`
- `core4/confvision-rust-processor/target/` (untracked main)
- `*.exe`, `home/confmonit.rar`

### C. Revisão manual

- `.env`, `confvision/.env` (main tree)
- `sql/*_local.csv` (já no processor gitignore)
- Binário `home/.../confvision/confvision`

---

## Anexo C — `.gitignore` (proposta, **não aplicada**)

| Arquivo | Lacuna sugerida |
|---------|-----------------|
| `core4/.gitignore` | `confvision-rust-processor/target/`, `**/target/`, `*.exe`, `*.rar` |
| `confvision-rust-processor/.gitignore` | já tem `/target/` — OK |
| `confvision/.gitignore` | verificar `.env`, `__pycache__` |

---

## Anexo D — Outros diretórios Git (`C:\sistemaconfmonit\`)

| Dir | Tipo | Remote | Branch | HEAD | Finalidade provável |
|-----|------|--------|--------|------|---------------------|
| `ConfVision-github` | clone shallow? | ConfVision.git | `main` | `b083990` | snapshot antigo Python |
| `ConfVision-repo` | clone | ConfVision.git | `main` | `d945789` | snapshot antigo |
| `_tmp_ConfVision` | clone | ConfVision.git | `main` | `37e4a75` | temporário RTMP watcher |

**Conteúdo exclusivo:** não auditado arquivo a arquivo; tratar como **candidatos a arquivo** após consolidar piloto + `main`.

---

## Anexo E — ConfVision clone vs piloto

- **Confirmado:** clone @ **`7282bc7`**, piloto @ **`a2caf47`** — **47 commits** de diferença na linha piloto.
- **Cargo.toml:** SHA256 **diferente** (piloto WIP vs clone committed).
- **WIP clone:** forte em **`decode/`**, **`capacity/`**, Docker profiling — **não** equivalente ao WIP Fases 1–6 do piloto.
- **Conteúdo exclusivo provável:** sim, no working tree do clone — **preservar via diff/patch** antes de desativar clone.

---

*Documento gerado na Etapa 2. Nenhum comando destrutivo ou write Git foi executado.*
