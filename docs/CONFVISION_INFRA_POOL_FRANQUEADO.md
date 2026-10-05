# Pool de infra ConfVision por franqueado

## Objetivo

Garantir que toda câmera cadastrada receba **`vis_mediamtx_node_id`** e **`worker_id`** (analítico) de forma automática, sem depender do front, evitando câmeras “ao vivo ok” mas **fora do sync** do worker.

## Obrigatoriedade

- Cada franqueado que cadastra câmeras deve ter **pool MTX** e **pool worker analítico** configurados no **Administrator ConfCam** (`/administrator/rtmp-falhas` → aba **Infra franqueado**).
- Sem pool completo, `CreateCamera` retorna erro antes de inserir a câmera.

## Tabelas Postgres

Migration: `home/confmonit/v4.0/confcam/sql/022_franqueado_infra_pool.sql`

- `vis_franqueado_mediamtx_pool` — nós permitidos + `max_cameras` opcional por linha
- `vis_franqueado_worker_pool` — `worker_id` analítico + `max_cameras` (default 30)

## Assign na criação

`visdata/createCameraOnce`:

1. `ValidateFranqueadoInfraPool(id_franqueado)`
2. `pickMediamtxForFranqueadoTx` — menor carga, pula nós cheios (cota franqueado + teto físico do nó)
3. `pickWorkerForFranqueadoTx` — se plano exige analítico (`CameraNeedsWorkerID`)
4. INSERT com `vis_mediamtx_node_id` e `worker_id` escolhidos

## API Administrator (ADM)

| Método | Path |
|--------|------|
| GET | `/api/administrator/infra-pool/franqueados?q=` |
| GET | `/api/administrator/infra-pool/status?id_franqueado=` |
| GET | `/api/administrator/infra-pool/mediamtx-nodes` |
| POST | `/api/administrator/infra-pool/mediamtx` |
| POST | `/api/administrator/infra-pool/worker` |

## Deploy

1. Aplicar SQL `022` no Postgres central.
2. Deploy ConfCam (build `confvision` em `home/confmonit/v4.0/confcam`).
3. Configurar pool para cada franqueado ativo no Administrator.
4. Backfill câmeras existentes sem `worker_id` (UPDATE manual ou script ops futuro).

## Bootstrap exemplo (SQL)

```sql
INSERT INTO vis_franqueado_mediamtx_pool (id_franqueado, vis_mediamtx_node_id, max_cameras, ordem, ativo)
VALUES ('ID_FRANQUEADO', 1, 180, 1, true)
ON CONFLICT (id_franqueado, vis_mediamtx_node_id) DO UPDATE SET max_cameras = EXCLUDED.max_cameras, ativo = true;

INSERT INTO vis_franqueado_worker_pool (id_franqueado, worker_id, worker_tipo, max_cameras, ordem, ativo)
VALUES ('ID_FRANQUEADO', 'worker-vps1', 'analitico', 30, 1, true)
ON CONFLICT (id_franqueado, worker_id, worker_tipo) DO UPDATE SET max_cameras = EXCLUDED.max_cameras, ativo = true;
```

Repetir linhas de worker para `worker-vps1-02` … conforme capacidade (ex. 6×30 para 180 MTX).
