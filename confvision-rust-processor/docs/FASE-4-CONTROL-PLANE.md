# Fase 4 — Control Plane

**Escopo:** separar **decisão** (Control Plane) de **execução** (Data Plane) — auditado no código em 2026-09-30.  
**Pré-requisitos:** Fases 1–3 (núcleo Rust, vídeo parcial, eventos analíticos).

---

## 0. Estado real vs objetivo

| Conceito spec | Implementação hoje |
|---------------|-------------------|
| Video Node | **Host EasyPanel/Proxmox** + env `WORKER_ID` / `PROCESSOR_ID` |
| Node Registry | Tabela **`vis_worker`** (heartbeat via `/vis_worker_ping`) |
| Node Agent | **Não existe binário** — Rust processor + Python workers fazem sync/ping |
| Camera Registry | **`vis_camera`** (+ joins `vis_mediamtx_node`) |
| Assignment | Coluna **`vis_camera.worker_id`** + filtro em `ListCamerasAnaliticas` |
| Desired state | DB + flags (`ativo`, `deteção`, `analitico_pausado`, stream policy) |
| Actual state | Rust `CameraRuntimeState` + ping `cameras_ativas` |
| Reconciliation | **Pull:** Rust `sync_cameras` diff start/stop; Go D5 assign `worker_id` |
| Push commands | **Não** — sem `START_CAMERA` HTTP para o Node |
| Lease / epoch | **Não** |
| `vis_servidor` dedicado | **Não** — D5 usa `RUST_PROCESSOR_BASE_URLS` (`servidor_id\|url`) |

**Fase 4 concluída?** **Não** — control plane **parcial** (Go+Postgres+worker_id); faltam drift formal, lease, node agent unificado, reconciliação CP→push.

---

## 1. Arquitetura atual

```text
                    CONTROL PLANE (parcial)
                    Go API + PostgreSQL
                           │
     ┌─────────────────────┼─────────────────────┐
     │                     │                     │
     ▼                     ▼                     ▼
vis_camera          vis_worker            vis_mediamtx_node
(worker_id)         (ping/heartbeat)      (capacidade RTMP)
     │                     │
     │    D5: /vis_rust_processor_capacity
     │         /vis_camera_assign_processor
     │         /ops/d5/auto_assign_analiticas
     ▼
              DATA PLANE (por host)
     MediaMTX ← RTMP ← câmeras
          │
          ▼
     confvision-rust-processor
          │  GET /vis_camera_sync_ativas?worker_id=
          │  POST /vis_worker_ping
          ▼
     CameraManager.sync_cameras (desired list → workers)
```

Serviços Python **ainda** podem usar **sync-agent** → Redis `confvision:sync:*` (cache), enquanto Rust fala **direto** com a API Go.

---

## 2. Source of truth

| Domínio | Fonte oficial | Cache / réplica |
|---------|---------------|-----------------|
| Cadastro câmera | **PostgreSQL** `vis_camera` | Redis sync-agent (opcional) |
| Assignment analítico | **`vis_camera.worker_id`** | — |
| Nó MediaMTX | **`vis_mediamtx_node_id`** | — |
| Registry processor | **`vis_worker`** + env URLs D5 | `/capacity-report` live |
| Config runtime areas | Sync payload `config_version` | Redis `confvision:sync:analitico:...` |
| Estado RTSP/câmera | Rust in-memory + ping batch | `vis_stream_*` colunas |

---

## 3. Node identity

| Campo | Origem |
|-------|--------|
| `worker_id` | Env Rust/Python — **deve** bater com DB após assign |
| `processor_id` | Env — métricas/health (distinto de worker lógico) |
| `worker_tipo` | ex. `rust_processor`, `analitico`, `motion` |
| `hostname` | Ping |
| `vis_mediamtx_node_id` | Env / ping |

IP **não** é identidade primária.

---

## 4. Heartbeat (`vis_worker_ping`)

Go: `UpsertWorkerPing` → `vis_worker` (ON CONFLICT por `worker_id, worker_tipo, vis_mediamtx_node_id`).

Rust envia (via `main.rs` ping loop): câmeras ativas, capacity, shard, queue, stream health batch.

**Liveness:** ping recente + `ativo`.  
**Readiness para novas câmeras:** D5 `assign_eligible` + `/capacity-report` `allow_new_camera`.

---

## 5. Camera assignment

1. **Manual / API:** `POST /vis_camera_assign_processor` (force `worker_id` ou auto best).  
2. **Auto D5:** `POST /ops/d5/auto_assign_analiticas` se `D5_AUTO_ASSIGN_ENABLED`.  
3. **Sync pull:** processor só instancia câmeras cujo `worker_id` = seu `WORKER_ID` (modo `SHARD_MODE=worker_id` + `SYNC_FILTER_WORKER_ID`).

**Ownership:** um `worker_id` por câmera analítica no DB — **dois Rust com o mesmo `WORKER_ID`** = risco de duplicidade (config error).

**Anti-duplicidade parcial:** hash shard local + worker filter; **não** há lease global.

---

## 6. Desired vs actual state

| Desired | Actual (Rust) |
|---------|---------------|
| Lista sync API | `CameraManager` handles map |
| `worker_id` no DB | Processador só vê câmeras atribuídas |
| `analitico_pausado` | Stream policy pause / não sync |

**Drift (exemplos):**

- DB `worker_id=B`, processor A ainda rodando cam (se A não filtra ou cache stale).  
- Mitigação atual: filtro SQL + re-sync periódico.  
- **Detecção formal de drift CP** — **não implementada** (métrica/endpoint pendente).

