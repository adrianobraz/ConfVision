# Deploy Xano — hierarquia REP (checklist manual)

**Regra:** o número no nome do arquivo = **ID online** no Xano (quando existir).  
Abrir arquivo: **Ctrl+P** e colar o nome (ex.: `2366_fp_preco_rep`).

---

## 1) Tabelas NOVAS (criar)

| ID | Arquivo | Nome Xano |
|----|---------|-----------|
| — | `tables/135_fp_admin_sessao.xs` | fp_admin_sessao |
| — | `tables/136_fp_preco_representante.xs` | fp_preco_representante |

## 2) Tabelas ALTERADAS (colar schema)

| Arquivo | Nome Xano |
|---------|-----------|
| `tables/120_fp_assinatura_produto.xs` | fp_assinatura_produto |
| `tables/121_fp_fatura.xs` | fp_fatura |
| `tables/131_fp_cupom_desconto.xs` | fp_cupom_desconto |

---

## 3) Functions NOVAS (criar)

| Arquivo | Nome |
|---------|------|
| `functions/479_fn_fp_franqueado_id_representante.xs` | fn_fp_franqueado_id_representante |
| `functions/480_fn_fp_preco_efetivo.xs` | fn_fp_preco_efetivo |
| `functions/481_fn_fp_preco_rep_salvar.xs` | fn_fp_preco_rep_salvar |
| `functions/482_fn_fp_catalogo_com_preco_rep.xs` | fn_fp_catalogo_com_preco_rep |
| `functions/483_fn_fp_repasse_gerar_apos_pagamento.xs` | fn_fp_repasse_gerar_apos_pagamento |
| `functions/484_fn_fp_franqueados_ids_representante.xs` | fn_fp_franqueados_ids_representante |
| `functions/485_fn_fp_admin_assert_escopo_franqueado.xs` | fn_fp_admin_assert_escopo_franqueado |
| `functions/486_fn_fp_repasse_aplicar_cupom.xs` | fn_fp_repasse_aplicar_cupom |
| `functions/487_fn_fp_fatura_na_carteira.xs` | fn_fp_fatura_na_carteira |
| `functions/488_fn_fp_assinatura_na_carteira.xs` | fn_fp_assinatura_na_carteira |

## 4) Functions ALTERADAS (colar)

| Arquivo | Nome |
|---------|------|
| `functions/433_fn_fp_admin_validar.xs` | fn_fp_admin_validar |
| `functions/434_fn_fp_admin_login.xs` | fn_fp_admin_login |
| `functions/435_fn_fp_franqueado_buscar.xs` | fn_fp_franqueado_buscar |
| `functions/442_fn_fp_fatura_gerar.xs` | fn_fp_fatura_gerar |
| `functions/446_fn_fp_assinatura_contratar.xs` | fn_fp_assinatura_contratar |
| `functions/465_fn_fp_cupom_validar.xs` | fn_fp_cupom_validar |
| `functions/466_fn_fp_cupom_aplicar.xs` | fn_fp_cupom_aplicar |
| `functions/467_fn_fp_cupom_criar.xs` | fn_fp_cupom_criar |
| `functions/471_fn_fp_fin_resumo.xs` | fn_fp_fin_resumo |
| `functions/473_fn_fp_fin_descontos_resumo.xs` | fn_fp_fin_descontos_resumo |

---

## 5) APIs — grupo `fp_franqFinanceiro` (ALTERAR pelo ID online)

