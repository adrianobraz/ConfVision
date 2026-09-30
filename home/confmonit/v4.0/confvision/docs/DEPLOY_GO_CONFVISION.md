# Deploy — ConfVision Go (API central)

Servidor de produção da **API Go** (não confundir com foxpro EasyPanel — lá ficam MediaMTX, Rust e YOLO).

| Ambiente | Host típico |
|----------|-------------|
| API central | Proxmox → `vision.confmonit2.com.br` |
| Binário no servidor | `/home/confmonit/v4.0/confvision/confvision` |
| Serviço | `confmonit4confvision` |

---

## 1. Build (Windows dev)

```powershell
cd c:\sistemaconfmonit\core4\home\confmonit\v4.0\confvision
.\build.ps1
```

Confirme na saída: `vis_rust_processor_capacity=True`.

---

## 2. Enviar binário

1. **FileZilla** (SFTP): substituir `/home/confmonit/v4.0/confvision/confvision` pelo arquivo gerado.
2. **SSH** no Proxmox:

```bash
cd /home/confmonit/v4.0/confvision
chmod +x confvision
sudo systemctl restart confmonit4confvision
sudo systemctl status confmonit4confvision --no-pager
```

---

## 3. PostgreSQL (coleta + stream policy)

**Não é obrigatório rodar SQL manualmente** se o binário novo já inclui coleta no boot: ao subir, o ConfVision executa `coleta_operacional_migration.sql` (idempotente — `IF NOT EXISTS`).

Tabelas/colunas criadas:

- `vis_camera`: colunas de stream policy (`stream_falhas_consecutivas`, `stream_motivo_pausa`, …)
- `vis_stream_relatorio`
- `vis_sistema_health`
- `vis_sistema_metric`

### Opcional — aplicar SQL à mão (psql)

```bash
# No servidor ou máquina com acesso ao Postgres central
psql "$POSTGRES_URL" -f src/modulos/visdata/coleta_operacional_migration.sql
```

Arquivo no repo: `home/confmonit/v4.0/confvision/src/modulos/visdata/coleta_operacional_migration.sql`

Após restart, nos logs do serviço deve aparecer: `[coleta] migration OK`.

---

## 4. Variáveis `.env` (servidor)

Arquivo no servidor: `/home/confmonit/v4.0/confvision/.env`  
Modelo versionado: `.env.producao.example` (copie trechos; **não commitar** `.env` com senhas).

### Obrigatório para D5 + coleta (piloto / multi-host)

```env
POSTGRES_URL=postgres://...

VIS_WORKER_API_KEY=<mesma chave usada pelos rust-processors>

# Formato: url simples OU servidor_id|url (vírgula entre entradas)
RUST_PROCESSOR_BASE_URLS=srv-foxpro|https://foxpro-rust-pilot.rkr351.easypanel.host,srv-foxpro|https://foxpro-rust-pilot-b.rkr351.easypanel.host

D5_AUTO_ASSIGN_ENABLED=1

COLETA_RELATORIO_ENABLED=1
COLETA_RELATORIO_INTERVAL=15m
```

Cada **novo host** Proxmox: adicione entradas `srv-confvision-XXX|https://...` na lista.

---

## 5. Testes pós-deploy

Substitua `BASE` e `KEY`:

```bash
# Health Postgres / worker API
curl -sS -H "Authorization: Bearer KEY" "https://BASE/vis_worker_health"

# D5 — todos os processors
curl -sS -H "Authorization: Bearer KEY" "https://BASE/vis_rust_processor_capacity"

# D5 — só um servidor lógico
curl -sS -H "Authorization: Bearer KEY" "https://BASE/vis_rust_processor_capacity?servidor_id=srv-foxpro"
```

Esperado: HTTP 200 e JSON com `processors` (não `rota nao implementada`).

UI: `/relatorio-operacional` (Administrator) — health/métricas após coleta.

---

## 6. Erros comuns

| Sintoma | Causa provável |
|---------|----------------|
| D5 → 404 `rota nao implementada` | Binário antigo sem `router.go` D5 — refazer build/deploy |
| D5 → processors unreachable | URL errada ou firewall; Rust sem HTTPS público |
| Coleta vazia | `COLETA_RELATORIO_ENABLED=0` ou migration falhou (ver logs) |
| Postgres indisponível | `POSTGRES_URL` errado no `.env` do servidor |

---

## Referências

- `build.ps1` — hints pós-build
- Escala multi-tenant: `core4-rust-pilot/deploy/tenant-stack/docs/PLANO_ATUALIZACAO_CONFVISION_ESCALA.md`
