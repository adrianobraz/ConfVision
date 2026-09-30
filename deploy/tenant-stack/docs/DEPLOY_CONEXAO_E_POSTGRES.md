# Deploy — correção de conexão + Postgres

Dois problemas operacionais corrigidos no código (branch **rust-pilot** ≥ commit deste pacote):

| # | Sintoma | Causa | Correção |
|---|---------|--------|----------|
| 1 | MediaMTX log `connection refused` em `127.0.0.1:8100/auth` no boot | MediaMTX subia antes do Guard HTTP | `start_mediamtx_guard.py` aguarda `GET /health` na porta 8100 |
| 2 | Container **rust-yolo-sidecar** crash loop `TypeError` MRO | Herança dupla `ThreadingMixIn` + `ThreadingHTTPServer` | `yolo_sidecar.py` usa só `ThreadingHTTPServer` |

---

## O que atualizar / compilar / redeploy

### 1. EasyPanel — serviço **confvision** (MediaMTX + Guard)

- **Git:** `adrianobraz/ConfVision`
- **Branch:** `rust-pilot` (ou `main` se `confvision/` já tiver o mesmo `start_mediamtx_guard.py`)
- **Dockerfile:** `confvision/Dockerfile.mediamtx` **ou** raiz conforme painel (contexto deve incluir `start_mediamtx_guard.py` com `_wait_guard_http`)
- **Ação:** **Rebuild + Redeploy** (não basta restart)
- **Stop:** serviço legado `confvision-rtmp-guard` separado (porta 8100 duplicada)

**Verificação pós-deploy (logs do container):**

```text
[START] Guard RTMP na :8100 ...
[START] Guard HTTP pronto (http://127.0.0.1:8100/health)
[START] MediaMTX /mediamtx /mediamtx.yml ...
```

### 2. EasyPanel — **rust-yolo-sidecar**

- **Branch:** `rust-pilot`
- **Build:** imagem que copia `confvision-rust-processor/scripts/yolo_sidecar.py`
- **Ação:** Rebuild + Redeploy
- **Env:** `YOLO_HTTP_PORT=8091`, `YOLO_WORKERS=auto`

**Verificação:** `GET http://<host>:8091/health` → JSON 200

### 3. EasyPanel — **foxpro-rust-pilot** e **foxpro-rust-pilot-b**

- **Ação:** Redeploy se imagem Rust desatualizada (não relacionado aos dois fixes acima, mas necessário para analíticas)
- **Compilar:** CI/EasyPanel build Rust (`cargo build --release` no Dockerfile)
- **Env:** ver `easypanel.env.producao-todas-cameras.example`

### 4. Proxmox — **confmonit4confvision** (Go central)

- **Código:** `home/confmonit/v4.0/confvision` (API + UI)
- **Compilar:** `go build` / pipeline habitual
- **Env críticos de conexão:**
  - `RTMP_GUARD_URL` → URL **interna** do Guard (IP foxpro + porta 8100 **ou** túnel; não deixar apontando para serviço Stop)
  - `VIS_WORKER_API_KEY`, Postgres, `RUST_PROCESSOR_BASE_URLS`

### 5. Postgres (foxpro / central)

Aplicar migration (uma vez):

```powershell
cd C:\sistemaconfmonit\core4-rust-pilot\confvision-rust-processor\scripts
.\apply-postgres-migrations.ps1 -DatabaseUrl "postgresql://USER:PASS@HOST:5432/DB"
```

Arquivo: `home/confmonit/v4.0/confvision/sql/migrations/20260928_vis_coleta_operacional.sql`

Inclui colunas `vis_camera` stream policy + tabelas `vis_stream_relatorio`, `vis_sistema_health`, `vis_sistema_metric`.

### 6. Ops — IPs banidos (RTMP)

Se câmeras “conectam e caem” por ban:

```powershell
.\rtmp-guard-unban-all.ps1 -GuardUrl "http://<foxpro>:8100" -AdminKey "<RTMP_GUARD_ADMIN_KEY>"
```

---

## Ordem recomendada de deploy

1. **Postgres** — migration  
2. **confvision** (MediaMTX+Guard) — rebuild  
3. **rust-yolo-sidecar** — rebuild  
4. **foxpro-rust-pilot** A + B — redeploy  
5. **Go central** — se houver release pendente  
6. `implantar-fases.ps1 -VerifyOnly`  

---

## O que **não** precisa recompilar

- Worker Python **confvision-worker** (permanece Stop)
- Xano / APIs `.xs`
