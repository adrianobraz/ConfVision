# Fase 1 — Provisionamento tenant (`ct_cli_*`)

Padrão para **Proxmox** (CT/LXC ou VM) ou **Docker Compose**: um tenant = identidade `WORKER_ID` / `PROCESSOR_ID` + teto `MAX_CAMERAS`.

Exemplos de nome:

- `ct_cli_rva` (200 câmeras)
- `ct_cli_powerseg` (200 câmeras)
- `ct_cli_softpark-shard-042` (fatia de 200–400 em escala Softpark)

Sem alteração no Xano.

---

## 1. Modelo por servidor GPU

```text
Servidor Proxmox (GPU, ~400 detecções)
├── ct_cli_rva          → MAX_CAMERAS=200
└── ct_cli_powerseg     → MAX_CAMERAS=200
```

Cada **tenant** = stack mínima (3 serviços) + câmeras no Postgres com `worker_id` daquele tenant (processamento Rust).

**MediaMTX:** pode ser **por tenant** (Compose abaixo) ou **compartilhado no host** (EasyPanel `confvision` único + vários Rust). Compose exemplo = MTX **por tenant** (isolamento simples).

---

## 2. Identidade (`WORKER_ID`)

| Campo | Regra |
|-------|--------|
| `TENANT_ID` | Ex.: `ct_cli_rva` — slug único |
| `PROCESSOR_ID` (Rust) | **Igual** `WORKER_ID` |
| `WORKER_ID` (Rust env) | **Igual** `PROCESSOR_ID` |
| Postgres `vis_camera.worker_id` | **Igual** para câmeras processadas por este Rust |
| `WORKER_TIPO` | `rust_processor` |
| `SHARD_MODE` | `worker_id` (piloto / multi-tenant lógico) |

Publisher Python (ingest):

| Campo | Regra |
|-------|--------|
| `WORKER_ID` | **`${TENANT_ID}-ingest`** recomendado enquanto Rust consome sync por `worker_id` das câmeras |

**Migração prática:**

1. Câmeras do cliente no Postgres com `worker_id = ct_cli_<slug>` (Rust processa).
2. Worker **ingest** publica RTMP no MediaMTX (câmeras ainda listadas no sync do ingest **ou** RTMP direto da câmera para o path `cam/{hash}`).
3. Durante piloto foxpro, muitos sites usam **worker host + MTX compartilhado**; replique env `MEDIAMTX_RTSP_BASE` interno.

Detalhe operacional: [RECUPERACAO_CAMERAS.md](../../../confvision-rust-processor/docs/RECUPERACAO_CAMERAS.md).

---

## 3. `MAX_CAMERAS`

| Cenário | Valor |
|---------|--------|
| Piloto | `1` |
| Tenant metade GPU | `200` |
| Tenant GPU cheio | `400` (validar benchmark Fase 2 GPU) |
| Sem limite hard | `0` (não recomendado em produção) |

Rust: env `MAX_CAMERAS` + truncamento no sync.  
Worker ingest: alinhar `MAX_CAMERAS` / shard para não sobrecarregar ingest.

---

## 4. EasyPanel (foxpro) — 3 apps por tenant

Replicar por cliente (ou usar 1 MTX + N Rust):

| App | Dockerfile | Env |
|-----|------------|-----|
| `confvision` ou `mtx-<tenant>` | `Dockerfile-mediamtx` | `env/mtx.env.example` |
| `worker-<tenant>-ingest` | `Dockerfile` (repo ConfVision) | `env/publisher.env.example` |
| `rust-<tenant>` | `confvision-rust-processor/Dockerfile` | `env/rust-processor.env.example` |

**Rede:** mesma rede Docker; Rust `MEDIAMTX_RTSP_BASE=rtsp://<hostname-mtx>:8554`.

Templates completos: [`../env/`](../env/).

---

## 5. Docker Compose (lab / CT)

```bash
cd deploy/tenant-stack
export TENANT_ID=ct_cli_rva
export MAX_CAMERAS=200
./scripts/tenant-init-env.sh
# Preencher segredos em .env.tenant
docker compose -f docker-compose.tenant.example.yml --env-file .env.tenant up -d --build
```

---

## 6. Postgres — atribuir câmeras ao tenant

```sql
-- Listar
SELECT id, nome, worker_id FROM vis_camera
WHERE cliente_id = <ID_CLIENTE> AND ativo = true;

-- Atribuir fatia ao tenant Rust (exemplo)
UPDATE vis_camera
SET worker_id = 'ct_cli_rva'
WHERE id IN (...);
```

Rollback:

```sql
UPDATE vis_camera SET worker_id = '<worker_python_anterior>' WHERE id = <id>;
```

---

## 7. Checklist pós-provisionamento

- [ ] `tenant-init-env.sh` executado; segredos preenchidos (não commitados).
- [ ] 3 containers healthy.
- [ ] `./scripts/tenant-stack-smoke.sh` OK.
- [ ] `/health`: `cameras_online` > 0 para câmeras atribuídas.
- [ ] Documentar host Proxmox + `TENANT_ID` na planilha de capacidade (400/GPU).

---

## 8. Escala Softpark (referência)

200.000 câmeras ÷ 400 câmeras/GPU ≈ **500 servidores**.  
Cada servidor repete este playbook com `TENANT_ID` diferente por fatia (`ct_cli_softpark-001` …).

Automação (Ansible/Terraform) deve gerar:

- `.env.tenant` por CT
- `WORKER_ID` / DNS interno
- regras SQL de shard

---

## Links

- [Fase 0 checklist](./FASE0_CHECKLIST_OPS.md)
- [Deploy Rust EasyPanel](../../../confvision-rust-processor/DEPLOY_EASYPANEL.md)
- [Deploy ConfVision EasyPanel](../../../DEPLOY_EASYPANEL.md)
