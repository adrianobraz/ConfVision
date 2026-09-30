# Deploy — Serviços / Créditos prepago

## 1. PostgreSQL (ConfVision)

```bash
POSTGRES_URL=... go run ./cmd/apply ../../009_tarifa_operacional.sql
# Go-live UMA vez — bonifica R$ 9999 ligação (franqueados existentes):
POSTGRES_URL=... go run ./cmd/apply ../../010_bonificacao_go_live.sql
# Saldo unico (consolida carteiras por canal em canal=credito):
POSTGRES_URL=... go run ./cmd/apply ../../011_saldo_credito_unico.sql
```

## 2. MySQL (apifunction)

```bash
mysql -u confmonit -p confmonitV4 < home/confmonit/v4.0/apifunction/sql/005_servicos_credito_schema.sql
# Se 005 ja foi aplicado antes do saldo unico:
mysql -u confmonit -p confmonitV4 < home/confmonit/v4.0/apifunction/sql/006_credito_recarga_canal.sql
```

## 3. Xano (push)

- `functions/531_fn_fp_credito_recarga_sync.xs`
- `apis/fp_franq_financeiro/fp_credito_recarga_sync_POST.xs`
- `functions/22_funcao_sistema_send_whats_event_central_alarm.xs` (bloqueio ligação por saldo)

## 4. Binários

- **apifunction** — rotas `/servico/tarifa/*`, `/credito/recarga/*`
- **confvision** — ops crédito/tarifa
- **franqueadopro** — `/atendCreditosRecarga`, config-atendimento UI

Env franqueadopro:

```
APIFUNCTION_URL=http://127.0.0.1:20001
WORKER_SECRET=...
```

Env apifunction:

```
XANO_API_FINANCEIRO=...
WORKER_SECRET=...
CONFVISION_OPS_URL=http://127.0.0.1:8086
```

## 5. admConfmonit

Rebuild → menu **Serviços → Créditos prepago**

- Tarifas operacionais (CEN piso / REP preço)
- Recarga para franqueado (UsaAdmConfmonit)
- Baixa fatura tipo `recarga_credito_servicos` → credita Postgres automaticamente

## Regras

| Item | Regra |
|------|--------|
| Todos prepago | Sem saldo → bloqueia ligação (#22) |
| Saldo | **Único** por franqueado (`canal=credito`); tarifas por serviço |
| Antigos go-live | Script 010 → R$ 9999 bonificação (não contabiliza) |
| Novos franqueados | Saldo zero até recarga |
| Recarga mínima | R$ 50 = R$ 50 crédito |
| Bonificação | `tipo=bonificacao`, sem fatura |
