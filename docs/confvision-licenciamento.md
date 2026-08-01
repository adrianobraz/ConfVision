# ConfVision — Licenciamento e faturamento

Especificação operacional do fluxo de licenças ConfVision integrado ao FranqueadoPro e admConfmonit.

## Regras de acesso (`UsaConfVision = "S"`)

O franqueado tem acesso ao ConfVision quando **qualquer** condição for verdadeira:

1. Assinatura **FranqueadoPro Pro+** ativa (`status = ativa`, `valido_ate > now`)
2. Pelo menos uma `vis_licenca` **paga e válida**:
   - `status` em `disponivel` ou `em_uso`
   - `pago_em` preenchido
   - `valido_ate` nulo ou `valido_ate > now`

Licenças `pendente` **não** liberam acesso.

## Ciclo de vida da licença

```
pendente → (pagamento manual) → disponivel → (vincula câmera) → em_uso → (vencimento) → expirada
```

- **`valido_ate`**: eixo do ciclo mensal; worker gera fatura ~5 dias antes
- **`pago_em`**: preenchido somente após confirmação manual em Financeiro
- **Desconto 20%**: franqueados com Pro+ ativo (`fn_fp_confvision_valor_com_desconto`)

## APIs Xano — principais

| Endpoint | Grupo | Uso |
|---|---|---|
| `fp_confvision_planos_listar` | fp_franqFinanceiro | admConfmonit — catálogo 18 planos |
| `fp_confvision_fatura_venda` | fp_franqFinanceiro | admConfmonit — venda com fatura aberta |
| `fp_confvision_comprar` | fp_franqueadoPro | Self-service ConfVision |
| `fp_confvision_resumo_franqueado` | fp_franqueadoPro | Resumo licenças + faturas abertas |
| `fp_confvision_planos_listar_franqueado` | fp_franqueadoPro | Catálogo self-service |
| `fp_fatura_registrar_pagamento` | fp_franqFinanceiro | Ativa licenças + sync MySQL |
| `fp_vis_licenca_listar_pendentes_renovacao` | fp_franqFinanceiro | Worker — anti-duplicata por `ciclo_ref` |
| `fp_fatura_gerar_automatica_confvision` | fp_franqFinanceiro | Worker — renovação com desconto Pro+ |

## Functions Xano

| Arquivo | Função |
|---|---|
| `447_fn_fp_confvision_valor_com_desconto.xs` | Desconto 20% Pro+ |
| `448_fn_fp_confvision_acesso_efetivo.xs` | Decide `UsaConfVision` S/N |
| `449_fn_franqueado_sync_usa_confvision.xs` | POST API Go legada |
| `450_fn_fp_confvision_planos_listar.xs` | 18 planos com preços |
| `451_fn_fp_confvision_ativar_licenca_pagamento.xs` | Ativa/renova após pagamento |
| `452_fn_fp_confvision_fatura_venda.xs` | Venda: licença pendente + fatura |

## API Go legada

**`POST /v4/franqueado/setUsaConfVision`**

```json
{ "fraId": "...", "usaConfVision": "S" }
```

Arquivos: `mFranqueado.go`, `cFranqueado.go`, `rFranqueado.go`

Configurável via `fp_config_financeiro.api_legada_url` (seed em `fn_fp_catalogo_seed`).

## fp-billing-worker

Ordem semanal:

1. `fp_fatura_suspender_vencidas` — suspende Pro+ inadimplente; marca renovação CV em atraso
2. `fp_vis_licenca_expirar_vencidas` — expira licenças + sync MySQL
3. `fp_assinatura_listar_pendentes_fatura` → `fp_fatura_gerar_automatica`
4. `fp_vis_licenca_listar_pendentes_renovacao` → `fp_fatura_gerar_automatica_confvision`

O worker repassa `ciclo_ref` da listagem (formato `VIS-{id}-{Ymd}`) para evitar faturas duplicadas.

## ConfVision self-service

Rota: `/minhas-licencas` — proxy para APIs `fp_franqueadoPro`.

Variável de ambiente: `XANO_API_FRANQUEADO_PRO` em `confvision/.env`.

## admConfmonit (UI externa)

Ao vender ConfVision:

1. `fp_confvision_planos_listar` — exibir 18 planos
2. `fp_confvision_fatura_venda` — gerar fatura aberta (licenças `pendente`)
3. Financeiro confirma pagamento → liberação automática

**Não usar** `fp_confvision_gerar_lote` no fluxo comercial (deprecated; delega para fatura_venda).

## Checklist de deploy

- [ ] Push Xano (`push_all_changes_to_xano`)
- [ ] Rodar `fn_fp_catalogo_seed` (config `api_legada_url`)
- [ ] Deploy API Go com `setUsaConfVision`
- [ ] Deploy ConfVision com `XANO_API_FRANQUEADO_PRO`
- [ ] Testar worker com `DRY_RUN=true`
- [ ] Teste E2E: venda → pagamento → sync → login CV