**Reconciliação:** local no Rust (start/stop workers); **migração A→B** = alterar DB + wait sync — sem `STOP` remoto orchestrado.

---

## 7. Comunicação CP → Node

| Mecanismo | Uso |
|-----------|-----|
| **HTTP GET** sync | Rust/Python pull config |
| **HTTP POST** ping | Push telemetry Node → CP |
| **HTTP GET** capacity | CP scrape processors (D5) |
| **Redis** | Cache sync-agent, fila eventos — **não** comandos CP |
| WebSocket/gRPC | **Não** |

Modelo **pull** — CP indisponível: Rust continua último sync até falhar API (comportamento atual: logs + retry loops).

---

## 8. Sync Agent (`confvision-sync-agent`)

**Código:** `core4-rust-pilot/sync_agent.py`, `config_cache.py` ( **ausente** em `core4/confvision/` — build context issue).

| Função | Atual | Fase 4 alvo |
|--------|-------|-------------|
| Buscar `/vis_camera_sync_ativas` | sync-agent | CP API (já existe) |
| Gravar Redis `confvision:sync:*` | sync-agent | Opcional cache; Rust pode ignorar |
| `config_version` / since | sync-agent + Go | Go já suporta `since_version` |
| Workers leem cache | Python motion/worker | Migrar leitura direta API ou manter cache |
| RTMP auth cache | `config_cache` keys | Manter até MediaMTX auth no Go |

**Status migração:** **AUDITADO** — **não desativar** sync-agent sem substituir consumidores Redis.

---

## 9. Config cache Redis

Chaves: `confvision:sync:{full|analitico|gravacao}:n{node}:w{worker}` + `:version`.

TTL: `CONFIG_CACHE_TTL_SEC`. Backend alternativo: memory.

**Verdade:** Postgres + resposta Go; Redis é **réplica de leitura** para reduzir carga API.

---

## 10. PostgreSQL (principais)

| Tabela | Papel CP |
|--------|----------|
| `vis_camera` | Registry + `worker_id` + analítico/gravação |
| `vis_worker` | Node/processor heartbeat |
| `vis_mediamtx_node` | Capacidade ingest RTMP |
| `vis_evento` | Eventos (Fase 3) — não assignment |
| `vis_stream_relatorio` | Coleta operacional stream |

**Não criar** `vis_node` duplicada — estender `vis_worker` / documentar mapping Video Node = (`worker_id`, `worker_tipo`, `vis_mediamtx_node_id`).

---

## 11. Redis roles

| Uso | CP? |
|-----|-----|
| Sync cache | Réplica config |
| `confvision:eventos` | Data plane fila |
| Locks/lease | **Não** hoje |
| Pub/Sub CP | **Não** |

---

## 12. Segurança

- API Go: `VIS_WORKER_API_KEY` / Bearer (Rust `ConfVisionClient`).  
- D5 scrape processors: URLs públicas EasyPanel — **restringir** rede/TLS em produção.  
- Não logar secrets (RTSP userinfo redigido no Rust).

---

## 13. Observabilidade CP (hoje)

| Pergunta | Onde |
|----------|------|
| Processors up? | `/vis_rust_processor_capacity`, `/health` |
| Câmeras por worker? | SQL `vis_camera.worker_id`, ping `cameras_ativas` |
| Sobrecarga? | capacity-report, admission |
| Drift assignment? | **Manual** (SQL + health por host) |

Métricas spec `nodes_*`, `assignment_changes` — **parcial** via health JSON.

---

## 14. APIs relevantes (Go)

| Método | Path |
|--------|------|
| GET | `/vis_camera_sync_ativas` |
| GET | `/vis_camera_query_ativas` (legacy) |
| POST | `/vis_worker_ping` |
| GET | `/vis_rust_processor_capacity` |
| POST | `/vis_camera_assign_processor` |
| POST | `/ops/d5/auto_assign_analiticas` |
| GET | `/vis_mediamtx_node` |

---

## 15. Data Plane (Rust)

- `CameraManager::sync_cameras` — reconciliação local.  
- `control_plane::VideoNodeIdentity` — documentação identidade.  
- Sharding: `SHARD_MODE`, `WORKER_SHARD_*` (hash fallback).

---

## 16. Testes (checklist §37–41)

| Área | Status |
|------|--------|
| Unit Rust manager/admission | ✅ existentes |
| D5 assign integration | ⚠️ manual foxpro |
| Duplicidade A+B same worker_id | ⚠️ procedimento ops |
| Drift endpoint | ❌ |
| Escala 1000 nodes | ❌ |

---

## 17. Não fazer (Fase 4)

Failover automático, multi-region, remover sync-agent sem validação, novo bus comandos, reescrever Rust, tabela node duplicada.

---

## 18. Alterações desta entrega

- Documento + `src/control_plane/mod.rs` (identidade).  
- ADRs 010–013.  
- **Sem** migration SQL nova, **sem** desativar sync-agent.

---

## 19. Próxima fase

**FASE 5 — Storage + alta disponibilidade.**

---

## Referências

- `home/confmonit/v4.0/confvision/src/modulos/visdata/rust_processor_d5.go`
- `home/confmonit/v4.0/confvision/src/modulos/visdata/workers.go`
- `home/confmonit/v4.0/confvision/docs/DEPLOY_GO_CONFVISION.md`
- `deploy/tenant-stack/docs/PLANO_ATUALIZACAO_CONFVISION_ESCALA.md`
