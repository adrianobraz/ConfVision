-- Schema edge ConfVision (Postgres local na VPS)
-- Eventos gravados localmente; sync opcional para Xano (UI/billing)

CREATE TABLE IF NOT EXISTS vis_evento_edge (
    id BIGSERIAL PRIMARY KEY,
    xano_evento_id BIGINT,
    vis_camera_id INT NOT NULL,
    id_franqueado TEXT,
    id_cliente TEXT,
    id_dispositivo TEXT,
    conta TEXT,
    particao TEXT,
    canal TEXT,
    tipo_deteccao TEXT DEFAULT 'humano',
    confianca NUMERIC(6, 4),
    snapshot_url TEXT,
    video_url TEXT,
    bbox_json TEXT,
    status TEXT DEFAULT 'capturando',
    clip_count INT DEFAULT 0,
    processado BOOLEAN DEFAULT FALSE,
    id_evento TEXT,
    id_processo TEXT,
    synced_at TIMESTAMPTZ,
    sync_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vis_evento_edge_sync
    ON vis_evento_edge (synced_at NULLS FIRST, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_vis_evento_edge_camera
    ON vis_evento_edge (vis_camera_id, created_at DESC);

CREATE TABLE IF NOT EXISTS vis_evento_clip_edge (
    id BIGSERIAL PRIMARY KEY,
    edge_evento_id BIGINT NOT NULL REFERENCES vis_evento_edge(id) ON DELETE CASCADE,
    xano_clip_id BIGINT,
    seq INT DEFAULT 1,
    video_url TEXT,
    duracao_seg INT,
    snapshot_url TEXT,
    synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vis_evento_clip_edge_evento
    ON vis_evento_clip_edge (edge_evento_id);
