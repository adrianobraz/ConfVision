# Fase 7 — Consolidação final (Fases 1–6)

**Data:** 2026-09-30  
**Branch Git:** `consolidate/confvision-phase7`  
**Repositório:** `C:\sistemaconfmonit\core4` (ConfVision.git)

---

## 1. Estrutura real (investigação)

### Hipótese das “três árvores”

| Caminho | Hipótese | Veredito |
|---------|----------|----------|
| `core4\home\confmonit\v4.0\confvision\` | Sistema Go / aplicação principal | **CONFIRMADO** — `app.go`, `visdata`, `visapi`, PostgreSQL |
| `core4-rust-pilot\` | Nova implementação Rust | **PARCIAL** — worktree `rust-pilot`; era bancada de dev; **código canônico agora em `core4\confvision-rust-processor\`** |
| `ConfVision\` | Sistema antigo | **PARCIAL** — clone Git **independente**, commit **`7282bc7`** (atrás); layout Python na **raiz**; referência legada, **não** o Go app |

### Correção essencial

O sistema **unificado de produção** não é só uma das três pastas:

```text
core4\
├── home\confmonit\v4.0\confvision\   ← Go (control plane)
├── confvision\                       ← Python (data plane legado)
└── confvision-rust-processor\      ← Rust (analítico principal)
```

Evidências: `git worktree list` (piloto ligado a `core4\.git`); `ConfVision\.git` separado; `go build` OK em `home/.../confvision`; `cargo test` 136 OK em `confvision-rust-processor`.

---

## 2. Papel das três árvores

| Árvore | Git | HEAD (investigação) | Papel real |
|--------|-----|---------------------|------------|
| **core4** | Object store principal | `consolidate/confvision-phase7` | **Canônico** ecossistema + Go + Python + Rust |
| **core4-rust-pilot** | Worktree | `a2caf47` + WIP local | Histórico branch `rust-pilot`; **preservar**, não apagar |
| **ConfVision** | Clone | `7282bc7` + WIP decode | Snapshot antigo; diff decode vs core4 (ex.: `buffer::acquire` em core4 mais novo) |

---

## 3. Arquitetura nova

- **Control plane:** Go (`vis_*`, D5 assign, `vis_worker_ping`, PostgreSQL).
- **Analítico principal:** Rust (`CameraManager`, RTSP, decode, motion gate, YOLO, eventos, Redis `confvision:eventos`, storage S3).
- **Legado ativo:** Python motion, sensor, sync-agent, timelapse, dvr, MediaMTX.
- **Desligado:** `confvision-worker` (`main.py` analítico Python).

---

## 4. Serviços (código + operação)

| Serviço | Existe no repo? | Entry / código | Rust substitui? | Necessário? |
|---------|-----------------|----------------|-----------------|-------------|
| **confvision-rust-processor** | Sim (`core4/confvision-rust-processor/`) | `cargo run` / Docker | N/A (principal) | **Sim** |
| **confvision-worker** | Sim (`confvision/main.py`) | EasyPanel `python -u main.py` | **Sim** (analítico) | **Não** (STOPPED) |
| **confvision-motion** | Sim (`motion_main.py`) | Python MOG2/clips | Parcial | **Sim** (legado) |
| **confvision-sensor** | Sim (`sensor_main.py`) | Python | Não | **Sim** |
| **confvision-sync-agent** | Sim (`sync_agent_main.py`) | Python + Redis cache | Não | **Sim** |
| **confvision-timelapse** | Sim (`timelapse_main.py`) | Python | Parcial (Rust tem módulo timelapse **sobre frames**, serviço Python separado) | **Sim** |
| **confvision-dvr** | Sim (`dvr_main.py`) | MediaMTX record + API Go | Não | **Sim** |
| **confvision** (MediaMTX) | Sim (`Dockerfile-mediamtx`, `mediamtx/`) | Binário upstream + guard | Não | **Sim** |

**Quem inicia:** EasyPanel / systemd (env em `confvision/easypanel/`). **Câmera:** filtro `worker_id`, shard, `ListCamerasAnaliticas` (Go) → sync Rust/Python.

---

## 5. confvision-worker

- Documentação operacional: **STOPPED** (`easypanel/README.md`, `tenant-stack/README.md`, FASE-1 doc).
- **Não** há auto-start no código Go; depende de deploy manual.
- **Regra:** não reativar enquanto Rust analítico estiver no mesmo `worker_id`/câmera.

---

## 6. RTSP — consumidores

| Consumidor | Abre RTSP? | Função |
|------------|------------|--------|
| **Rust processor** | Sim (por câmera analítica assignada) | Analítico |
| **confvision-worker** | Sim (se ligado) | **Duplicaria Rust — proibido** |
| **motion** | Sim (paths configurados) | Clips movimento |
| **timelapse** | Sim | Frames timelapse |
| **dvr** | Via MediaMTX record (não necessariamente 2º RTSP pull analítico) | Gravação contínua |
| **sensor** | Event-driven (`event_capture`) | Terminal/sensor |

**Duplicação necessária documentada:** motion/timelapse/DVR podem coexistir com Rust **desde que** política de shard/worker_id e paths MediaMTX evitem segundo analítico na mesma câmera.

---

## 7. Fases 1–6 consolidadas

| Fase | Artefatos em `core4` |
|------|----------------------|
| 1 | `confvision-rust-processor/docs/FASE-1-RUST-CORE.md`, core RTSP/pipeline |
| 2 | FASE-2, motion/unificação |
| 3 | FASE-3, `events/` (catalog, dedup, engine) |
| 4 | FASE-4, `control_plane/` |
| 5 | FASE-5, `media/storage.rs`, docs storage/HA |
| 6 | FASE-6, runbooks scale/ops |

Índices: `confvision/docs/FASE-5-STORAGE-HA-INDEX.md`, `FASE-6-SCALE-OPERATIONS-INDEX.md`.

---

## 8. PostgreSQL

Migrations ConfVision relevantes (inventário — **não aplicadas nesta Fase 7 no ambiente local**):

| Migration | Caminho | Idempotência | Executada? |
|-----------|---------|--------------|------------|
| stream_policy | `home/.../sql/migrations/20260326_vis_camera_stream_policy.sql` | IF NOT EXISTS (verificar arquivo) | **NÃO VERIFICADO** (sem PG prod aqui) |
| coleta operacional | `home/.../sql/migrations/20260928_vis_coleta_operacional.sql` | idem | **NÃO VERIFICADO** |
| stream_error_diag | `confvision-rust-processor/sql/migrations/20260326_vis_camera_stream_error_diag.sql` | idem | **NÃO VERIFICADO** |

Scripts apply: `confvision-rust-processor/scripts/apply-postgres-migrations.ps1`, `home/.../scripts/apply-coleta-migration/`.

---

## 9. Redis

- Rust: `REDIS_URL`, `EVENT_QUEUE_KEY=confvision:eventos`, DLQ, startup check (`src/redis/mod.rs`).
- Python sync-agent: cache `confvision:sync:*` (`config_cache.py`).
- **Integração:** mesma infraestrutura esperada; conectividade **NÃO EXECUTADA** sem credenciais de ambiente.

---

## 10. Go ↔ processamento

- Sync câmeras / ping: endpoints `vis_*` em `router.go`, `vis_worker_ping`.
- D5: `rust_processor_d5.go`, `capacity-report` Rust, assign `worker_id`.
- Eventos: Rust → API Go (`vis_evento`); DVR Python → `vis_gravacao_segmento`.

---

## 11. Sistema antigo (`ConfVision\`)

- Commit **`7282bc7`**; Rust decode **diferente** de core4 (sem `buffer::acquire` em trechos comparados).
- **Registrar:** possível WIP local no clone; **não copiado** — core4 consolidado é **mais recente** que o clone.

---

## 12. Componentes

| Estado | Itens |
|--------|--------|
| **Substituído** | Analítico Python worker (operacionalmente) |
| **Mantido** | motion, sensor, sync-agent, timelapse, dvr, MediaMTX, Go |
| **Em migração** | Motion Python vs motion gate Rust; timelapse dual |
| **Futuro** | Lease anti-duplicata; desligar worker definitivamente no repo |

---

Ver também: `ARCHITECTURE_FINAL.md`, `SERVICOS_PROCESSAMENTO.md`, `DEPLOY_FASE_7.md`, `RELATORIO_FASE_7.md`.
