-- Configuracao SMTP por franqueado (PostgreSQL confmonit)
-- Aplicar: POSTGRES_URL=... go run ./cmd/apply ../../012_smtp_config.sql

CREATE TABLE IF NOT EXISTS ops_franqueado_smtp_config (
    id_franqueado       TEXT PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    nome_config         TEXT NOT NULL DEFAULT '',
    descricao           TEXT NOT NULL DEFAULT '',
    ativo               BOOLEAN NOT NULL DEFAULT TRUE,
    provedor            TEXT NOT NULL DEFAULT 'outro',

    remetente_nome      TEXT NOT NULL DEFAULT '',
    remetente_email     TEXT NOT NULL DEFAULT '',

    smtp_host           TEXT NOT NULL DEFAULT '',
    smtp_porta          INT NOT NULL DEFAULT 587,
    smtp_seguranca      TEXT NOT NULL DEFAULT 'starttls',
    smtp_autenticacao   BOOLEAN NOT NULL DEFAULT TRUE,
    smtp_usuario        TEXT NOT NULL DEFAULT '',
    smtp_senha          TEXT NOT NULL DEFAULT '',
    smtp_timeout_seg    INT NOT NULL DEFAULT 30
);

CREATE INDEX IF NOT EXISTS idx_ops_smtp_ativo
    ON ops_franqueado_smtp_config (id_franqueado)
    WHERE ativo = TRUE;
