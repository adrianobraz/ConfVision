-- Tarifas operacionais (Postgres) — espelho efetivo para bloqueio/debito na #22
CREATE TABLE IF NOT EXISTS ops_tarifa_operacional (
    id_central      TEXT NOT NULL,
    canal           TEXT NOT NULL,
    valor_tentativa NUMERIC(12, 4) NOT NULL DEFAULT 0,
    valor_minuto    NUMERIC(12, 4) NOT NULL DEFAULT 0,
    valor_unidade   NUMERIC(12, 4) NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id_central, canal)
);

-- Controle go-live bonificacao (nao re-executar)
CREATE TABLE IF NOT EXISTS ops_credito_go_live (
    id              SERIAL PRIMARY KEY,
    executado_em    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    qtd_franqueados INT NOT NULL DEFAULT 0,
    valor_bonificado NUMERIC(14, 4) NOT NULL DEFAULT 9999,
    observacao      TEXT DEFAULT ''
);
