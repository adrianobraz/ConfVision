# Auditoria de consolidação — 6 fases ConfVision

**Data:** 2026-09-30  
**Escopo:** código e docs reais em `c:\sistemaconfmonit\core4` e `c:\sistemaconfmonit\core4-rust-pilot`.  
**Regra:** nenhuma fase foi reexecutada; apenas inventário e conflitos.

---

## ETAPA 1–2 — Git e backup

### Repositórios locais (mesmo remote GitHub)

| Diretório | Branch local | Remote | Situação |
|-----------|--------------|--------|----------|
| `core4-rust-pilot` | `rust-pilot` (= `origin/rust-pilot`) | `https://github.com/adrianobraz/ConfVision.git` | **Grande WIP não commitado** (Fases 1–6 Rust/docs) |
| `core4` | `main` (tracking `origin/fix/confvision-rtmp-postgres-nal`) | mesmo | **WIP separado**: Go/portal untracked + `motion_worker.py` + artefatos |

**Commits publicados vs local:**

- `core4-rust-pilot`: `origin/rust-pilot` @ `a2caf47`; working tree **+14 arquivos modificados** e **dezenas untracked** (docs FASE 1–6, `control_plane/`, `media/storage.rs`, motion shadow, etc.).
- `core4`: último commit local `bc22286` (D5/deploy Go); **não contém** o WIP Rust das 6 fases no working tree do piloto.

**Backup código antes de consolidar (recomendado):**

1. **Não** usar `git reset --hard` / `git clean -fd`.
2. Commit ou stash por repo antes de merge; branch `pre-consolidacao-6-fases` **só protege commits**, não arquivos untracked.
3. Copiar tarball ou `git stash -u` em `core4-rust-pilot` prioridade.
4. **PostgreSQL:** backup **não executado** nesta auditoria (sem `POSTGRES_URL` / ambiente produção).

**Risco crítico:** duas árvores de trabalho no mesmo GitHub sem um único “source of truth” commitado.

---

## ETAPA 4 — Matriz de consolidação

| Fase | Planejado (spec) | Implementado | Parcial | Problema principal | Correção necessária | Status auditoria |
|------|------------------|--------------|---------|-------------------|---------------------|------------------|
| **1** Rust core | Processor por câmera, RTSP, pipeline, health | `confvision-rust-processor` completo em **rust-pilot** | E2E carga produção | Worker Python legado vs piloto Stop | Publicar Rust WIP; manter política worker Stop | **Parcial** — código OK, ops pendente |
| **2** Vídeo unificado | Menos RTSP duplicado | Rust ingest + motion gate; shadow MOG2 | Timelapse/sensor/DVR ainda Python RTSP | Dupla ingestão motion/timelapse **ainda possível** | Shadow medido; não desligar Python sem paridade | **Parcial** |
| **3** Eventos | Fila, dedup, PG | Redis queue, capture Rust, `vis_evento` Go | Idempotência fila | Go não consome Redis | Documentado; opcional outbox futuro | **Parcial** |
| **4** Control Plane | Node agent, lease, desired state | `worker_id`, `vis_worker` ping, D5 assign, pull sync | Node Agent binário, lease, failover | **Sem lease** — split brain | Migration lease **não** inventada; doc risco | **Parcial / conflitante** spec vs código |
| **5** Storage / HA | Media abstrata, HA testada | `MediaStorage`, docs HA; upload tolerante | Chaos tests, purge, orphan job | S3 retry Rust ≠ Python | Testes falha; janitor futuro | **Parcial** (doc > runtime) |
| **6** Scale / ops | Scheduler, benchmarks | D5 scheduler + capacity Rust + coleta PG | Load test, camera weight | Throughput **não medido** | SCALE_TESTING em staging | **Parcial** (doc > medição) |

---

## ETAPA 3 — Detalhe por fase

### FASE 1 — Rust Processor

| Item | Conteúdo |
|------|----------|
| **Status** | **Implementado** (código); **Parcial** (publicação GitHub / produção) |
| **Arquivos** | `core4-rust-pilot/confvision-rust-processor/src/**` (`main.rs`, `camera/`, `rtsp/`, `pipeline/`, `worker/`, `decode/`, `yolo/`) |
| **Documentação** | `confvision-rust-processor/docs/FASE-1-RUST-CORE.md`, ADR-001–006 |
| **Testes** | `cargo test` → **136 passed** (2026-09-30, rust-pilot) |
| **Problemas** | Código das fases **não commitado**; `core4/confvision-rust-processor/` só artefato `target/` |
| **Pendências** | Stress multi-câmera, E2E eventos produção |

