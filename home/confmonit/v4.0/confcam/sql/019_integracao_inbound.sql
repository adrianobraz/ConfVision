-- Webhook inbound (Moni arm/disarm -> ConfVision)

ALTER TABLE vis_integracao_config
    ADD COLUMN IF NOT EXISTS webhook_inbound_ativo BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS codigo_evento_armar TEXT,
    ADD COLUMN IF NOT EXISTS codigo_evento_desarmar TEXT;
