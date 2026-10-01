-- Capacidade de processamento ConfVision (vagas de camera — pool unico)
-- Aplicar apos 003_extend_schema.sql

CREATE TABLE IF NOT EXISTS vis_capacidade_config (
    id_central          TEXT NOT NULL,
    id_representante    TEXT NOT NULL DEFAULT '',
    quantidade_minima   INT NOT NULL DEFAULT 10,
    preco_base_camera   NUMERIC(12, 2) NOT NULL DEFAULT 11.50,
    preco_venda_camera  NUMERIC(12, 2),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id_central, id_representante)
);

CREATE TABLE IF NOT EXISTS vis_capacidade_contrato (
    id                          SERIAL PRIMARY KEY,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado               TEXT NOT NULL,
    id_central                  TEXT NOT NULL,
    id_representante            TEXT DEFAULT '',
    quantidade_contratada       INT NOT NULL,
    quantidade_minima_snapshot  INT NOT NULL DEFAULT 0,
    preco_por_camera            NUMERIC(12, 2) NOT NULL,
    valor_mensal                NUMERIC(12, 2) NOT NULL,
    status                      TEXT NOT NULL DEFAULT 'pendente',
    pago_em                     TIMESTAMPTZ,
    valido_ate                  TIMESTAMPTZ,
    id_fatura                   TEXT,
    id_pagamento                TEXT,
    observacao                  TEXT
);

CREATE INDEX IF NOT EXISTS idx_vis_capacidade_contrato_franqueado
    ON vis_capacidade_contrato (id_franqueado, status);

CREATE INDEX IF NOT EXISTS idx_vis_capacidade_contrato_ativo
    ON vis_capacidade_contrato (id_franqueado)
    WHERE status = 'ativo';

-- Config padrao (fallback quando central especifica ainda nao cadastrada)
INSERT INTO vis_capacidade_config (id_central, id_representante, quantidade_minima, preco_base_camera)
VALUES ('*', '', 10, 11.50)
ON CONFLICT (id_central, id_representante) DO NOTHING;
