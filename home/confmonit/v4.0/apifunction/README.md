# apifunction

Microservico Go para **transferencia de vinculo** (CLI/FRA/REP) usado pelo admConfmonit.

- Nao altera a API V4 (`api/src/V4`)
- MySQL `confmonitV4` (mesmo banco)
- Auth: JWT do `POST /v4/admfinanceiro/logar` (mesma `API_KEY`)
- Log: tabela `transferencia_vinculo_log`

## Regras de bloqueio (preview)

- Eventos > 100 (MySQL `evento`)
- Qualquer fatura MySQL (`faturas`)
- Processo aberto (`processo.DataAtenFim IS NULL`)
- Tickets abertos
- Entidade bloqueada/cancelada
- Franqueado: `UsaConfVision=S`, SMS em `listaEnvio`, master `AdmFinanceiro=S`
- Representante: `UsaAdmConfmonit=S`

## Build

```powershell
cd home/confmonit/v4.0/apifunction
.\build.ps1
```

## SQL

```bash
mysql -u confmonit -p confmonitV4 < sql/transferencia_vinculo_log.sql
```

## admConfmonit (React)

```env
VITE_API_FUNCTION_URL=http://10.2.2.1:20001
```

Chamar com header `Authorization: Bearer <token>` retornado no login admfinanceiro.
