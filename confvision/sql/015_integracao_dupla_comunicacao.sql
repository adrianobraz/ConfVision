-- Dupla comunicacao: envia ao sistema integrado e tambem ao ConfMonit (terminal)

ALTER TABLE vis_integracao_config
    ADD COLUMN IF NOT EXISTS dupla_comunicacao BOOLEAN NOT NULL DEFAULT FALSE;
