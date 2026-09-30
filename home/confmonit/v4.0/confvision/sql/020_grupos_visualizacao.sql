-- Grupos de visualização e mosaicos (ConfVision)

CREATE TABLE IF NOT EXISTS vis_grupo_visualizacao (
    id              SERIAL PRIMARY KEY,
    id_franqueado   VARCHAR(64) NOT NULL,
    nome            VARCHAR(255) NOT NULL,
    descricao       TEXT,
    ativo           BOOLEAN NOT NULL DEFAULT TRUE,
    layout_mosaic   JSONB NOT NULL DEFAULT '{"modo":"auto"}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vis_grupo_vis_fra ON vis_grupo_visualizacao (id_franqueado);
CREATE INDEX IF NOT EXISTS idx_vis_grupo_vis_ativo ON vis_grupo_visualizacao (id_franqueado, ativo);

CREATE TABLE IF NOT EXISTS vis_grupo_visualizacao_cliente (
    id          SERIAL PRIMARY KEY,
    grupo_id    INT NOT NULL REFERENCES vis_grupo_visualizacao (id) ON DELETE CASCADE,
    id_cliente  VARCHAR(64) NOT NULL,
    UNIQUE (grupo_id, id_cliente)
);

CREATE INDEX IF NOT EXISTS idx_vis_grupo_vis_cli ON vis_grupo_visualizacao_cliente (id_cliente);
CREATE INDEX IF NOT EXISTS idx_vis_grupo_vis_cli_grupo ON vis_grupo_visualizacao_cliente (grupo_id);

CREATE TABLE IF NOT EXISTS vis_grupo_visualizacao_camera (
    id              SERIAL PRIMARY KEY,
    grupo_id        INT NOT NULL REFERENCES vis_grupo_visualizacao (id) ON DELETE CASCADE,
    id_cliente      VARCHAR(64) NOT NULL,
    vis_camera_id   INT NOT NULL REFERENCES vis_camera (id) ON DELETE CASCADE,
    ordem           INT NOT NULL DEFAULT 0,
    ativo           BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (grupo_id, vis_camera_id)
);

CREATE INDEX IF NOT EXISTS idx_vis_grupo_vis_cam_grupo ON vis_grupo_visualizacao_camera (grupo_id, ordem);
CREATE INDEX IF NOT EXISTS idx_vis_grupo_vis_cam_cli ON vis_grupo_visualizacao_camera (grupo_id, id_cliente);
