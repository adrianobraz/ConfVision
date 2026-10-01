-- ConfVision — extensao schema (licencas, gravacao, funcoes auxiliares)
-- Aplicar apos 002_central_schema.sql

-- ---------------------------------------------------------------------------
-- Licencas prepagas
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS vis_licenca (
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado   TEXT,
    plano           TEXT,
    unidade         TEXT DEFAULT 'camera',
    valor           NUMERIC(12, 2),
    pago_em         TIMESTAMPTZ,
    valido_ate      TIMESTAMPTZ,
    status          TEXT DEFAULT 'disponivel',
    id_dispositivo  TEXT,
    id_fatura       TEXT,
    id_pagamento    TEXT,
    observacao      TEXT,
    vis_camera_id   INT REFERENCES vis_camera (id)
);

CREATE INDEX IF NOT EXISTS idx_vis_licenca_franqueado ON vis_licenca (id_franqueado);
CREATE INDEX IF NOT EXISTS idx_vis_licenca_status ON vis_licenca (id_franqueado, status);
CREATE INDEX IF NOT EXISTS idx_vis_licenca_unidade ON vis_licenca (id_franqueado, unidade, status);

-- ---------------------------------------------------------------------------
-- Gravacao storage (Contabo S3 por franqueado)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS vis_gravacao_storage (
    id                  SERIAL PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado       TEXT,
    s3_endpoint         TEXT,
    s3_bucket           TEXT,
    s3_tenant_id        TEXT,
    s3_access_key       TEXT,
    s3_secret_key       TEXT,
    segmento_minutos    INT DEFAULT 5,
    status              TEXT DEFAULT 'ativo',
    provisionado_em     TIMESTAMPTZ,
    cancelado_em        TIMESTAMPTZ,
    observacao          TEXT
);

CREATE INDEX IF NOT EXISTS idx_vis_gravacao_storage_franqueado ON vis_gravacao_storage (id_franqueado, status);

-- ---------------------------------------------------------------------------
-- Segmentos de gravacao (DVR / motion / timelapse)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS vis_gravacao_segmento (
    id                      SERIAL PRIMARY KEY,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    vis_camera_id           INT REFERENCES vis_camera (id),
    vis_gravacao_storage_id INT REFERENCES vis_gravacao_storage (id),
    id_franqueado           TEXT,
    id_cliente              TEXT,
    inicio_em               TIMESTAMPTZ,
    fim_em                  TIMESTAMPTZ,
    duracao_seg             INT,
    s3_key                  TEXT,
    s3_url                  TEXT,
    tamanho_bytes           BIGINT,
    status                  TEXT,
    erro_msg                TEXT,
    uploaded_em             TIMESTAMPTZ,
    expira_em               TIMESTAMPTZ,
    tipo                    TEXT
);

CREATE INDEX IF NOT EXISTS idx_vis_gravacao_segmento_camera ON vis_gravacao_segmento (vis_camera_id, inicio_em DESC);
CREATE INDEX IF NOT EXISTS idx_vis_gravacao_segmento_franqueado ON vis_gravacao_segmento (id_franqueado, inicio_em DESC);

-- FK licenca -> camera (apos ambas existirem)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'vis_licenca_vis_camera_id_fkey'
    ) THEN
        ALTER TABLE vis_licenca
            ADD CONSTRAINT vis_licenca_vis_camera_id_fkey
            FOREIGN KEY (vis_camera_id) REFERENCES vis_camera (id);
    END IF;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
