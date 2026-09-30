-- Embedded migration (apifunction startup) — espelha confvision/sql/016_parceiro_dupla_cobranca.sql

ALTER TABLE ops_franqueado_atendimento_politica
    ADD COLUMN IF NOT EXISTS parceiro_dupla_comunicacao BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS ops_parceiro_ativacao (
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado   TEXT NOT NULL,
    id_parceiro     TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pendente',
    qtd_clientes    INT NOT NULL DEFAULT 0,
    preco_unitario  NUMERIC(12, 2) NOT NULL DEFAULT 0,
    valor_total     NUMERIC(12, 2) NOT NULL DEFAULT 0,
    fp_fatura_id    INT NULL,
    ciclo_ref       TEXT NOT NULL DEFAULT '',
    periodo_inicio  DATE NULL,
    periodo_fim     DATE NULL,
    pago_ate        DATE NULL,
    detalhe         JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_ops_parceiro_ativ_fra_status
    ON ops_parceiro_ativacao (id_franqueado, status, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_parceiro_ativ_fra_pendente
    ON ops_parceiro_ativacao (id_franqueado)
    WHERE status = 'pendente';

CREATE TABLE IF NOT EXISTS ops_parceiro_excecao (
    id                   BIGSERIAL PRIMARY KEY,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado        TEXT NOT NULL,
    id_cliente           TEXT NOT NULL,
    nome_cliente         TEXT NOT NULL DEFAULT '',
    id_parceiro          TEXT NOT NULL,
    id_parceiro_anterior TEXT NOT NULL DEFAULT '',
    id_vinculo_anterior  TEXT NOT NULL DEFAULT '',
    status               TEXT NOT NULL DEFAULT 'pendente',
    preco_unitario       NUMERIC(12, 2) NOT NULL DEFAULT 0,
    valor_total          NUMERIC(12, 2) NOT NULL DEFAULT 0,
    fp_fatura_id         INT NULL,
    ciclo_ref            TEXT NOT NULL DEFAULT '',
    periodo_inicio       DATE NULL,
    periodo_fim          DATE NULL,
    pago_ate             DATE NULL,
    id_vinculo_ativo     TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_ops_parceiro_exc_fra_status
    ON ops_parceiro_excecao (id_franqueado, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ops_parceiro_exc_cliente
    ON ops_parceiro_excecao (id_franqueado, id_cliente, status);

CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_parceiro_exc_pendente
    ON ops_parceiro_excecao (id_franqueado, id_cliente)
    WHERE status = 'pendente';

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
