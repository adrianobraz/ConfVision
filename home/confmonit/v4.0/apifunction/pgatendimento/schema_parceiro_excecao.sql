-- Espelho local: ops_parceiro_excecao (ver confvision/sql/017_parceiro_excecao_pendente.sql)

CREATE TABLE IF NOT EXISTS ops_parceiro_excecao (
    id                  BIGSERIAL PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado       TEXT NOT NULL,
    id_cliente          TEXT NOT NULL,
    nome_cliente        TEXT NOT NULL DEFAULT '',
    id_parceiro         TEXT NOT NULL,
    id_parceiro_anterior TEXT NOT NULL DEFAULT '',
    id_vinculo_anterior  TEXT NOT NULL DEFAULT '',
    status              TEXT NOT NULL DEFAULT 'pendente',
    preco_unitario      NUMERIC(12, 2) NOT NULL DEFAULT 0,
    valor_total         NUMERIC(12, 2) NOT NULL DEFAULT 0,
    fp_fatura_id        INT NULL,
    ciclo_ref           TEXT NOT NULL DEFAULT '',
    periodo_inicio      DATE NULL,
    periodo_fim         DATE NULL,
    pago_ate            DATE NULL,
    id_vinculo_ativo    TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_ops_parceiro_exc_fra_status
    ON ops_parceiro_excecao (id_franqueado, status, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_parceiro_exc_pendente
    ON ops_parceiro_excecao (id_franqueado, id_cliente)
    WHERE status = 'pendente';