| ID | Arquivo | Query |
|----|---------|-------|
| 2281 | `apis/fp_franq_financeiro/2281_fp_fin_dashboard_POST.xs` | fp_fin_dashboard |
| 2287 | `apis/fp_franq_financeiro/2287_fp_admin_login_POST.xs` | fp_admin_login |
| 2288 | `apis/fp_franq_financeiro/2288_fp_franqueado_buscar_POST.xs` | fp_franqueado_buscar |
| 2260 | `apis/fp_franq_financeiro/2260_fp_catalogo_listar_POST.xs` | fp_catalogo_listar |
| 2262 | `apis/fp_franq_financeiro/2262_fp_catalogo_salvar_POST.xs` | fp_catalogo_salvar |
| 2263 | `apis/fp_franq_financeiro/2263_fp_assinatura_salvar_POST.xs` | fp_assinatura_salvar |
| 2265 | `apis/fp_franq_financeiro/2265_fp_assinatura_listar_POST.xs` | fp_assinatura_listar |
| 2266 | `apis/fp_franq_financeiro/2266_fp_assinatura_liberar_manual_POST.xs` | fp_assinatura_liberar_manual |
| 2267 | `apis/fp_franq_financeiro/2267_fp_assinatura_suspender_POST.xs` | fp_assinatura_suspender |
| 2270 | `apis/fp_franq_financeiro/2270_fp_fatura_listar_POST.xs` | fp_fatura_listar |
| 2272 | `apis/fp_franq_financeiro/2272_fp_fatura_registrar_pagamento_POST.xs` | fp_fatura_registrar_pagamento |
| 2319 | `apis/fp_franq_financeiro/2319_fp_cupom_listar_POST.xs` | fp_cupom_listar |
| 2320 | `apis/fp_franq_financeiro/2320_fp_cupom_criar_POST.xs` | fp_cupom_criar |
| 2331 | `apis/fp_franq_financeiro/2331_fp_fin_inadimplentes_listar_POST.xs` | fp_fin_inadimplentes_listar |
| 2332 | `apis/fp_franq_financeiro/2332_fp_fin_demonstrativo_POST.xs` | fp_fin_demonstrativo |
| 2333 | `apis/fp_franq_financeiro/2333_fp_fin_fluxo_caixa_POST.xs` | fp_fin_fluxo_caixa |
| 2335 | `apis/fp_franq_financeiro/2335_fp_fin_prestacao_contas_POST.xs` | fp_fin_prestacao_contas |
| 2337 | `apis/fp_franq_financeiro/2337_fp_fin_fechamento_preview_POST.xs` | fp_fin_fechamento_preview |
| 2338 | `apis/fp_franq_financeiro/2338_fp_fin_fechamento_salvar_POST.xs` | fp_fin_fechamento_salvar |
| 2342 | `apis/fp_franq_financeiro/2342_fp_fin_descontos_listar_POST.xs` | fp_fin_descontos_listar |

## 6) APIs NOVAS — grupo `fp_franqFinanceiro` (CRIAR; IDs reservados 2366–2369)

| ID | Arquivo | Query |
|----|---------|-------|
| **2366** | `apis/fp_franq_financeiro/2366_fp_preco_rep_listar_POST.xs` | fp_preco_rep_listar |
| **2367** | `apis/fp_franq_financeiro/2367_fp_preco_rep_salvar_POST.xs` | fp_preco_rep_salvar |
| **2368** | `apis/fp_franq_financeiro/2368_fp_repasse_listar_POST.xs` | fp_repasse_listar |
| **2369** | `apis/fp_franq_financeiro/2369_fp_repasse_aplicar_cupom_POST.xs` | fp_repasse_aplicar_cupom |

> **Não confundir** com IDs 2347–2350 do grupo `relatoriocentraldisparos` (Central de Disparos).

---

## 7) APIs — grupo `fp_franqueadoPro` (ALTERAR)

| ID | Arquivo | Query |
|----|---------|-------|
| 2312 | `apis/fp_franqueado_pro/2312_fp_catalogo_listar_publico_POST.xs` | fp_catalogo_listar_publico |
| 2322 | `apis/fp_franqueado_pro/2322_fp_cupom_validar_POST.xs` | fp_cupom_validar |
| 2343 | `apis/fp_franqueado_pro/2343_fp_ecossistema_listar_POST.xs` | fp_ecossistema_listar |

---

## 8) ConfVision (ALTERAR — arquivo já alinhado ao ID online)

| ID | Arquivo | Query |
|----|---------|-------|
| 2364 | `apis/conf_vision/2364_vis_evento_by_cliente_page_GET.xs` | vis_evento_by_cliente_page |
| 2365 | `apis/conf_vision/2365_vis_camera_gravacao_flush_vis_camera_id_POST.xs` | vis_camera/gravacao/flush |

---

## 9) Central de Disparos (NÃO misturar com financeiro)

| ID | Arquivo | Query |
|----|---------|-------|
| 2347 | `apis/relatoriocentraldisparos/2347_fp_cd_eventos_pendentes_POST.xs` | fp_cd_eventos_pendentes |
| 2348 | `apis/relatoriocentraldisparos/2348_fp_cd_finalizados_robo_POST.xs` | fp_cd_finalizados_robo |
| 2349 | `apis/relatoriocentraldisparos/2349_fp_cd_finalizados_bot_POST.xs` | fp_cd_finalizados_bot |
| 2350 | `apis/relatoriocentraldisparos/2350_fp_cd_ligacao_historico_POST.xs` | fp_cd_ligacao_historico |

---

## 10) Fora do Xano

1. SQL: `home/confmonit/v4.0/api/sql/20260718_usuarios_adm_financeiro.sql`
2. API Go: `home/confmonit/v4.0/api/src/V4/modulos/admfinanceiro/`
3. Front: `admConfmonit/` (rebuild)

---

## Ordem recomendada

1. Tabelas → 2. Functions → 3. APIs alterar → 4. APIs criar (2366–2369) → 5. Testar login CEN/REP
