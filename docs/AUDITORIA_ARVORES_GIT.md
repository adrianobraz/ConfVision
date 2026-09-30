# Auditoria de árvores, clones e worktrees ConfVision

**Data da auditoria:** 2026-09-30  
**Escopo:** `C:\sistemaconfmonit\` — cinco diretórios solicitados.  
**Regras:** somente investigação; nenhuma alteração Git, movimentação ou limpeza.

---

## 1. Diretórios encontrados

| Diretório | Existe | Tamanho aprox. (disco) | `.git` | Notas |
|-----------|--------|------------------------|--------|--------|
| `core4` | Sim | ~3,0 GB | Diretório (`C:\sistemaconfmonit\core4\.git`) | Worktree **principal** do repo compartilhado |
| `core4-fase0-push` | Sim | ~0 GB (checkout mínimo) | Arquivo → `core4\.git\worktrees\...` | Worktree vinculado |
| `core4-push-wt` | Sim | ~0 GB (checkout mínimo) | Arquivo (worktree) | Worktree vinculado |
| `core4-rust-pilot` | Sim | ~6,4 GB | Arquivo (worktree) | Worktree; volume grande (`target/`, builds) |
| `ConfVision` | Sim | ~3,5 GB | **Diretório `.git` próprio** | **Clone independente** (não listado em `git worktree list` de `core4`) |

**Outros diretórios Git no mesmo pai (fora do escopo formal, decisão manual):** `ConfVision-github`, `ConfVision-repo`, `_tmp_ConfVision` — também contêm `motion_main.py` / layout `rust-pilot`; não auditados neste relatório.

---

## 2. Git de cada diretório

Remote comum: **`https://github.com/adrianobraz/ConfVision.git`**

| Diretório | `--show-toplevel` | `--git-common-dir` | Branch atual | HEAD | Tracking |
|-----------|-------------------|--------------------|--------------|------|----------|
| `core4` | `core4` | `core4/.git` | `main` | `bc22286` | `origin/fix/confvision-rtmp-postgres-nal` |
| `core4-fase0-push` | `core4-fase0-push` | `core4/.git` | `chore/fase0-apply-ops` | `264c8c2` | `origin/chore/fase0-apply-ops` |
| `core4-push-wt` | `core4-push-wt` | `core4/.git` | `confvision/fase0-scripts` | `57e7554` | `origin/confvision/fase0-scripts` |
| `core4-rust-pilot` | `core4-rust-pilot` | `core4/.git` | `rust-pilot` | `a2caf47` | `origin/rust-pilot` |
| `ConfVision` | `ConfVision` | `ConfVision/.git` | `rust-pilot` | `7282bc7` | `origin/rust-pilot` |

**Commit anterior (HEAD~1) — comparação de branches:**

| Árvore | HEAD | HEAD~1 |
|--------|------|--------|
| `core4` (`main`) | `bc22286` | `8e303c1` |
| `core4-rust-pilot` (`rust-pilot`) | `a2caf47` | `6c807a7` |
| `ConfVision` (`rust-pilot`) | `7282bc7` | `4f104a0` |
| `core4-fase0-push` | `264c8c2` | (branch dedicada) |
| `core4-push-wt` | `57e7554` | (branch dedicada) |

`main` (`bc22286`) e `rust-pilot` no worktree piloto (`a2caf47`) compartilham timestamp de commit (2026-09-30 01:43 -0300) mas são **hashes diferentes** — branches divergentes no mesmo repositório lógico.

---

## 3. Worktrees identificados

Saída de `git -C core4 worktree list`:

```text
C:/sistemaconfmonit/core4            bc22286 [main]
C:/sistemaconfmonit/core4-fase0-push 264c8c2 [chore/fase0-apply-ops]
C:/sistemaconfmonit/core4-push-wt    57e7554 [confvision/fase0-scripts]
C:/sistemaconfmonit/core4-rust-pilot a2caf47 [rust-pilot]
```

