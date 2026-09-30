# Atualização do servidor ConfVision (pós-consolidação)

Branch Git: **`consolidate/confvision-rust`** (ou `main` após merge).  
Layout: `core4` — Go em `home/confmonit/v4.0/confvision/`, Python em `confvision/`, Rust em `confvision-rust-processor/`.

---

## 1. Backup

- Snapshot PostgreSQL (`pg_dump` do schema ConfVision).
- Copiar envs EasyPanel / `.env` atuais (fora do Git).
- Anotar commit atual: `git rev-parse HEAD`.
- Opcional: imagem Docker / tag dos serviços em execução.

---

## 2. Atualização Git

```bash
cd /caminho/core4
git fetch origin
git checkout consolidate/confvision-rust
git pull origin consolidate/confvision-rust
```

---

## 3. Build Rust

```bash
cd confvision-rust-processor
cargo build --release
# ou rebuild da imagem Docker conforme DEPLOY_EASYPANEL.md
cargo test   # validar antes do deploy
```

---

## 4. Build Go

```bash
cd home/confmonit/v4.0/confvision
go build -o confvision .
# redeploy do serviço Go conforme Proxmox/EasyPanel existente
```

---

## 5. Atualização dos serviços (ordem sugerida)

1. **confvision** (MediaMTX + Guard) — sem downtime longo se possível.
2. **confvision-rust-processor** — analítico principal.
3. **confvision-sync-agent**, **motion**, **sensor**, **timelapse**, **dvr** — Python legado.
4. **confvision-worker** — permanece **STOPPED**.

---

## 6. PostgreSQL

**APLICAR NO AMBIENTE** somente migrations ainda não executadas. Inventário:

| Arquivo | Onde | Notas |
|---------|------|-------|
| `home/confmonit/v4.0/confvision/sql/migrations/20260326_vis_camera_stream_policy.sql` | Go tree | stream policy |
| `home/confmonit/v4.0/confvision/sql/migrations/20260928_vis_coleta_operacional.sql` | Go tree | coleta operacional |
| `confvision-rust-processor/sql/migrations/20260326_vis_camera_stream_error_diag.sql` | Rust tree | diagnóstico 404 |
| `confvision-rust-processor/sql/fase0_pilot_ops.sql` | Rust | ops piloto — revisar antes |
| Scripts em `home/.../scripts/apply-coleta-migration/` | Go | coleta — se ainda não aplicado |

Antes de cada script: `\dt`, `\d vis_camera`, verificar colunas. **Sem DROP TABLE.** Idempotente quando possível (`IF NOT EXISTS`).

---

## 7. Redis

- Confirmar URL/credenciais nos envs Rust e Python.
- Filas: `confvision:eventos` (Rust capture).
- Cache sync: `confvision:sync:*` (sync-agent).

---

## 8. MediaMTX

- Redeploy `Dockerfile-mediamtx` / env `mediamtx.env`.
- `RTMP_PUBLISH_SECRET` alinhado com Go.

---

## 9. Inicialização

- Subir Go → MediaMTX → Rust → sync-agent → demais Python.

---

## 10. Health checks

- Go: `GET /vis_health` (sem auth worker).
- Rust: endpoint health/metrics documentado em `confvision-rust-processor/README.md`.
- MediaMTX: API `:9997` se exposta.

---

## 11. Teste de câmera

- Câmera piloto online no Postgres (`stream_ok`, `worker_id` correto).
- Rust recebe sync e abre RTSP **uma vez** para analítico.
- Verificar logs sem loop 404 / NAL.

---

## 12. Teste de evento

- Disparo analítico → Redis → capture → `vis_evento` + S3.

---

## 13. Rollback

Ver [DEPLOY_ATUALIZACAO.md](./DEPLOY_ATUALIZACAO.md) § Rollback abaixo e `RELATORIO_CONSOLIDACAO_FINAL.md`.

### Rollback rápido

1. Anotar **commit novo** e **commit anterior** (`git log -2`).
2. `git checkout <commit-anterior>` ou redeploy imagem anterior.
3. Restaurar envs salvos no passo 1.
4. Se migration foi aplicada: **não** reverter SQL destrutivo; restaurar backup PG se necessário.
5. Serviços: voltar versão Go/Rust/Python; manter **worker STOPPED** se já estava.
6. Validar `vis_health` e uma câmera piloto.

| Item | Registro |
|------|----------|
| Commit anterior | `git rev-parse backup/consolidation-preserve-2026-09-30` ou tag anotada |
| Commit consolidado | `git rev-parse consolidate/confvision-rust` |
| Migrations aplicadas | listar manualmente após deploy |
| Serviços alterados | rust-processor, Go, MediaMTX, Python legado |
