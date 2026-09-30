# Deploy — Fase 7 (baseado no código em `core4`)

Branch: **`consolidate/confvision-phase7`** (ou `main` após merge PR).  
Paths reais no servidor após clone do repo ConfVision (layout `core4`).

---

## 1. Backup

- `pg_dump` do schema ConfVision (PostgreSQL operacional).
- Exportar envs EasyPanel / `.env` (não estão no Git).
- Anotar: `git rev-parse HEAD` antes do pull.

---

## 2. Atualização Git

```bash
cd /caminho/para/core4   # ou checkout do repo com layout home/ + confvision/
git fetch origin
git checkout consolidate/confvision-phase7
git pull origin consolidate/confvision-phase7
```

---

## 3. Build

### Rust

```bash
cd confvision-rust-processor
cargo test
cargo build --release
```

Ou imagem Docker conforme `confvision-rust-processor/Dockerfile` e `DEPLOY_EASYPANEL.md`.

### Go

```bash
cd home/confmonit/v4.0/confvision
go build -o confvision .
```

---

## 4. PostgreSQL

**APLICAR NO AMBIENTE** apenas o que ainda não existir (verificar colunas antes).

Ordem sugerida:

1. `home/confmonit/v4.0/confvision/sql/migrations/20260326_vis_camera_stream_policy.sql`
2. `confvision-rust-processor/sql/migrations/20260326_vis_camera_stream_error_diag.sql`
3. `home/confmonit/v4.0/confvision/sql/migrations/20260928_vis_coleta_operacional.sql`

Ferramentas: `confvision-rust-processor/scripts/apply-postgres-migrations.ps1` (Windows) ou `psql -f` manual.

---

## 5. Redis

- Confirmar `REDIS_URL` no env do **rust-processor** e serviços Python que usam fila/cache.
- Keys: `confvision:eventos`, DLQ, cache sync (ver `.env.example`).

---

## 6. Rust (serviço)

- Deploy container/processo **confvision-rust-processor**.
- Env: `deploy/tenant-stack/env/rust-processor.env.example`, `easypanel.env.producao-automatico.example`.
- `WORKER_ID` alinhado com PostgreSQL / D5 Go.

---

## 7. Serviços Python

Redeploy (sem alterar entrypoints):

- `confvision-sync-agent` → `python -u sync_agent_main.py` (cwd `confvision/`)
- `confvision-motion` → `motion_main.py`
- `confvision-sensor` → `sensor_main.py`
- `confvision-timelapse` → `timelapse_main.py`
- `confvision-dvr` → `dvr_main.py`

**Manter `confvision-worker` STOPPED.**

---

## 8. MediaMTX

- Serviço **confvision** com `Dockerfile-mediamtx` / `mediamtx.env`.
- `RTMP_PUBLISH_SECRET` igual ao Go.

---

## 9. Inicialização

Go → MediaMTX → Rust → sync-agent → motion/sensor/timelapse/dvr.

---

## 10. Health checks

- Go: `GET /vis_health`
- Rust: endpoints documentados em `confvision-rust-processor/README.md`
- Scripts: `confvision-rust-processor/scripts/foxpro-stack-verify.sh`, `phase-c-verify.sh` (se ambiente foxpro)

---

## 11. Teste de câmera

- Câmera com `worker_id` do processador Rust; logs RTSP sem loop 404.
- Script referência: `scripts/verify-stream-policy.sh`

---

## 12. Teste de evento

- `scripts/d3-event-verify.sh` / fila Redis `confvision:eventos`.

---

## 13. Teste de gravação

- DVR: segmentos + POST `vis_gravacao_segmento` (ver `dvr_main.py`).

---

## 14. Rollback

1. `git checkout <commit-anterior>` (ex. tag antes do deploy ou `backup/consolidation-preserve-2026-09-30`).
2. Redeploy binários/imagens anteriores.
3. Restaurar envs salvos.
4. PG: restaurar dump se migration irreversível foi aplicada.
5. Manter worker STOPPED.

Ver também: `docs/DEPLOY_ATUALIZACAO.md`.
