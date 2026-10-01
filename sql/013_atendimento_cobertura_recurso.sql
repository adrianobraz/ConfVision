-- Cobertura horaria por recurso (IA e Parceiro) + log envio parceiro
-- Aplicar: POSTGRES_URL=... go run ./cmd/apply ../../013_atendimento_cobertura_recurso.sql

ALTER TABLE ops_franqueado_atendimento_politica
    ADD COLUMN IF NOT EXISTS ia_cobertura_horaria TEXT NOT NULL DEFAULT '24h',
    ADD COLUMN IF NOT EXISTS parceiro_cobertura_horaria TEXT NOT NULL DEFAULT '24h';

UPDATE ops_franqueado_atendimento_politica
SET ia_cobertura_horaria = COALESCE(NULLIF(TRIM(cobertura_horaria), ''), '24h'),
    parceiro_cobertura_horaria = COALESCE(NULLIF(TRIM(cobertura_horaria), ''), '24h')
WHERE ia_cobertura_horaria = '24h'
  AND parceiro_cobertura_horaria = '24h'
  AND cobertura_horaria IS NOT NULL
  AND TRIM(cobertura_horaria) <> ''
  AND TRIM(cobertura_horaria) <> '24h';

CREATE TABLE IF NOT EXISTS ops_parceiro_envio_log (
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado   TEXT NOT NULL,
    id_cliente      TEXT NOT NULL,
    id_processo     TEXT NOT NULL DEFAULT '',
    alarm_events_id BIGINT,
    id_parceiro     TEXT DEFAULT '',
    cti_grupo       TEXT DEFAULT '',
    acao_final      TEXT DEFAULT '',
    detalhe         JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_ops_parceiro_envio_fra_data
    ON ops_parceiro_envio_log (id_franqueado, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ops_parceiro_envio_processo
    ON ops_parceiro_envio_log (id_processo);
