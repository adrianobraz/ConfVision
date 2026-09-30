-- Domínios da marca (Central) — espelho para deploy manual
CREATE TABLE IF NOT EXISTS fp_central_dominio_marca (
    id_central   VARCHAR(64) PRIMARY KEY,
    dominio      VARCHAR(253) NOT NULL DEFAULT '',
    apps         JSONB NOT NULL DEFAULT '{}',
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fp_central_dominio_marca_updated ON fp_central_dominio_marca (updated_at DESC);
