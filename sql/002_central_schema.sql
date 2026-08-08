-- ConfVision — schema central (PostgreSQL confmonit)
-- Substitui dados operacionais do Xano (fase 1: cameras, areas, nos, workers)
-- Aplicar: psql -U confmonit -d confmonit -f 002_central_schema.sql

-- ---------------------------------------------------------------------------
-- Nós MediaMTX (servidor1, servidor2, …)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS vis_mediamtx_node (
    id                  SERIAL PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    nome                TEXT,
    nome_exibicao       TEXT,
    hostname            TEXT,
    provedor            TEXT,
    datacenter          TEXT,
    ip_publico          TEXT,
    rtmp_public         TEXT,
    hls_public          TEXT,
    rtsp_internal       TEXT DEFAULT 'rtsp://127.0.0.1:8554',
    guard_url           TEXT,
    max_cameras         INT NOT NULL DEFAULT 200,
    ordem               INT NOT NULL DEFAULT 1,
    status              TEXT DEFAULT 'ativo',
    observacao          TEXT,
    -- pontos / capacidade
    limite_pontos       NUMERIC(10, 2) NOT NULL DEFAULT 90,
    peso_online         NUMERIC(10, 4) NOT NULL DEFAULT 0.125,
    peso_analitico      NUMERIC(10, 4) NOT NULL DEFAULT 0.833,
    max_online_ref      INT NOT NULL DEFAULT 800,
    max_analitico_ref   INT NOT NULL DEFAULT 40,
    online_ativas       INT NOT NULL DEFAULT 0,
    analitico_ativas    INT NOT NULL DEFAULT 0,
    pontos_atual        NUMERIC(10, 2) NOT NULL DEFAULT 0,
    cpu_percent         NUMERIC(5, 2),
    mem_percent         NUMERIC(5, 2),
    ultimo_ping_em      TIMESTAMPTZ,
    custo_mensal_brl    NUMERIC(12, 2)
);

CREATE INDEX IF NOT EXISTS idx_vis_mediamtx_node_status ON vis_mediamtx_node (status);
CREATE INDEX IF NOT EXISTS idx_vis_mediamtx_node_ordem ON vis_mediamtx_node (ordem);

-- ---------------------------------------------------------------------------
-- Câmeras
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS vis_camera (
    id                      SERIAL PRIMARY KEY,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ativo                   BOOLEAN DEFAULT TRUE,
    bloqueado               BOOLEAN DEFAULT FALSE,
    nome                    TEXT,
    id_franqueado           TEXT,
    id_cliente              TEXT,
    id_dispositivo          TEXT,
    conta                   TEXT,
    particao                TEXT,
    canal                   TEXT,
    setor                   TEXT,
    protocolo               TEXT,
    rtsp_url_sec            TEXT,
    onvif_host              TEXT,
    onvif_porta             INT,
    onvif_usuario           TEXT,
    onvif_senha             TEXT,
    confianca_min           NUMERIC(6, 4),
    cooldown_seg            INT,
    somente_armado          BOOLEAN DEFAULT FALSE,
    deteccao_humano         BOOLEAN DEFAULT FALSE,
    deteccao_veiculo        BOOLEAN DEFAULT FALSE,
    status                  TEXT,
    ultimo_evento_em        TIMESTAMPTZ,
    worker_id               TEXT,
    ultimo_ping_em          TIMESTAMPTZ,
    zonauser                TEXT,
    captura_sensor          BOOLEAN DEFAULT FALSE,
    captura_analitico       BOOLEAN DEFAULT FALSE,
    analitico_pausado       BOOLEAN DEFAULT FALSE,
    id_setor                TEXT,
    snapshot_url            TEXT,
    vis_licenca_id          INT,
    plano                   TEXT,
    ativado_em              TIMESTAMPTZ,
    evento_grava_foto       BOOLEAN,
    evento_grava_video      BOOLEAN,
    vis_licenca_gravacao_id INT,
    grava_continua          BOOLEAN DEFAULT FALSE,
    retencao_dias           INT,
    gravacao_ativada_em     TIMESTAMPTZ,
    gravacao_status         TEXT,
    grava_movimento         BOOLEAN DEFAULT FALSE,
    grava_timelapse         BOOLEAN DEFAULT FALSE,
    gravacao_flush_pedido   BOOLEAN DEFAULT FALSE,
    modo_deteccao           TEXT,
    vis_mediamtx_node_id    INT REFERENCES vis_mediamtx_node (id)
);

CREATE INDEX IF NOT EXISTS idx_vis_camera_franqueado ON vis_camera (id_franqueado);
CREATE INDEX IF NOT EXISTS idx_vis_camera_cliente ON vis_camera (id_cliente);
CREATE INDEX IF NOT EXISTS idx_vis_camera_dispositivo ON vis_camera (id_dispositivo);
CREATE INDEX IF NOT EXISTS idx_vis_camera_worker ON vis_camera (worker_id);
CREATE INDEX IF NOT EXISTS idx_vis_camera_node ON vis_camera (vis_mediamtx_node_id);
CREATE INDEX IF NOT EXISTS idx_vis_camera_analitico ON vis_camera (ativo, deteccao_humano, analitico_pausado);

