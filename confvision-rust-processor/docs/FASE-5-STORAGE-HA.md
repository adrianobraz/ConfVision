# Fase 5 — Storage + Alta Disponibilidade

**Data da auditoria:** 2026-09-30  
**Escopo de código:** `core4-rust-pilot/confvision-rust-processor`, `core4/confvision` (Python workers), `core4/home/confmonit/v4.0/confvision` (Go visdata/visapi).

Auditoria baseada no **código atual**, não apenas em documentos de arquitetura anteriores.

---

## 1. Objetivo desta fase

Separar conceitualmente **metadata**, **mídia** e **runtime state**; documentar retenção, SPOFs e comportamento em falha; introduzir abstração mínima de media storage no Rust **sem** migrar tudo para object storage nem reescrever DVR.

---

## 2. Classificação de dados (A–E)

| Categoria | Exemplos | Onde está hoje (evidência) |
|-----------|----------|----------------------------|
| **A — Configuração** | `vis_camera`, `vis_mediamtx_node`, `worker_id`, plano, RTSP | **PostgreSQL** (`visdata/cameras.go`, sync APIs) |
| **B — Metadata** | `vis_evento`, `vis_evento_clip`, `vis_gravacao_segmento`, URLs | **PostgreSQL** (colunas `snapshot_url`, `video_url`, `s3_key`) |
| **C — Runtime** | fila eventos, cache sync, ping worker | **Redis** + memória processo; **PostgreSQL** `vis_worker.ultimo_ping_em` |
| **D — Media** | JPG/MP4 evento, segmentos DVR, timelapse | **Contabo S3** (path-style) + disco local temporário |
| **E — Cache** | `config_cache`, credenciais gravacao | **Redis** (`confvision:sync:*`) + TTL memória Python |

**PostgreSQL e vídeo:** busca por `bytea`/`BYTEA` no repositório **não encontrou** colunas binárias de vídeo. Mídia de evento = URLs + S3 keys em texto.

---

## 3. Checklist de conclusão (status honesto)

| Item | Status |
|------|--------|
| Dados classificados | ✅ doc |
| PostgreSQL auditado (schema via código) | ✅ parcial — **tamanho/volume não medidos** neste ambiente |
| Redis auditado | ✅ doc |
| Media Storage auditado | ✅ doc |
| DVR auditado | ✅ doc — **sem reescrita** |
| Timelapse / Snapshot / Clip auditados | ✅ doc |
| Retenção documentada | ✅ onde existe no código; resto **NÃO DEFINIDO** |
| Limpeza documentada | ✅ parcial |
| Orphan / missing media | ✅ **procedimento**; job automático **pendente** |
| Storage abstraction | ✅ `MediaStorage` Rust |
| Falha Redis / PG / Storage / Node | ⚠️ **comportamento inferido do código**; testes chaos **NÃO EXECUTADOS** |
| Restart / recovery / graceful shutdown | ✅ doc + código Rust shutdown |
| Backup / restore | ⚠️ **não auditado em infra** — marcar pendência |
| RPO / RTO | **A DEFINIR** (negócio) |
| SPOFs | ✅ tabela em `HIGH_AVAILABILITY.md` |
| Capacidade Node / storage | ⚠️ métricas parciais (`vis_worker` ping); projeções **estimativa** |
| Métricas Fase 5 consolidadas | ⚠️ **propostas** — maioria ainda não exposta |
| `DEPENDENCY_GRAPH.md` | ✅ |
| `STORAGE_ARCHITECTURE.md` | ✅ |
| `HIGH_AVAILABILITY.md` | ✅ |
| ADRs atualizados | ✅ ADR-014–018 |

---

## 4. Referências cruzadas

- [STORAGE_ARCHITECTURE.md](./STORAGE_ARCHITECTURE.md)
- [HIGH_AVAILABILITY.md](./HIGH_AVAILABILITY.md)
- [DEPENDENCY_GRAPH.md](./DEPENDENCY_GRAPH.md)
- [ARCHITECTURE_DECISIONS.md](./ARCHITECTURE_DECISIONS.md)

---

## 5. Próxima fase

**FASE 6 — ESCALA GLOBAL E OPERAÇÃO** (somente após fechar pendências operacionais desta fase: testes de falha, backup/restore validado, métricas de storage).
