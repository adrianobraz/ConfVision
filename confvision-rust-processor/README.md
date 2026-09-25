# ConfVision Rust Processor

Serviço **independente** do ConfVision — *processing plane* em Rust.

| Item | Descrição |
|------|-----------|
| **Objetivo** | Worker/processador que sincroniza câmeras via API Go, conecta RTSP (MediaMTX), conta frames e expõe health/metrics |
| **Fase atual** | **Fase 1** — validação de processamento RTSP real (sem YOLO/GPU/eventos completos) |
| **Não substitui** | API **Go**, workers **Python**, MediaMTX, Postgres direto, Xano |

## Onde fica o código

| Contexto | Caminho |
|----------|---------|
| Desenvolvimento (`core4`) | `C:\sistemaconfmonit\core4\confvision-rust-processor` |
| Repositório GitHub | `confvision-rust-processor/` em [adrianobraz/ConfVision](https://github.com/adrianobraz/ConfVision) |

Visão geral dos projetos: [`ARQUITETURA_PROJETOS.md`](../ARQUITETURA_PROJETOS.md) (este repositório). Ambiente dev: `C:\sistemaconfmonit\core4\ARQUITETURA_CORE4.md`.

## Arquitetura

```
                ┌──────────────────────┐
                │       Go API         │
                │  PostgreSQL / vis_*  │
                └──────────┬───────────┘
                           │
                sync /vis_camera_sync_ativas
                ping /vis_worker_ping
                           │
          ┌────────────────┴────────────────┐
          │                                 │
 ┌────────▼────────┐              ┌────────▼────────┐
 │ Python Worker   │              │ Rust Processor   │
 │ (existente)     │              │ (este serviço)   │
 └────────┬────────┘              └────────┬────────┘
          │                                 │
          └──────────────┬──────────────────┘
                         │
                   ┌─────▼─────┐
                   │ MediaMTX  │
                   │ RTSP      │
                   └───────────┘
```

- **Go:** control plane — URL base em `CONFVISION_API_URL` (nunca `*.xano.io`).
- **Python:** continua responsável por analítico/YOLO/eventos em produção; use `PROCESSOR_ID` (processo) distinto de `WORKER_ID` (worker lógico) quando coexistirem.
- **MediaMTX:** streams RTSP; path `cam/{hash12}` quando necessário (`RTMP_PUBLISH_SECRET` alinhado ao Go).

## Fase 1 — escopo

**Inclui:**

- Sincronização periódica de câmeras ativas (filtro analítico / `worker_id` conforme env)
- Ping periódico para a API Go (`worker_tipo=rust_processor`)
- Sessões RTSP (crate `retina`), reconexão, contagem de frames
- HTTP: `/health`, `/ready`, `/metrics`
- Coexistência com worker Python

**Não inclui (neste momento):**

- YOLO, CUDA
- Processamento completo de eventos (`vis_evento`)
- Redis ativo (apenas placeholder)
- Postgres direto
- Xano (`XANO_*` rejeitado; URL Xano recusada)
- Substituição do Go ou do Python

## Árvore do projeto

```
confvision-rust-processor/
├── Cargo.toml
├── Cargo.lock
├── Dockerfile
├── .dockerignore
├── .gitignore
├── .env.example
├── easypanel.env.example
├── DEPLOY_EASYPANEL.md
├── confvision-rust-processor.service.example
├── README.md
└── src/
    ├── main.rs
    ├── config.rs
    ├── api/client.rs
    ├── camera/
    ├── rtsp/session.rs
    ├── worker/
    ├── health/
    ├── metrics/
    ├── events/    # placeholder fase 2
    ├── redis/     # placeholder
    └── media/     # placeholder S3
```

Não versionar `target/` nem `.env` com segredos reais.

## Variáveis de ambiente

Referência completa: [`.env.example`](.env.example) e [`easypanel.env.example`](easypanel.env.example) (placeholders — **não** commitar valores reais).

| Variável | Função |
|----------|--------|
| `CONFVISION_API_URL` | Base da API Go (obrigatória) |
| `VIS_WORKER_API_KEY` | Header `X-Vis-Worker-Key` |
| `RTMP_PUBLISH_SECRET` | Salt para path `cam/{hash12}` |
| `MEDIAMTX_RTSP_BASE` | RTSP base se a API não enviar por câmera |
| `PROCESSOR_ID` | ID do processo (health/métricas); não substitui `WORKER_ID` |
| `WORKER_ID` | Identidade lógica do worker (ping e sync com `SHARD_MODE=worker_id`) |
| `WORKER_TIPO` | Default `rust_processor` |
| `SHARD_MODE` | `auto`, `worker_id` ou `hash` (igual `ConfVision/sharding.py`) |
| `WORKER_SHARD_INDEX`, `WORKER_SHARD_TOTAL` | Shard por hash (`camera_id % total == index`) |
| `MEDIAMTX_NODE_ID` | Query `vis_mediamtx_node_id` na sync (0 = sem filtro) |
| `MAX_CAMERAS`, `SYNC_INTERVAL_SEC`, `PING_INTERVAL_SEC` | Limite **local** e intervalos |
| `HTTP_HOST`, `HTTP_PORT` | Servidor local (default `8090`) |
| `LOG_LEVEL` | `info` ou `debug` |
| `RTSP_*`, `FRAME_BUFFER_MAX` | Timeouts e buffer RTSP |
| `RTSP_SIMULATE` | Apenas teste local sem RTSP real |
| `MOTION_ANALYSIS_MAX_FPS`, `MOTION_FRAME_STRIDE`, `DECODE_FRAME_STRIDE` | Teto e stride de decode+motion ([`docs/ANALYSIS_CPU_LEGACY.md`](docs/ANALYSIS_CPU_LEGACY.md)) |
| `ANALYSIS_LEGACY_VPS` | `1` = defaults tipo worker Python (FPS 2, strides 3/5) |
| `ANALYSIS_ONLY_ON_MOTION` | Decode frequente só após movimento (≈ `YOLO_ONLY_ON_MOTION`) |
| `MOTION_GATE_PROBE_MAX_FPS`, `MOTION_GATE_MISS_FRAMES` | Probe em cena parada / desarmar |
| `REDIS_URL`, `S3_*` | Reservados — Fase 1 não usa |

## Integração com Go

Endpoints usados na Fase 1:

- `GET /vis_camera_sync_ativas` — lista câmeras para o processor
- `POST /vis_worker_ping` — heartbeat (`WorkerPingInput` compatível)

Campos enviados no ping incluem `worker_id`, `worker_tipo`, `hostname`, `versao`, `cameras_ativas`, `ultimo_ping_em`, `ativo`, `max_cameras`, etc. Detalhes no final deste README (compatibilidade).

**Nenhum endpoint Go novo é obrigatório na Fase 1.**

## Integração com MediaMTX

- Conexão RTSP TCP via `retina`
- URL por câmera da API ou montada com `MEDIAMTX_RTSP_BASE` + hash
- Em deploy Docker, `MEDIAMTX_RTSP_BASE` deve ser **alcançável de dentro do container**

## EasyPanel

Piloto com **1 câmera**, serviço Docker **separado**: [`DEPLOY_EASYPANEL.md`](DEPLOY_EASYPANEL.md).

Build context no GitHub: pasta **`confvision-rust-processor/`** na raiz do repo ConfVision.

## Como executar

1. [Rust](https://rustup.rs) 1.75+ (Dockerfile usa 1.88).
2. Copie `.env.example` → `.env` e preencha **localmente** (não commitar).
3. Build e run:

```bash
cd confvision-rust-processor
cargo build --release
# Linux/macOS:
./target/release/confvision-rust-processor
# Windows:
target\release\confvision-rust-processor.exe
```

Docker:

```bash
docker build -t confvision-rust-processor:0.1.0 .
docker run --env-file rust-processor.env -p 8090:8090 confvision-rust-processor:0.1.0
```

## Como testar

**Sem RTSP:**

```bash
# Linux/macOS
export RTSP_SIMULATE=1
# Windows
set RTSP_SIMULATE=1
cargo test
cargo run
```

**Uma câmera real:**

1. No Postgres (via fluxo Go), `vis_camera.worker_id = '<WORKER_ID>'` só na câmera de teste.
2. Python mantém outro `WORKER_ID` nas demais.
3. `SHARD_MODE=worker_id`, `WORKER_ID` igual ao Postgres, `MAX_CAMERAS=1`.
4. Verificar logs (`rtsp connected`, frames) e `GET http://localhost:8090/ready`.

**Health:**

- `GET /health` — agregados + contagem por status (`online`, `offline`, `reconnecting`, `starting`)
- `GET /ready`
- `GET /metrics` — métricas globais (`metrics`) + array `cameras[]` com contadores por câmera

**Multi-câmera (Fase 5.0):** uma task Tokio por câmera, pipeline/decoder/motion/RTSP isolados por sessão; limite apenas via `MAX_CAMERAS` (env).

## O que NÃO fazer

- Apontar `CONFVISION_API_URL` para Xano ou definir `XANO_*`.
- Reutilizar o mesmo `worker_id` / ping do Python sem `worker_tipo` distinto.
- Publicar o monorepo `core4` inteiro no GitHub para “subir o Rust”.
- Mover a API Go para este repositório ou pasta.
- Assumir que `core4\confvision` (Python) é este serviço.
- Commitar `.env`, `target/` ou logs com credenciais.

## Fase 6.0 — Dynamic Capacity

O processor estima **capacidade dinâmica** (câmeras/FPS) com base na carga observada e nos recursos medidos.

- **`MAX_CAMERAS`** continua existindo apenas como **hard safety limit** local (use `0` para sem teto). **Não** é capacidade estimada nem capacidade do servidor.
- **`CAPACITY_MODE=dynamic`** (default) liga o Capacity Engine; `disabled` desliga estimativas.
- Coleta periódica (`CAPACITY_SAMPLE_INTERVAL_SEC`, default 5s) — **fora** do caminho crítico de cada frame.
- Janela mínima `CAPACITY_MIN_SAMPLE_SEC` (default 30s): até lá, `capacity_state=unknown`.
- Métricas em `GET /metrics` → objeto `capacity` (completo) e resumo em `GET /health`.

### How capacity is calculated

1. Média móvel sobre `CAPACITY_HISTORY_SIZE` amostras (CPU, RAM, FPS, câmeras online, drops, fila, latência).
2. Para cada recurso medido (CPU/RAM; GPU/VRAM só se disponíveis via infra existente):

   `capacity_fps ≈ current_fps × (target × safety_factor) / current_usage`

   `capacity_cameras ≈ capacity_fps / average_fps_per_camera`

3. A capacidade final usa o **mínimo** entre recursos (recurso limitante → `limiting_resource`).
4. Valores são **arredondados para baixo**; sem amostra suficiente → campos `null` e `unknown`.
5. Cada métrica de recurso inclui `scope`: `host`, `container`, `process` ou `unavailable` (nunca mascarar container como host).

Variáveis: ver [`.env.example`](.env.example) (`CAPACITY_*`).

API interna (Fase 7): `CapacityEngine::current()`, `snapshot()`, `estimate()`.

## Roadmap (próximas fases — não implementadas)

| Fase | Ideia |
|------|--------|
| **7** | Assignment automático (Cluster Manager) — consome `CapacityEstimate` |
| **2** | Eventos, filas (Redis), integração analítica ampliada |
| **3** | YOLO/GPU ou delegação coordenada com Python |

Alterações de contrato Go devem ser documentadas e feitas **no projeto Go**, não aqui sem alinhamento.

## Dependências Rust (resumo)

| Crate | Uso |
|-------|-----|
| tokio | async, sinais |
| axum | HTTP |
| reqwest | cliente Go |
| retina | RTSP |
| hash-ids | path `cam/{hash12}` |
| serde, tracing, chrono, dotenvy | JSON, logs, tempo, `.env` |

## Contrato Go — compatibilidade (Fase 1)

`POST /vis_worker_ping` aceita campos do `WorkerPingInput` Go, incluindo `worker_id`, `worker_tipo`, `hostname`, `versao`, `cameras_ativas`, `ultimo_ping_em`, `ativo`, `shard_index`, `shard_total`, `yolo_device`, `queue_backend`, `vis_mediamtx_node_id`, `max_cameras`.

O Rust envia `worker_id` a partir de `WORKER_ID` (não de `PROCESSOR_ID`). Use `worker_tipo=rust_processor` para não colidir com ping Python (`analitico`) no mesmo ID.

Desligamento: SIGINT/SIGTERM → para RTSP, ping final `ativo=false`, encerra HTTP.

## ANTES DE ALTERAR O CONFVISION

1. Confirme se a mudança é Go, Python, Rust, MediaMTX ou deploy.
2. Verifique o caminho compleo no disco (`core4` vs clone `ConfVision`).
3. Rust no GitHub = só pasta `confvision-rust-processor/`.
4. Não merge/rebase cego entre `core4` e `ConfVision`.

Regras completas: seção homônima em [`ARQUITETURA_PROJETOS.md`](../ARQUITETURA_PROJETOS.md).