-- ---------------------------------------------------------------------------
-- Áreas de detecção
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS vis_camera_area (
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    vis_camera_id   INT REFERENCES vis_camera (id) ON DELETE CASCADE,
    nome            TEXT,
    ativo           BOOLEAN DEFAULT TRUE,
    poligono_json   TEXT,
    cor             TEXT
);

CREATE INDEX IF NOT EXISTS idx_vis_camera_area_camera ON vis_camera_area (vis_camera_id);
CREATE INDEX IF NOT EXISTS idx_vis_camera_area_ativo ON vis_camera_area (ativo);

-- ---------------------------------------------------------------------------
-- Workers (heartbeat)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS vis_worker (
    id                      SERIAL PRIMARY KEY,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    worker_id               TEXT NOT NULL,
    worker_tipo             TEXT NOT NULL DEFAULT 'analitico',
    hostname                TEXT,
    versao                  TEXT,
    cameras_ativas          INT DEFAULT 0,
    ultimo_ping_em          TIMESTAMPTZ,
    ativo                   BOOLEAN DEFAULT TRUE,
    shard_index             INT,
    shard_total             INT,
    max_cameras             INT,
    yolo_device             TEXT,
    queue_backend           TEXT,
    vis_mediamtx_node_id    INT NOT NULL DEFAULT 0,
    cpu_percent             NUMERIC(5, 2),
    mem_percent             NUMERIC(5, 2),
    load_1m                 NUMERIC(8, 2)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_vis_worker_unique
    ON vis_worker (worker_id, worker_tipo, vis_mediamtx_node_id);

-- ---------------------------------------------------------------------------
-- Eventos (fase 1b — estrutura pronta)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS vis_evento (
    id                  SERIAL PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    vis_camera_id       INT REFERENCES vis_camera (id),
    id_franqueado       TEXT,
    id_cliente          TEXT,
    id_dispositivo      TEXT,
    conta               TEXT,
    particao            TEXT,
    canal               TEXT,
    tipo_deteccao       TEXT,
    confianca           NUMERIC(6, 4),
    snapshot_url        TEXT,
    video_url           TEXT,
    bbox_json           TEXT,
    processado          BOOLEAN DEFAULT FALSE,
    alarm_events_id     INT,
    ignorado            BOOLEAN DEFAULT FALSE,
    status              TEXT,
    id_evento           TEXT,
    id_processo         TEXT,
    started_at          TIMESTAMPTZ,
    ended_at            TIMESTAMPTZ,
    clip_count          INT DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_vis_evento_camera ON vis_evento (vis_camera_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_vis_evento_franqueado ON vis_evento (id_franqueado, created_at DESC);

CREATE TABLE IF NOT EXISTS vis_evento_clip (
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    vis_evento_id   INT REFERENCES vis_evento (id) ON DELETE CASCADE,
    seq             INT DEFAULT 1,
    video_url       TEXT,
    duracao_seg     INT,
    snapshot_url    TEXT
);

CREATE INDEX IF NOT EXISTS idx_vis_evento_clip_evento ON vis_evento_clip (vis_evento_id);

-- ---------------------------------------------------------------------------
-- Função: recalcular pontos de um nó
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION vis_mediamtx_node_recalc_pontos(p_node_id INT)
RETURNS VOID AS $$
DECLARE
    v_online INT;
    v_analitico INT;
    v_peso_online NUMERIC;
    v_peso_analitico NUMERIC;
    v_limite NUMERIC;
    v_pontos NUMERIC;
    v_status TEXT;
BEGIN
    SELECT peso_online, peso_analitico, limite_pontos
    INTO v_peso_online, v_peso_analitico, v_limite
    FROM vis_mediamtx_node WHERE id = p_node_id;

    IF NOT FOUND THEN
        RETURN;
    END IF;

    SELECT COUNT(*) INTO v_online
    FROM vis_camera
    WHERE vis_mediamtx_node_id = p_node_id
      AND ativo = TRUE
      AND COALESCE(deteccao_humano, FALSE) = FALSE;

    SELECT COUNT(*) INTO v_analitico
    FROM vis_camera
    WHERE vis_mediamtx_node_id = p_node_id
      AND ativo = TRUE
      AND deteccao_humano = TRUE;

    v_pontos := (v_online * v_peso_online) + (v_analitico * v_peso_analitico);

    IF v_pontos >= v_limite THEN
        v_status := 'cheio';
    ELSIF v_pontos >= (v_limite * 0.75) THEN
        v_status := 'alerta';
    ELSE
        v_status := 'ativo';
    END IF;

    UPDATE vis_mediamtx_node
    SET online_ativas = v_online,
        analitico_ativas = v_analitico,
        pontos_atual = v_pontos,
        status = v_status
    WHERE id = p_node_id;
END;
$$ LANGUAGE plpgsql;

-- Nó padrão (servidor1) se vazio
INSERT INTO vis_mediamtx_node (nome, nome_exibicao, status, ordem)
SELECT 'servidor1', 'Servidor 1', 'ativo', 1
WHERE NOT EXISTS (SELECT 1 FROM vis_mediamtx_node LIMIT 1);
