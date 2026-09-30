-- Catalogo financeiro admConfmonit (PostgreSQL — substitui Xano fp_produto_catalogo / fp_pacote_cota)

CREATE TABLE IF NOT EXISTS fp_catalogo_produto (
    id                      SERIAL PRIMARY KEY,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_central              VARCHAR(64) NOT NULL DEFAULT '',
    id_representante        VARCHAR(64) NOT NULL DEFAULT '',
    produto                 VARCHAR(64) NOT NULL DEFAULT '',
    plano                   VARCHAR(64) NOT NULL DEFAULT '',
    nome_exibicao           VARCHAR(200) NOT NULL DEFAULT '',
    valor_mensal            NUMERIC(12, 2) NOT NULL DEFAULT 0,
    valor_piso_breakglass   NUMERIC(12, 2) NOT NULL DEFAULT 0,
    periodicidade_padrao    VARCHAR(20) NOT NULL DEFAULT 'mensal',
    retencao_dias           INT NOT NULL DEFAULT 30,
    limites_json            JSONB NOT NULL DEFAULT '{}',
    modulos_json            JSONB NOT NULL DEFAULT '{}',
    fp_pacote_cota_id       INT NULL,
    ativo                   CHAR(1) NOT NULL DEFAULT 'S',
    observacao              TEXT NOT NULL DEFAULT '',
    UNIQUE (id_central, id_representante, produto, plano)
);

CREATE INDEX IF NOT EXISTS idx_fp_cat_prod_central ON fp_catalogo_produto (id_central, ativo);
CREATE INDEX IF NOT EXISTS idx_fp_cat_prod_produto ON fp_catalogo_produto (produto, plano);

CREATE TABLE IF NOT EXISTS fp_pacote_cota (
    id                      SERIAL PRIMARY KEY,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_central              VARCHAR(64) NOT NULL DEFAULT '',
    nome                    VARCHAR(120) NOT NULL DEFAULT '',
    quantidade              INT NOT NULL DEFAULT 0,
    valor                   NUMERIC(12, 2) NOT NULL DEFAULT 0,
    valor_piso_breakglass   NUMERIC(12, 2) NOT NULL DEFAULT 0,
    ativo                   CHAR(1) NOT NULL DEFAULT 'S',
    observacao              TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_fp_pacote_cota_central ON fp_pacote_cota (id_central, ativo);

CREATE TABLE IF NOT EXISTS fp_central_preco_config (
    id                          SERIAL PRIMARY KEY,
    id_central                  VARCHAR(64) NOT NULL UNIQUE,
    modo_preco                  VARCHAR(20) NOT NULL DEFAULT 'livre',
    valor_unitario_minimo_cota  NUMERIC(12, 2) NOT NULL DEFAULT 0,
    observacao                  TEXT NOT NULL DEFAULT '',
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fp_preco_pacote_cota (
    id                  SERIAL PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_representante    VARCHAR(64) NOT NULL DEFAULT '',
    fp_pacote_cota_id   INT NOT NULL,
    valor_venda         NUMERIC(12, 2) NOT NULL DEFAULT 0,
    ativo               CHAR(1) NOT NULL DEFAULT 'S',
    observacao          TEXT NOT NULL DEFAULT '',
    atualizado_em       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (id_representante, fp_pacote_cota_id)
);

CREATE INDEX IF NOT EXISTS idx_fp_preco_pacote_cota_rep ON fp_preco_pacote_cota (id_representante, ativo);