**Total:** 4 worktrees (incluindo o diretório principal `core4`) ligados a **um** object store: `core4\.git` (pasta `worktrees/` presente).

---

## 4. Clones identificados

| Clone | Object store | Relação com `core4` |
|-------|--------------|---------------------|
| `C:\sistemaconfmonit\ConfVision` | `ConfVision\.git` | Mesmo remote GitHub; **não** é worktree de `core4` |
| (menção) `ConfVision-github`, `ConfVision-repo`, `_tmp_ConfVision` | próprios | Revisão manual |

---

## 5. Projetos independentes (no sentido de árvore Git)

- **Repositório lógico A:** `core4` + 3 worktrees (`fase0-push`, `push-wt`, `rust-pilot`).
- **Repositório lógico B:** `ConfVision` (clone completo separado).

O **ecossistema ConfMonit** (pastas `home/confmonit/v4.0/`, XanoScript em `apis/`, etc.) existe de forma **completa no worktree `main` (`core4`)**; nos worktrees de branch `rust-pilot` / Fase 0 o layout no disco segue o **layout flat** do repo GitHub (Python na raiz), não o layout `main` com `core4\confvision\`.

---

## 6. Branches

Branches locais visíveis a partir de qualquer worktree do repo `core4` (exemplo):

- `main` → checkout em `core4`
- `rust-pilot` → checkout em `core4-rust-pilot`
- `chore/fase0-apply-ops` → checkout em `core4-fase0-push`
- `confvision/fase0-scripts` → checkout em `core4-push-wt`
- `feat/stream-404-pause-ui` (sem worktree dedicado no list)

No clone `ConfVision`: branches `main` (local `128dcbe`, **ahead 1 / behind 4** vs `origin/main`) e `rust-pilot` (checkout atual, **atrás** do piloto worktree).

---

## 7. Remotes

Todos apontam para **`origin` → `https://github.com/adrianobraz/ConfVision.git`**.

---

## 8. Commits (último por árvore)

| Árvore | Commit | Mensagem (resumo) |
|--------|--------|-------------------|
| `core4` | `bc22286` | feat(confvision): D5 multi-host, deploy Go… |
| `core4-rust-pilot` | `a2caf47` | docs(deploy): tenant-stack escala… |
| `core4-fase0-push` | `264c8c2` | chore(fase0): apply-ops Postgres… |
| `core4-push-wt` | `57e7554` | chore(confvision): scripts Fase 0 verify… |
| `ConfVision` | `7282bc7` | fix(docker): runtime FFmpeg… rust-pilot 502 |

---

## 9. WIP não commitado

| Árvore | Linhas `git status --short` (aprox.) | Destaques |
|--------|--------------------------------------|-----------|
| `core4` | ~62 | Go WIP em `home/.../confvision/`; docs em `confvision/docs/` untracked; `AGENTS.md` modificado; artefatos `confvision-rust-processor/target/` untracked no `main` |
| `core4-rust-pilot` | ~101 | Rust (`Cargo.toml`, `src/`, Dockerfiles); **docs Fase 1–6** em `confvision-rust-processor/docs/` majoritariamente **untracked** (`??`) |
| `ConfVision` | ~61 | Rust/docker WIP; diverge do piloto |
| `core4-fase0-push` | limpo | — |
| `core4-push-wt` | limpo | — |

---

## 10. Arquivos exclusivos (evidência)