### FASE 2 — Pipeline vídeo

| Item | Conteúdo |
|------|----------|
| **Status** | **Parcial** |
| **Arquivos Rust** | `motion/session.rs`, `shadow.rs`, `hooks.rs`, `recording.rs` (stub), `timelapse/mod.rs`, `sensor/mod.rs` stubs |
| **Python** | `core4/confvision/motion_worker.py` (mod local), `core4-rust-pilot/motion_worker.py`; `motion_main.py`, `timelapse_*`, `dvr_main.py` |
| **Documentação** | `FASE-2-UNIFICACAO-VIDEO.md`, ADR-007 |
| **Testes** | Unit Rust; shadow compare script `motion_shadow_compare.py` — comparação operacional pendente |
| **Problemas** | **Duplicidade RTSP:** Rust analítico + motion/timelapse Python podem rodar juntos |
| **Pendências** | Métricas shadow; clip motion no Rust |

### FASE 3 — Eventos

| Item | Conteúdo |
|------|----------|
| **Status** | **Parcial** |
| **Arquivos** | `events/catalog.rs`, `dedup.rs`, `engine.rs`, `queue.rs`, `capture/process.rs`; Go `eventos.go` |
| **Documentação** | `FASE-3-PROCESSAMENTO-E-EVENTOS.md`, ADR-008–009 |
| **Testes** | Unit dedup/queue; integração Redis+PG manual |
| **Problemas** | Cooldown ≠ idempotência global; DLQ manual |
| **Pendências** | Event engine unificado motion |

### FASE 4 — Control Plane

| Item | Conteúdo |
|------|----------|
| **Status** | **Parcial** (spec **Node Agent** ≠ código) |
| **Arquivos Go** | `rust_processor_d5.go`, `workers.go`, `cameras.go`, `mediamtx.go`, `coleta_operacional.go` |
| **Arquivos Rust** | `control_plane/mod.rs`, sync loop `main.rs`, `sharding/` |
| **Python** | `sync_agent_main.py`, `config_cache.py` **só em rust-pilot** — **ausente em `core4/confvision`** |
| **Documentação** | `FASE-4-CONTROL-PLANE.md`, ADR-010–013 |
| **Testes** | Ping/sync manuais; sem teste lease |
| **Problemas** | Doc fala Node Agent; runtime = Rust integrado. **config_cache** não alinhado entre clones |
| **Pendências** | Lease, migration segura, sync-agent no mesmo repo commitado |

### FASE 5 — Storage / HA

| Item | Conteúdo |
|------|----------|
| **Status** | **Parcial** |
| **Arquivos** | `media/storage.rs`, `media/upload.rs`; Python `storage.py`, `gravacao_storage.py`; docs STORAGE/HA |
| **Documentação** | FASE-5 + DEPENDENCY_GRAPH + ADR-014–018 |
| **Testes** | Unit `MediaStorage`; chaos **NÃO EXECUTADO** |
| **Problemas** | Docs afirmam procedimentos não validados em prod |
| **Pendências** | Orphan/missing media job; backup PG validado |

### FASE 6 — Escala / operação

| Item | Conteúdo |
|------|----------|
| **Status** | **Parcial** |
| **Arquivos** | D5 Go, `capacity/*` Rust, `coleta_operacional`; docs RUNBOOK, CAPACITY_PLANNING, SCALE_TESTING |
| **Documentação** | `FASE-6-SCALE-AND-OPERATIONS.md`, ADR-019–024 |
| **Testes** | Load/API/GPU **NÃO EXECUTADO** |
| **Problemas** | Scheduler existe (D5) mas doc fase 4 dizia “não failover” — consistente |
| **Pendências** | Benchmarks; unificar métricas dashboard |

---

## ETAPA 5–7 — Responsabilidade por serviço (conflitos)

