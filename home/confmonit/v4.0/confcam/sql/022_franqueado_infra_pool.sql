-- Pool obrigatorio de infra ConfVision por franqueado (MediaMTX + worker analitico)
-- Aplicar no Postgres central antes de deploy ConfCam com assign automatico.

CREATE TABLE IF NOT EXISTS vis_franqueado_mediamtx_pool (
    id                  SERIAL PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado       TEXT NOT NULL,
    vis_mediamtx_node_id INT NOT NULL REFERENCES vis_mediamtx_node (id),
    max_cameras         INT,
    ordem               INT NOT NULL DEFAULT 1,
    ativo               BOOLEAN NOT NULL DEFAULT TRUE,
    CONSTRAINT uq_franqueado_mtx_pool UNIQUE (id_franqueado, vis_mediamtx_node_id)
);

CREATE INDEX IF NOT EXISTS idx_franqueado_mtx_pool_fra
    ON vis_franqueado_mediamtx_pool (id_franqueado, ativo);

CREATE TABLE IF NOT EXISTS vis_franqueado_worker_pool (
    id                  SERIAL PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado       TEXT NOT NULL,
    worker_id           TEXT NOT NULL,
    worker_tipo         TEXT NOT NULL DEFAULT 'analitico',
    max_cameras         INT NOT NULL DEFAULT 30,
    ordem               INT NOT NULL DEFAULT 1,
    ativo               BOOLEAN NOT NULL DEFAULT TRUE,
    CONSTRAINT uq_franqueado_worker_pool UNIQUE (id_franqueado, worker_id, worker_tipo)
);

CREATE INDEX IF NOT EXISTS idx_franqueado_worker_pool_fra
    ON vis_franqueado_worker_pool (id_franqueado, ativo);
