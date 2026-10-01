-- Integracao de eventos (Moni, etc.) + codigo interno cliente + imagem publica

ALTER TABLE vis_evento
    ADD COLUMN IF NOT EXISTS codigo_imagem_publico TEXT,
    ADD COLUMN IF NOT EXISTS imagem_liberada_em TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS vis_integracao_imagem_id INT;

CREATE INDEX IF NOT EXISTS idx_vis_evento_codigo_imagem
    ON vis_evento (codigo_imagem_publico)
    WHERE codigo_imagem_publico IS NOT NULL;

CREATE TABLE IF NOT EXISTS vis_cliente_ext (
    id              SERIAL PRIMARY KEY,
    id_franqueado   TEXT NOT NULL,
    id_cliente      TEXT NOT NULL,
    codigo_interno  TEXT,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (id_franqueado, id_cliente)
);

CREATE INDEX IF NOT EXISTS idx_vis_cliente_ext_franqueado
    ON vis_cliente_ext (id_franqueado, id_cliente);

CREATE TABLE IF NOT EXISTS vis_integracao_config (
    id                      SERIAL PRIMARY KEY,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado           TEXT NOT NULL,
    sistema                 TEXT NOT NULL,
    nome                    TEXT,
    ativo                   BOOLEAN NOT NULL DEFAULT TRUE,
    eventos_url             TEXT,
    auth_tipo               TEXT DEFAULT 'none',
    auth_user               TEXT,
    auth_pass               TEXT,
    auth_token              TEXT,
    empresa_codigo          TEXT,
    evento_codigo           TEXT,
    setor_padrao            TEXT,
    particao_padrao         TEXT,
    identificacao_padrao    TEXT,
    codigo_integracao_imagem TEXT DEFAULT '046',
    enviar_imagem           BOOLEAN NOT NULL DEFAULT TRUE,
    config_json             TEXT DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_vis_integracao_franqueado
    ON vis_integracao_config (id_franqueado, sistema);

CREATE TABLE IF NOT EXISTS vis_integracao_log (
    id                  SERIAL PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado       TEXT,
    vis_integracao_id   INT REFERENCES vis_integracao_config (id) ON DELETE SET NULL,
    vis_evento_id       INT REFERENCES vis_evento (id) ON DELETE SET NULL,
    sistema             TEXT,
    sucesso             BOOLEAN NOT NULL DEFAULT FALSE,
    http_status         INT,
    mensagem            TEXT,
    payload_resumo      TEXT
);

CREATE INDEX IF NOT EXISTS idx_vis_integracao_log_franqueado
    ON vis_integracao_log (id_franqueado, created_at DESC);