| Componente | Responsável hoje | Rust substitui? | Ação consolidação |
|------------|------------------|-----------------|-------------------|
| **confvision-worker** (Python analítico) | **Stop** (piloto) | **Sim** (ingest+YOLO+capture) | Não reativar; código legado preservado no repo |
| **confvision-rust-processor** | Analítico RTSP+YOLO+eventos | N/A | **Fonte canônica** — commit WIP |
| **confvision-motion** | Clips MOG2 + RTSP | Parcial (gate luma Rust) | **Manter** até paridade shadow |
| **confvision-timelapse** | Timelapse/movimento | Stubs Rust apenas | **Manter** |
| **confvision-sensor** | Poll eventos sensor | Stub Rust | **Manter** |
| **confvision-dvr** | MTX record + S3 + `vis_gravacao_segmento` | Não | **Manter** |
| **confvision-sync-agent** | Cache Redis config | Não (Rust usa API) | **Manter** para Python; alinhar arquivos no repo |
| **Control Plane Go** | PG, API, D5, eventos | N/A | Unificar clone `core4` untracked Go com rust-pilot |
| **Node Agent** (spec) | — | — | **Não existe**; Rust + ping = data plane |

**Duplicidade crítica:** dois clones Git + motion Python vs Rust analítico na mesma câmera se ambos ativos.

---

## ETAPA 9–10 — PostgreSQL (código vs banco)

### Migrations versionadas encontradas

| Arquivo | Conteúdo |
|---------|----------|
| `home/confmonit/v4.0/confvision/sql/migrations/20260326_vis_camera_stream_policy.sql` | colunas stream policy `vis_camera` |
| `home/confmonit/v4.0/confvision/sql/migrations/20260928_vis_coleta_operacional.sql` | stream diag + `vis_stream_relatorio`, `vis_sistema_health`, `vis_sistema_metric` |
| `confvision-rust-processor/sql/migrations/20260326_vis_camera_stream_*.sql` | variantes piloto (error_diag) |
| `coleta_operacional_migration.sql` (embed Go) | duplicate lógica coleta |

**Aplicação:** `StartColetaOperacionalBackground` aplica embed na subida Go; scripts `apply-coleta-migration.ps1` em **core4 untracked**.

### Tabela código vs banco (expectativa)

| Recurso | Código espera | Tabela/coluna PG | Lease dedicado |
|---------|---------------|------------------|----------------|
| Camera | `vis_camera`, RTSP, stream_* | Sim (migrations acima) | **Não** |
| Node | `vis_worker`, `vis_mediamtx_node` | Sim | **Não** |
| Assignment | `vis_camera.worker_id` | Sim | **Não** |
| Event | `vis_evento`, `vis_evento_clip` | Sim | N/A |
| Media | URLs + `vis_gravacao_segmento` | Sim | N/A |
| Heartbeat | `vis_worker.ultimo_ping_em` | Sim | N/A |
| Scheduler | D5 env + capacity HTTP | Sem tabela scheduler | N/A |

**Fases 4–6 não exigem novas tabelas** para funcionar como implementado. **Lease/failover** exigiria design **novo** — **não criar** na consolidação sem spec aprovada.

**Banco real:** não inspecionado nesta auditoria (sem conexão). Validar com `apply-coleta-migration` / `pg-audit` no ambiente.

---

## Documentação vs código

| Doc | Alinhado? |
|-----|-----------|
| FASE 1–3 Rust | Sim com rust-pilot WIP |
| FASE 4 Node Agent separado | **Não** — código usa Rust integrado |
| FASE 5 chaos tests | **Não** — marcados NÃO EXECUTADO |
| FASE 6 load tests | **Não** — idem |
| `core4/confvision/docs/PLANO_MIGRACAO_RUST.md` | Referência; pode divergir do WIP |

---

## Índices de docs (localização)

| Path | Conteúdo |
|------|----------|
| `core4-rust-pilot/confvision-rust-processor/docs/FASE-*.md` | Fases 1–6 completas (untracked) |
| `core4/confvision/docs/FASE-5/6-*-INDEX.md` | Links (untracked) |
| Este arquivo | Consolidação |

---

## Próximas etapas (após aprovação — **sem commit ainda**)

1. Escolher **repo canônico** (`core4-rust-pilot` recomendado para Rust + home/confmonit).
2. Merge/reconcile `core4` Go untracked → piloto ou vice-versa.
3. `.gitignore`: excluir `target/`, `*.exe`, `home/confmonit.rar`.
4. Migrations: garantir `20260928` aplicada em prod; **sem DROP**.
5. Testes Go + integração staging.
6. Commits organizados + push (branch `rust-pilot` ou PR → `main`).
7. `DEPLOY_ATUALIZACAO.md`, `CHECKLIST_DEPLOY.md`, `RELATORIO_CONSOLIDACAO.md` pós-correção.

**Ver também:** [ARCHITECTURE_FINAL.md](./ARCHITECTURE_FINAL.md)
