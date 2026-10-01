-- Histórico de trocas de parceiro por cliente (Postgres confmonit)
-- Aplicar: POSTGRES_URL=... psql -f 018_parceiro_historico.sql

CREATE TABLE IF NOT EXISTS ops_parceiro_historico (
    id               BIGSERIAL PRIMARY KEY,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado    TEXT NOT NULL,
    id_cliente       TEXT NOT NULL DEFAULT '',
    nome_cliente     TEXT NOT NULL DEFAULT '',
    evento           TEXT NOT NULL,
    id_parceiro_de   TEXT NOT NULL DEFAULT '',
    id_parceiro_para TEXT NOT NULL DEFAULT '',
    id_vinculo_de    TEXT NOT NULL DEFAULT '',
    id_vinculo_para  TEXT NOT NULL DEFAULT '',
    fp_fatura_id     INT NULL,
    detalhe          JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_ops_parceiro_hist_fra
    ON ops_parceiro_historico (id_franqueado, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ops_parceiro_hist_cliente
    ON ops_parceiro_historico (id_franqueado, id_cliente, created_at DESC);
