-- Zera dados financeiros espelhados no PostgreSQL (apifunction)
-- Executar APOS zerar tabelas fp_* no Xano. Nao remove estrutura (CREATE IF NOT EXISTS).
-- Uso: psql -f 010_zero_financeiro_postgres.sql

BEGIN;

-- Governanca cascata REP->Central
TRUNCATE TABLE fp_gov_suspensao_cascata RESTART IDENTITY CASCADE;
TRUNCATE TABLE fp_gov_restricao CASCADE;
TRUNCATE TABLE fp_gov_repasse CASCADE;

-- Cobranca consolidada
TRUNCATE TABLE fp_ordem_cobranca_log RESTART IDENTITY CASCADE;
TRUNCATE TABLE fp_servico_cobranca RESTART IDENTITY CASCADE;
TRUNCATE TABLE fp_cobranca_config CASCADE;

-- Catalogo admConfmonit (PostgreSQL nativo)
TRUNCATE TABLE fp_catalogo_produto RESTART IDENTITY CASCADE;
TRUNCATE TABLE fp_pacote_cota RESTART IDENTITY CASCADE;
TRUNCATE TABLE fp_central_preco_config RESTART IDENTITY CASCADE;

-- Espelho financeiro Xano (dashboard KPI)
TRUNCATE TABLE fp_fin_shadow_diff RESTART IDENTITY CASCADE;
TRUNCATE TABLE fp_fin_sync_state CASCADE;
TRUNCATE TABLE fp_fin_conta_pagar CASCADE;
TRUNCATE TABLE fp_fin_caixa_movimento CASCADE;
TRUNCATE TABLE fp_fin_pagamento CASCADE;
TRUNCATE TABLE fp_fin_assinatura CASCADE;
TRUNCATE TABLE fp_fin_fatura CASCADE;

-- Receptor DNS (se existir no mesmo banco)
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'fp_receptor_dns') THEN
    EXECUTE 'TRUNCATE TABLE fp_receptor_dns RESTART IDENTITY CASCADE';
  END IF;
END $$;

COMMIT;
