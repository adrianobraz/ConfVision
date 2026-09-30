# Deploy — Módulo Contrato (admConfmonit + apifunction MySQL)

## 1. MySQL (confmonitV4)

```bash
mysql -u confmonit -p confmonitV4 < home/confmonit/v4.0/apifunction/sql/transferencia_vinculo_log.sql
mysql -u confmonit -p confmonitV4 < home/confmonit/v4.0/apifunction/sql/001_fp_contrato_schema.sql
mysql -u confmonit -p confmonitV4 < home/confmonit/v4.0/apifunction/sql/002_fp_fatura_contabil.sql
mysql -u confmonit -p confmonitV4 < home/confmonit/v4.0/apifunction/sql/003_fp_fatura_estornada.sql
```

## 2. apifunction (Go)

```powershell
cd home/confmonit/v4.0/apifunction
.\build.ps1
```

Copiar `deploy/apifunction/` para a VPS. No `.env`:

- `API_KEY` — mesma da API V4
- `BD_*` — MySQL confmonitV4
- `WORKER_SECRET` — chave do fp-billing-worker
- `XANO_API_FINANCEIRO` — URL base API fp_franqFinanceiro (espelha faturas em fp_fatura)
- `APIFUNCTION_PORT=20001`

Reiniciar serviço: `confmonit4apifunction.service`

### Confirmar binário atualizado na VPS

**1. Health com versão** (binário antigo só retorna `servico`; novo traz `versao` e `auth`):

```bash
curl -s http://185.130.61.4:20001/health
```

Esperado após deploy:

```json
{
  "ok": true,
  "servico": "apifunction",
  "versao": "2026.08.06-contrato-contabil",
  "auth": {
    "jwt_bearer": true,
    "adm_session_headers": true,
    "breakglass": true
  }
}
```

Se **não** aparecer `versao` / `auth.adm_session_headers`, o binário **não foi atualizado**.

**2. Teste auth admConfmonit** (sem JWT — só headers X-Adm):

```bash
curl -s -X POST http://185.130.61.4:20001/contrato/preview \
  -H "Content-Type: application/json" \
  -H "X-Adm-Token: teste" \
  -H "X-Adm-Central: qualquer-central" \
  -H "X-Adm-Tipo: CEN" \
  -d "{\"itens\":[]}"
```

| Resposta | Significado |
|----------|-------------|
| `authorization obrigatorio` | Binário **antigo** — refazer deploy |
| `{"ok":true,"dados":{...}}` | Binário **novo** — auth OK |

**3. Reiniciar na VPS:**

```bash
sudo systemctl restart confmonit4apifunction
sudo systemctl status confmonit4apifunction
curl -s http://127.0.0.1:20001/health
```

## 3. admConfmonit (React)

```powershell
cd admConfmonit
npm install
npm run build
```

Publicar `dist/` (ou pasta deploy habitual).

`.env` produção: `VITE_API_FUNCTION_URL=http://<host>:20001`

Menu: **Gestão de Caixa → Contrato franqueado**

Na primeira vez: botão **Seed catálogo** (produtos, cotas, licenças CV).

## 4. FranqueadoPro

Rebuild franqueadoadmin com variável:

- `APIFUNCTION_URL=http://127.0.0.1:20001` (ou IP interno do apifunction)

O middleware de licença consulta `/contrato/efetivo` antes do Xano legado.

## 5. fp-billing-worker

`.env`:

- `APIFUNCTION_URL=http://127.0.0.1:20001`
- `WORKER_SECRET` — igual apifunction
- `XANO_API_FINANCEIRO` — opcional (legado)

```powershell
cd home/confmonit/v4.0/fp-billing-worker
go build -o fp-billing-worker .
```

## Fluxo operacional

1. Central ou REP (conforme `UsaAdmConfmonit`) monta contrato
2. **Salvar e gerar fatura** → status `aguardando_pagamento` → franqueado **travado**
3. **Confirmar pago** na lista de faturas → status `ativo` → franqueado **liberado**
4. **Estornar pagamento** (fatura paga) → fatura `estornada`, contrato `suspenso`, nova fatura `aberta` → franqueado **bloqueado**
5. Worker suspende se fatura vencer; gera renovações conforme periodicidade
6. Faturas espelhadas em **fp_fatura** (Xano) — visíveis em Balanço/Faturamento *(opcional; omitir `XANO_API_FINANCEIRO` para operar só no MySQL)*

## Contabilidade (Xano)

Com `XANO_API_FINANCEIRO` configurado no apifunction:

| Evento módulo Contrato | Espelho Xano |
|------------------------|--------------|
| Gerar fatura | `fp_fatura` tipo `contrato_franqueado` + itens |
| Confirmar pagamento | `fp_pagamento` + repasse + log financeiro |
| Estornar pagamento | MySQL only (`POST /contrato/fatura/estornar-pagamento`); log em `fp_contrato_log` |
| Excluir/cancelar fatura | `fp_fatura.status = cancelada` |

Push Xano (functions 522–524, APIs 2391–2393) antes de ativar em produção.

## Regra UsaAdmConfmonit

| Flag | Quem edita contrato |
|------|---------------------|
| N | Central |
| S | Representante (carteira) |
