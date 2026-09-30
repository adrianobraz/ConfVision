-- Espelho financeiro read-only (sync Metadata API -> Postgres)
-- Fonte canonica: pgfinmirror/schema.sql (go:embed migrate no apifunction)

CREATE TABLE IF NOT EXISTS fp_fin_fatura (
    id                  INTEGER PRIMARY KEY,
    created_at          TIMESTAMPTZ,
    id_franqueado       TEXT NOT NULL DEFAULT '',
    id_representante    TEXT NOT NULL DEFAULT '',
    id_central          TEXT NOT NULL DEFAULT '',
    referencia          TEXT NOT NULL DEFAULT '',
    status              TEXT NOT NULL DEFAULT '',
    tipo                TEXT NOT NULL DEFAULT '',
    valor_total         NUMERIC(14, 2) NOT NULL DEFAULT 0,
    vencimento_em       TIMESTAMPTZ,
    pago_em             TIMESTAMPTZ,
    ciclo_ref           TEXT NOT NULL DEFAULT '',
    observacao          TEXT NOT NULL DEFAULT '',
    synced_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fp_fin_fatura_rep ON fp_fin_fatura (id_representante);
CREATE INDEX IF NOT EXISTS idx_fp_fin_fatura_cen ON fp_fin_fatura (id_central);
CREATE INDEX IF NOT EXISTS idx_fp_fin_fatura_fra ON fp_fin_fatura (id_franqueado);
CREATE INDEX IF NOT EXISTS idx_fp_fin_fatura_status ON fp_fin_fatura (status);

CREATE TABLE IF NOT EXISTS fp_fin_pagamento (
    id                  INTEGER PRIMARY KEY,
    created_at          TIMESTAMPTZ,
    fp_fatura_id        INTEGER NOT NULL DEFAULT 0,
    valor               NUMERIC(14, 2) NOT NULL DEFAULT 0,
    metodo              TEXT NOT NULL DEFAULT '',
    pago_em             TIMESTAMPTZ,
    observacao          TEXT NOT NULL DEFAULT '',
    synced_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fp_fin_pagamento_fatura ON fp_fin_pagamento (fp_fatura_id);

CREATE TABLE IF NOT EXISTS fp_fin_assinatura (
    id                  INTEGER PRIMARY KEY,
    created_at          TIMESTAMPTZ,
    id_franqueado       TEXT NOT NULL DEFAULT '',
    id_representante    TEXT NOT NULL DEFAULT '',
    id_central          TEXT NOT NULL DEFAULT '',
    produto             TEXT NOT NULL DEFAULT '',
    plano               TEXT NOT NULL DEFAULT '',
    status              TEXT NOT NULL DEFAULT '',
    valor               NUMERIC(14, 2) NOT NULL DEFAULT 0,
    observacao          TEXT NOT NULL DEFAULT '',
    synced_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fp_fin_assinatura_rep ON fp_fin_assinatura (id_representante);
CREATE INDEX IF NOT EXISTS idx_fp_fin_assinatura_cen ON fp_fin_assinatura (id_central);

CREATE TABLE IF NOT EXISTS fp_fin_caixa_movimento (
    id                  INTEGER PRIMARY KEY,
    created_at          TIMESTAMPTZ,
    tipo                TEXT NOT NULL DEFAULT '',
    valor               NUMERIC(14, 2) NOT NULL DEFAULT 0,
    descricao           TEXT NOT NULL DEFAULT '',
    movimento_em        TIMESTAMPTZ,
    admin_usuario       TEXT NOT NULL DEFAULT '',
    synced_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fp_fin_conta_pagar (
    id                  INTEGER PRIMARY KEY,
    created_at          TIMESTAMPTZ,
    fornecedor          TEXT NOT NULL DEFAULT '',
    descricao           TEXT NOT NULL DEFAULT '',
    valor               NUMERIC(14, 2) NOT NULL DEFAULT 0,
    status              TEXT NOT NULL DEFAULT '',
    vencimento_em       TIMESTAMPTZ,
    pago_em             TIMESTAMPTZ,
    admin_usuario       TEXT NOT NULL DEFAULT '',
    synced_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fp_fin_sync_state (
    chave               TEXT PRIMARY KEY,
    valor               TEXT NOT NULL DEFAULT '',
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fp_fin_shadow_diff (
    id                  BIGSERIAL PRIMARY KEY,
    escopo_tipo         TEXT NOT NULL DEFAULT '',
    escopo_id           TEXT NOT NULL DEFAULT '',
    competencia         TEXT NOT NULL DEFAULT '',
    campo               TEXT NOT NULL DEFAULT '',
    valor_postgres      TEXT NOT NULL DEFAULT '',
    valor_xano          TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
