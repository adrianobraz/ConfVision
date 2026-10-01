# Deploy — Configuração de Atendimento (checklist completo)

> **Situação:** código pronto no repositório local; **nada disso entra em produção até você executar os passos abaixo.**  
> **Ligações e WhatsApp atuais não param** enquanto a função #22 no Xano **não** consultar a config Postgres (ainda não plugada).

---

## Ordem recomendada (resumo)

| # | O quê | Onde |
|---|--------|------|
| 0 | Backup | Postgres + MySQL + anotar versões em produção |
| 1 | Schema Postgres `008` | PostgreSQL |
| 2 | Export clientes MySQL → colar no `009` | MySQL → Postgres |
| 3 | Liberação inicial `009` | PostgreSQL |
| 4 | Catálogo créditos (opcional, Meu plano) | MySQL |
| 5 | Build + deploy **confvision** (Go) | VPS Core4 |
| 6 | Build + deploy **franqueadoadmin** | VPS Core4 |
| 7 | Build + deploy **eventgateway** + **taskxano** | VPS |
| 8 | `.env` com `POSTGRES_URL` em todos os serviços | VPS |
| 9 | Push Xano (só se quiser alinhar online) | Xano CLI |
| 10 | Conferência | telas + logs |

---

## 0 — Backup

```powershell
# Postgres (ajuste host/usuário)
pg_dump "postgres://confmonit:SENHA@HOST:5432/confmonit?sslmode=disable" -Fc -f backup_pre_atendimento.dump

# MySQL confmonitV4 — backup via phpMyAdmin ou mysqldump
```

---

## 1 — PostgreSQL: schema (`008`)

> **Rodar no servidor Core4** (rede `10.2.2.120`) ou pgAdmin conectado ao Postgres.  
> Desta máquina de dev externa o Postgres **bloqueia** (`pg_hba` / timeout).

**Script rápido (no servidor):**
```powershell
cd confvision\sql
.\apply_atendimento.ps1
```

**Arquivo:** `confvision/sql/008_atendimento_schema.sql`  
**Roda em:** PostgreSQL (pgAdmin, DBeaver, psql) — **NUNCA no MySQL/phpMyAdmin**

```powershell
cd c:\sistemaconfmonit\core4\confvision\sql\cmd\apply

$env:POSTGRES_URL = "postgres://confmonit:SENHA@HOST:5432/confmonit?sslmode=disable"

go run . ..\..\008_atendimento_schema.sql
```

**Cria:** `ops_franqueado_atendimento_politica`, `ops_cliente_atendimento_config`, listas, grade, `ops_ia_bloqueio_cliente`, créditos, staging, etc.

**Conferir:**

```sql
SELECT tablename FROM pg_tables
WHERE schemaname = 'public' AND tablename LIKE 'ops_%atend%'
ORDER BY 1;
```

---

## 2 — MySQL: gerar INSERTs dos clientes ativos

**Arquivo:** `home/confmonit/v4.0/api/sql/20260811_mysql_export_clientes_postgres.sql`  
**Roda em:** MySQL (`confmonitV4`) — phpMyAdmin ou cliente MySQL

1. Execute a query.
2. Exporte a coluna `sql_postgres` (todas as linhas).
3. Abra `confvision/sql/009_atendimento_liberacao_inicial.sql`.
4. Cole os `INSERT INTO ops_stg_cliente_ativo ...` **entre** os comentários `-- PASTE` e `-- FIM PASTE`.

> Se não colar nada, o `009` ainda tenta popular a staging a partir de `ops_alarm_events` (fallback parcial).

---

## 3 — PostgreSQL: liberação inicial (`009`)

**Arquivo:** `confvision/sql/009_atendimento_liberacao_inicial.sql`  
**Roda em:** PostgreSQL — **NUNCA no MySQL**

```powershell
cd c:\sistemaconfmonit\core4\confvision\sql\cmd\apply
go run . ..\..\009_atendimento_liberacao_inicial.sql
```

**Efeito para clientes da staging:**

- `inteligencia_artificial = TRUE`
- `finalizacao_automatica = TRUE`
- parceiro e e-mail = **FALSE**
- política franqueado: IA + autofim modo `todos`

**Conferir:**

```sql
SELECT COUNT(*) FROM ops_stg_cliente_ativo;
SELECT COUNT(*) total,
       SUM(inteligencia_artificial::int) com_ia,
       SUM(finalizacao_automatica::int) com_autofim
FROM ops_cliente_atendimento_config;
```

---

## 4 — MySQL: catálogo créditos (opcional — compra Meu plano)

**Arquivo:** `home/confmonit/v4.0/apifunction/sql/004_fp_catalogo_creditos_atendimento.sql`  
**Roda em:** MySQL `confmonitV4` (produtos crédito ligação/SMS/WhatsApp/e-mail + finalização avulsa).

Pode rodar **depois**; a tela Config Atendimento funciona sem isso (créditos só leitura até checkout existir).

---

## 5 — Deploy confvision (Go Core4)

**Código:** `home/confmonit/v4.0/confvision/` — rotas `/ops/atendimento/*` em `visdata`.

### `.env` (servidor)

```env
POSTGRES_URL=postgres://confmonit:SENHA@HOST:5432/confmonit?sslmode=disable
```

### Build e publicar

```powershell
cd c:\sistemaconfmonit\core4\home\confmonit\v4.0\confvision
.\build.ps1
# Subir binário + recursos via FileZilla (fluxo habitual)
# systemctl restart confmonit4confvision
```

### Teste rápido (substitua id franqueado)

```http
GET https://SEU_DOMINIO/ops/atendimento/politica?id_franqueado=XXX
GET https://SEU_DOMINIO/ops/atendimento/resolve?id_franqueado=XXX&id_cliente=YYY&recurso=inteligencia_artificial
```

