-- Cobranca consolidada: conta corrente por servico + ordens (Postgres apifunction)

CREATE TABLE IF NOT EXISTS fp_cobranca_config (
  id_franqueado     VARCHAR(64) PRIMARY KEY,
  modo              VARCHAR(20) NOT NULL DEFAULT 'consolidado',
  dia_mensal        SMALLINT NOT NULL DEFAULT 29,
  dia_quinzenal_1   SMALLINT NOT NULL DEFAULT 14,
  dia_quinzenal_2   SMALLINT NOT NULL DEFAULT 29,
  id_central        VARCHAR(64) NOT NULL DEFAULT '',
  id_representante  VARCHAR(64) NOT NULL DEFAULT '',
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fp_servico_cobranca (
  id                BIGSERIAL PRIMARY KEY,
  id_franqueado     VARCHAR(64) NOT NULL,
  ref_tipo          VARCHAR(64) NOT NULL,
  ref_id            VARCHAR(128) NOT NULL,
  descricao         VARCHAR(500) NOT NULL DEFAULT '',
  data_inicio       DATE NOT NULL,
  periodicidade     VARCHAR(20) NOT NULL DEFAULT 'mensal',
  valor_ciclo       NUMERIC(12,2) NOT NULL DEFAULT 0,
  dias_ciclo        SMALLINT NOT NULL DEFAULT 15,
  pago_ate          DATE NULL,
  saldo_ajuste      NUMERIC(12,2) NOT NULL DEFAULT 0,
  status_ciclo      VARCHAR(20) NOT NULL DEFAULT 'em_ajuste',
  valor_piso        NUMERIC(12,2) NOT NULL DEFAULT 0,
  margem_central    NUMERIC(12,2) NOT NULL DEFAULT 0,
  margem_rep        NUMERIC(12,2) NOT NULL DEFAULT 0,
  ativo             BOOLEAN NOT NULL DEFAULT TRUE,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (id_franqueado, ref_tipo, ref_id)
);

CREATE INDEX IF NOT EXISTS idx_fp_servico_cob_fra ON fp_servico_cobranca (id_franqueado, ativo);

CREATE TABLE IF NOT EXISTS fp_ordem_cobranca_log (
  id                BIGSERIAL PRIMARY KEY,
  id_franqueado     VARCHAR(64) NOT NULL,
  ciclo_ref         VARCHAR(64) NOT NULL,
  vencimento        DATE NOT NULL,
  valor_total       NUMERIC(12,2) NOT NULL DEFAULT 0,
  valor_ajustes     NUMERIC(12,2) NOT NULL DEFAULT 0,
  qtd_itens         INT NOT NULL DEFAULT 0,
  fp_fatura_id      INT NULL,
  status            VARCHAR(20) NOT NULL DEFAULT 'gerada',
  detalhe           JSONB NOT NULL DEFAULT '{}',
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (id_franqueado, ciclo_ref)
);
