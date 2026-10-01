-- =============================================================================
-- LIBERACAO INICIAL — PostgreSQL APENAS (pgAdmin, psql, DBeaver Postgres)
-- NÃO rodar no MySQL / phpMyAdmin
-- =============================================================================
-- Ordem:
--   1) Aplicar 008_atendimento_schema.sql
--   2) No MySQL rodar: home/confmonit/v4.0/api/sql/20260811_mysql_export_clientes_postgres.sql
--   3) Copiar coluna sql_postgres e colar abaixo (entre marcos PASTE)
--   4) Executar este arquivo inteiro no PostgreSQL
-- =============================================================================

BEGIN;

-- Staging (tabela permanente para migracao)
CREATE TABLE IF NOT EXISTS ops_stg_cliente_ativo (
    id_franqueado TEXT NOT NULL,
    id_cliente    TEXT NOT NULL,
    nome_cliente  TEXT DEFAULT '',
    PRIMARY KEY (id_franqueado, id_cliente)
);

TRUNCATE ops_stg_cliente_ativo;

-- --- PASTE: cole aqui os INSERT gerados pelo MySQL ---
-- INSERT INTO ops_stg_cliente_ativo (...) VALUES (...) ON CONFLICT DO NOTHING;
-- --- FIM PASTE ---

-- Fallback: clientes que já tiveram evento no Postgres (se MySQL ainda não foi colado)
INSERT INTO ops_stg_cliente_ativo (id_franqueado, id_cliente)
SELECT DISTINCT id_franqueado, id_cliente
FROM ops_alarm_events
WHERE id_franqueado IS NOT NULL AND id_franqueado <> ''
  AND id_cliente    IS NOT NULL AND id_cliente    <> ''
ON CONFLICT DO NOTHING;

-- Cria config (false) para cada cliente da staging
INSERT INTO ops_cliente_atendimento_config (
    id_franqueado, id_cliente,
    inteligencia_artificial, finalizacao_automatica,
    parceiro_monitoramento, email_ativo,
    cobertura_horaria, ativo, updated_at
)
SELECT
    t.id_franqueado, t.id_cliente,
    FALSE, FALSE, FALSE, FALSE,
    '24h', TRUE, NOW()
FROM ops_stg_cliente_ativo t
ON CONFLICT (id_franqueado, id_cliente) DO NOTHING;

-- Liga IA + autofim; parceiro e e-mail ficam desligados
UPDATE ops_cliente_atendimento_config c
SET
    inteligencia_artificial = TRUE,
    finalizacao_automatica  = TRUE,
    parceiro_monitoramento  = FALSE,
    email_ativo             = FALSE,
    updated_at              = NOW()
FROM ops_stg_cliente_ativo t
WHERE c.id_franqueado = t.id_franqueado
  AND c.id_cliente    = t.id_cliente;

-- Politica do franqueado
INSERT INTO ops_franqueado_atendimento_politica (
    id_franqueado,
    inteligencia_artificial, ia_modo,
    finalizacao_automatica, autofim_modo,
    parceiro_monitoramento, parceiro_modo,
    email_ativo, email_modo,
    cobertura_horaria, updated_at
)
SELECT DISTINCT
    t.id_franqueado,
    TRUE, 'todos',
    TRUE, 'todos',
    FALSE, 'todos',
    FALSE, 'todos',
    '24h', NOW()
FROM ops_stg_cliente_ativo t
ON CONFLICT (id_franqueado) DO UPDATE SET
    inteligencia_artificial = TRUE,
    ia_modo                 = 'todos',
    finalizacao_automatica  = TRUE,
    autofim_modo            = 'todos',
    parceiro_monitoramento  = FALSE,
    email_ativo             = FALSE,
    updated_at              = NOW();

COMMIT;

-- Conferencia:
-- SELECT COUNT(*) FROM ops_stg_cliente_ativo;
-- SELECT COUNT(*) total,
--        SUM(inteligencia_artificial::int) com_ia,
--        SUM(finalizacao_automatica::int) com_autofim
-- FROM ops_cliente_atendimento_config;