---

## 6 — Deploy franqueadopromais (FranqueadoPro)

**Código:** `home/confmonit/v4.0/franqueadopromais/`

- Nova tela: `/carregar-config-atendimento`
- Bloqueio IA: Postgres (telefones IA ainda via Xano `WhatsEventCadFranq`)
- Permissão: `atendimento.configuracao` (plano Pro)

### `.env` (servidor — **obrigatório para config + bloqueio IA**)

```env
POSTGRES_URL=postgres://confmonit:SENHA@HOST:5432/confmonit?sslmode=disable
```

### Build

```powershell
cd c:\sistemaconfmonit\core4\home\confmonit\v4.0\franqueadopromais
.\build.ps1 -Package
# Subir deploy\ + binário
# systemctl restart confmonit4franqueadopro
```

### Conferir na UI

1. Login franqueado plano Pro.
2. Menu Atendimento → **Configuração de Atendimento**.
3. `/carregar-inteligencia-artificial` → bloqueios (Postgres).

---

## 7 — Deploy eventgateway + taskxano

### eventgateway `.env`

```env
POSTGRES_URL=postgres://confmonit:SENHA@HOST:5432/confmonit?sslmode=disable
# Ligue só DEPOIS de 008+009 OK e testado:
OPS_SIDECAR_ENABLED=true
```

**Comportamento:**

- `OPS_SIDECAR_ENABLED=false` ou sem `POSTGRES_URL` → **igual hoje** (sem sidecar autofim Postgres).
- `true` → grava `ops_alarm_events` + enfileira autofim **depois** de encaminhar ao Xano (WhatsApp/ligação intactos).
- Erro/sem config Postgres no autofim → **fail-open** (não bloqueia).

```powershell
cd c:\sistemaconfmonit\core4\home\confmonit\v4.0\eventgateway
.\build.ps1

cd c:\sistemaconfmonit\core4\home\confmonit\v4.0\taskxano
.\build.ps1
# Reiniciar serviços no servidor
```

### taskxano `.env`

```env
POSTGRES_URL=postgres://confmonit:SENHA@HOST:5432/confmonit?sslmode=disable
# URLs autofim Xano (já existentes no .env de produção)
```

---

## 8 — Xano (opcional — alinhar online com local)

Pull já foi feito localmente. **Produção Xano só muda se você der push.**

### O que importa para autofim (evitar fila duplicada)

Push da função **#22** com bloco `bot_finalizaeventoauto` **comentado** (já está assim no local):

`functions/22_funcao_sistema_send_whats_event_central_alarm.xs`

### Push completo (se quiser sincronizar tudo)

```powershell
cd c:\sistemaconfmonit\core4
xano workspace push -b v1 -w 1
```

> Revise diff antes do push. Renomes de ID (520_, 524_, etc.) foram alinhados localmente.

### O que **não** precisa criar no Xano

- Tabelas `ops_*` → Postgres
- APIs Config Atendimento → Go
- Integração `#22` com `GET /ops/atendimento/resolve` (fail-open) — **implementada localmente**; requer push Xano + deploy ConfVision

---

## 9 — Rollback / deploy conservador

Se quiser **máxima segurança** na primeira subida:

1. Faça passos **1–6** (schema + confvision + franqueadopro).
2. Deixe `OPS_SIDECAR_ENABLED=false` no eventgateway.
3. **Não** dê push Xano ainda.
4. Teste telas Config Atendimento + bloqueio IA.
5. Depois: `OPS_SIDECAR_ENABLED=true` + push função #22 comentada + restart eventgateway/taskxano.

---

## 10 — Checklist final

- [ ] `008` aplicado no Postgres
- [ ] INSERTs MySQL colados no `009`
- [ ] `009` aplicado — clientes ativos com IA+autofim TRUE
- [ ] `POSTGRES_URL` no confvision, franqueadoadmin, eventgateway, taskxano
- [ ] Binários Go rebuild + restart serviços
- [ ] Tela Config Atendimento abre e salva
- [ ] Bloqueio IA lista/adiciona/remove (Postgres)
- [ ] Evento real: WhatsApp/ligação **continuam** (função #22 inalterada online)
- [ ] (Opcional) Sidecar autofim + push #22 comentada

---

## Arquivos principais (referência)

| Arquivo | Função |
|---------|--------|
| `confvision/sql/008_atendimento_schema.sql` | Schema Postgres |
| `confvision/sql/009_atendimento_liberacao_inicial.sql` | Migração clientes ativos |
| `home/confmonit/v4.0/api/sql/20260811_mysql_export_clientes_postgres.sql` | Export MySQL → INSERTs |
| `home/confmonit/v4.0/apifunction/sql/004_fp_catalogo_creditos_atendimento.sql` | Produtos crédito MySQL |
| `home/confmonit/v4.0/confvision/src/modulos/visdata/atendimento.go` | API `/ops/atendimento/*` |
| `home/confmonit/v4.0/franqueadopromais/recursos/modulos/atendimento/config-atendimento/` | UI Config |
| `home/confmonit/v4.0/eventgateway/postgres_sidecar.go` | Sidecar autofim |
| `functions/22_funcao_sistema_send_whats_event_central_alarm.xs` | Disparo WhatsApp (push autofim comentado) |

---

## Pendências (não bloqueiam este deploy)

- Integrar função #22 com `GET /ops/atendimento/resolve` (fail-open) — **feito no repo**; falta push Xano + binário ConfVision no core-4
- Compra créditos Meu plano + débito por canal
- Pipeline e-mail/SMTP
- Seeds Xano `atendimento.configuracao` em `fn_fp_plano_regras_default`
- Espelhar em `franqueadoadmin` (painel central, se necessário)
