CREATE TABLE IF NOT EXISTS fp_central_whitelabel (
    id_central   VARCHAR(64) PRIMARY KEY,
    tema_json    JSONB NOT NULL DEFAULT '{}',
    logos_json   JSONB NOT NULL DEFAULT '{}',
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fp_central_whitelabel_updated ON fp_central_whitelabel (updated_at DESC);
