-- Coleta operacional (relatório stream + health/metric a cada 30 min).
-- Idempotente: IF NOT EXISTS. Não altera tabelas existentes.

-- vis_camera stream policy (reaplicar se migration 20260326 ainda não rodou no ambiente)
ALTER TABLE vis_camera ADD COLUMN IF NOT EXISTS ultimo_stream_ok_em TIMESTAMPTZ;
ALTER TABLE vis_camera ADD COLUMN IF NOT EXISTS stream_falhas_consecutivas INT NOT NULL DEFAULT 0;
ALTER TABLE vis_camera ADD COLUMN IF NOT EXISTS stream_tentativas_horarias INT NOT NULL DEFAULT 0;
ALTER TABLE vis_camera ADD COLUMN IF NOT EXISTS stream_policy_generation BIGINT NOT NULL DEFAULT 0;
ALTER TABLE vis_camera ADD COLUMN IF NOT EXISTS stream_motivo_pausa TEXT;
ALTER TABLE vis_camera ADD COLUMN IF NOT EXISTS stream_ultimo_erro TEXT;
ALTER TABLE vis_camera ADD COLUMN IF NOT EXISTS stream_erro_classe VARCHAR(64);
ALTER TABLE vis_camera ADD COLUMN IF NOT EXISTS stream_ultimo_erro_em TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS vis_stream_relatorio (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado VARCHAR(64),
    vis_camera_id INT REFERENCES vis_camera(id) ON DELETE SET NULL,
    fonte VARCHAR(32) NOT NULL,
    motivo_codigo VARCHAR(64),
    titulo TEXT,
    dica TEXT,
    severidade VARCHAR(16),
    detalhe_json JSONB,
    referencia_dedupe VARCHAR(256)
);

CREATE INDEX IF NOT EXISTS idx_vis_stream_relatorio_created ON vis_stream_relatorio (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_vis_stream_relatorio_fra ON vis_stream_relatorio (id_franqueado, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_vis_stream_relatorio_cam ON vis_stream_relatorio (vis_camera_id, created_at DESC);

CREATE TABLE IF NOT EXISTS vis_sistema_health (
    id BIGSERIAL PRIMARY KEY,
    coletado_em TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    componente VARCHAR(64) NOT NULL,
    vis_mediamtx_node_id INT,
    base_url TEXT,
    status VARCHAR(16) NOT NULL,
    http_status INT,
    latencia_ms INT,
    mensagem TEXT,
    detalhe_json JSONB
);

CREATE INDEX IF NOT EXISTS idx_vis_sistema_health_coletado ON vis_sistema_health (coletado_em DESC);
CREATE INDEX IF NOT EXISTS idx_vis_sistema_health_comp ON vis_sistema_health (componente, coletado_em DESC);

CREATE TABLE IF NOT EXISTS vis_sistema_metric (
    id BIGSERIAL PRIMARY KEY,
    coletado_em TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    escopo VARCHAR(32) NOT NULL DEFAULT 'global',
    id_franqueado VARCHAR(64),
    chave VARCHAR(128),
    valor_num DOUBLE PRECISION,
    valor_text TEXT,
    tags_json JSONB,
    metricas_json JSONB
);

CREATE INDEX IF NOT EXISTS idx_vis_sistema_metric_coletado ON vis_sistema_metric (coletado_em DESC);
CREATE INDEX IF NOT EXISTS idx_vis_sistema_metric_escopo ON vis_sistema_metric (escopo, coletado_em DESC);