- **Docs consolidação Fases 1–6 (Markdown de auditoria/arquitetura):** presentes em **`core4\confvision\docs\`** (`AUDITORIA_CONSOLIDACAO_6_FASES.md`, `ARCHITECTURE_FINAL.md`, índices Fase 5/6) — **não** no HEAD commitado do piloto; **FASE-1…FASE-6** detalhados só no disco do **`core4-rust-pilot`** (untracked).
- **Branch `chore/fase0-apply-ops`:** commit inclui `home/confmonit/v4.0/confvision/scripts/fase0-apply-ops/` e testes Go associados; parte já também rastreada em `main` (ex.: `fase0-apply-ops/main.go` em `git ls-files` no `core4`).
- **Ecossistema Go completo:** `home/confmonit/v4.0/confvision/` — **`core4` (`main`)**; ausente no checkout físico dos worktrees Fase 0 (sparse/mínimo).
- **Clone `ConfVision`:** branch `rust-pilot` **desatualizada** vs worktree piloto; `Cargo.toml` **hash diferente** do piloto (WIP no piloto).

---

## 11. Componentes duplicados

Matriz de presença no disco (serviços EasyPanel / processamento):

| Componente | core4 | fase0-push | push-wt | rust-pilot | ConfVision |
|------------|:-----:|:----------:|:-------:|:----------:|:----------:|
| **Rust Processor** (`confvision-rust-processor/`) | Não (só untracked `target/`) | Não | Não | **Sim (fonte)** | Sim (versão antiga) |
| **Motion** | `confvision/motion_main.py` | `motion_main.py` (raiz) | idem | idem | idem |
| **Sensor** | `confvision/` | raiz | raiz | raiz | raiz |
| **Sync Agent** | só `easypanel/sync-agent.env` | raiz `sync_agent*.py` | idem | idem | idem |
| **Timelapse / DVR** | `confvision/*_main.py` | raiz | raiz | raiz | raiz |
| **Worker analítico** | `confvision/main.py` | `main.py` raiz | idem | idem | idem |
| **Docs Fase 1–6** (`FASE-*.md` processor) | Índices em `confvision/docs/` | Não | Não | **Sim (untracked)** | Não |
| **Go app ConfVision** | `home/.../confvision/` | Não (checkout) | Não | subset no Git | Não no disco típico |

**Duplicação real:** não é “cópia byte a byte” entre `core4\confvision\` e `rust-pilot\motion_main.py` — são **layouts de branch diferentes** (`main` vs `rust-pilot`). Entre **`core4-rust-pilot`** e **`ConfVision`**: mesmo layout de branch, mas **commits e WIP diferentes** (`a2caf47`+ vs `7282bc7`+).

---

## 12. Relações entre as árvores

```text
C:\sistemaconfmonit\
│
├── core4\                          ← git principal (.git) + worktree [main]
│   ├── apis/, tables/, …           ← XanoScript / ecossistema workspace
│   ├── confvision\                 ← Python vídeo (layout branch main)
│   └── home\confmonit\v4.0\
│       └── confvision\             ← Go app ConfVision (canônico produção)
│
├── core4-fase0-push\               ← worktree [chore/fase0-apply-ops]
├── core4-push-wt\                  ← worktree [confvision/fase0-scripts]
├── core4-rust-pilot\               ← worktree [rust-pilot] + WIP Fases 1–6
│   └── confvision-rust-processor\
│
└── ConfVision\                     ← clone Git separado (.git próprio)
    └── confvision-rust-processor\  (atrás do piloto)
```

**Datas de criação de pasta (Windows):** `ConfVision` ~2026-08-03; `core4-rust-pilot` ~2026-09-26; `core4-push-wt` ~2026-09-28; `core4-fase0-push` ~2026-09-29.

---

## 13. Diretório canônico por função

| Função | Canônico recomendado |
|--------|----------------------|
| Repositório Git / object store | `C:\sistemaconfmonit\core4\.git` |
| Ecossistema + Xano + branch `main` | `C:\sistemaconfmonit\core4\` |
| ConfVision **aplicação Go** | `core4\home\confmonit\v4.0\confvision\` |
| ConfVision **vídeo Python** (produção layout `main`) | `core4\confvision\` |
| **Rust processor** + WIP Fases 1–6 | `core4-rust-pilot\confvision-rust-processor\` |
| Scripts Fase 0 apply-ops (branch) | worktree `core4-fase0-push` |
| Scripts Fase 0 verify/diagnose (branch) | worktree `core4-push-wt` |
| Clone legado `rust-pilot` | `ConfVision\` — **não** assumir paridade com piloto |

---

## 14. Diretórios que precisam de decisão manual

| Diretório | Motivo |
|-----------|--------|
| `ConfVision` | Clone separado; WIP e commit antigos; possível redundância com `core4-rust-pilot` |
| `ConfVision-github`, `ConfVision-repo`, `_tmp_ConfVision` | Clones adicionais não mapeados nesta auditoria |
| `core4-fase0-push`, `core4-push-wt` | Worktrees pequenos; branches podem ser mergeadas/arquivadas no GitHub — **não remover pasta sem `git worktree remove`** |
| `core4-rust-pilot` | **WIP crítico não commitado** (docs + Rust) |
| `core4` | WIP Go + docs untracked no `main` |

---

## Investigações específicas (Etapas 8–11)

### `core4-fase0-push`

- **Tipo:** worktree Git.
- **Origem/nome:** criado ~2026-09-29; nome sugere branch de push/ops da Fase 0.
- **Branch/commit:** `chore/fase0-apply-ops` @ `264c8c2`.
- **Alterações exclusivas:** conteúdo da branch (apply-ops, testes visdata); working tree limpo.
- **Remoção:** **REVISÃO MANUAL NECESSÁRIA** — só após merge/arquivamento da branch.

### `core4-push-wt`

- **Tipo:** worktree Git (sufixo `wt`).
- **Branch/commit:** `confvision/fase0-scripts` @ `57e7554`.
- **Finalidade:** scripts verify/diagnose/pg-audit Fase 0.
- **Alterações exclusivas:** diferenças entre commits `57e7554` e `264c8c2` (famílias de scripts Fase 0).
- **Remoção:** **REVISÃO MANUAL NECESSÁRIA**.

### `core4-rust-pilot`

- **Tipo:** worktree (não clone).
- **Branch/commit:** `rust-pilot` @ `a2caf47`; tracking `origin/rust-pilot`.
- **WIP:** ~101 entradas de status; docs Fase 1–6 untracked; código Rust modificado.
- **vs `core4`:** milhares de paths diferem entre `main` e `rust-pilot` (layout + histórico).
- **Remoção:** **NÃO** — fonte ativa do processor Rust.

### `ConfVision`

- **Tipo:** **CLONE** independente.
- **Branch/commit:** `rust-pilot` @ `7282bc7` (atrás de `a2caf47` no worktree piloto).
- **Relação:** mesmo produto GitHub; **não** contém ecossistema `home/`; Python na raiz como no layout `rust-pilot`.
- **Remoção:** **REVISÃO MANUAL NECESSÁRIA** — verificar WIP local antes de qualquer arquivamento.

---

## 15. Workspace Cursor

Multi-root habitual (informação de sessão): `core4`, `receptor-4`, `dialyze`.  
Não há `.code-workspace` em `sistemaconfmonit` apontando para os worktrees; regras em `core4\.cursor\rules\` referem-se a Xano, não aos paths dos worktrees.

---

## 16. Separação conceitual

| Conceito | O que é |
|----------|---------|
| **Ecossistema** | `core4` + `home/confmonit/v4.0/*` + XanoScript + outros produtos |
| **ConfVision aplicação** | Go em `home/confmonit/v4.0/confvision/` |
| **ConfVision vídeo / processamento** | Python (`core4/confvision` ou raiz no `rust-pilot`) + **Rust** (`confvision-rust-processor`) |
| **Worktrees / clones** | Mecanismos Git — não são “produtos” |

---

## Referências

- [ESTRUTURA_ECOSISTEMA.md](./ESTRUTURA_ECOSISTEMA.md) — seção «Árvores de Desenvolvimento e Git»
- [../confvision/docs/AUDITORIA_CONSOLIDACAO_6_FASES.md](../confvision/docs/AUDITORIA_CONSOLIDACAO_6_FASES.md)
