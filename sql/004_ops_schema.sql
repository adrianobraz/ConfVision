-- Ops: grade CVG, arme/desarme janelas, autofim (PostgreSQL confmonit)
-- Aplicar: POSTGRES_URL=... go run ./cmd/apply ../../004_ops_schema.sql

-- ---------------------------------------------------------------------------
-- Grade horária ConfVision (CVG)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS vis_cliente_grade_config (
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado   TEXT,
    id_cliente      TEXT,
    grade_ativa     BOOLEAN DEFAULT FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_vis_cliente_grade_config_cliente_fra
    ON vis_cliente_grade_config (id_cliente, id_franqueado);

CREATE INDEX IF NOT EXISTS idx_vis_cliente_grade_config_cliente
    ON vis_cliente_grade_config (id_cliente);

CREATE TABLE IF NOT EXISTS vis_cliente_grade_slot (
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado   TEXT,
    id_cliente      TEXT,
    id_dispositivo  TEXT DEFAULT '',
    dia_semana      INT,
    hora            TEXT,
    acao            TEXT,
    ativo           BOOLEAN DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_vis_cliente_grade_slot_dia_hora
    ON vis_cliente_grade_slot (dia_semana, hora, ativo);

CREATE INDEX IF NOT EXISTS idx_vis_cliente_grade_slot_cliente
    ON vis_cliente_grade_slot (id_cliente);

CREATE TABLE IF NOT EXISTS vis_cliente_grade_exec (
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    slot_id         INT NOT NULL REFERENCES vis_cliente_grade_slot(id),
    data_ref        TEXT NOT NULL,
    hora_ref        TEXT NOT NULL,
    executado_em    TIMESTAMPTZ,
    resultado       TEXT,
    detalhe         JSONB DEFAULT '{}'::jsonb
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_vis_cliente_grade_exec_slot_data_hora
    ON vis_cliente_grade_exec (slot_id, data_ref, hora_ref);

CREATE TABLE IF NOT EXISTS vis_cliente_grade_escopo (
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado   TEXT,
    id_cliente      TEXT,
    id_dispositivo  TEXT DEFAULT '',
    grade_ativa     BOOLEAN DEFAULT FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_vis_cliente_grade_escopo
    ON vis_cliente_grade_escopo (id_cliente, id_franqueado, id_dispositivo);

-- ---------------------------------------------------------------------------
-- Arme/desarme automático — janelas (cadastro)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ops_arme_janela (
    id                  INT PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    whatsappeventocad_id INT,
    dias                TEXT[] DEFAULT '{}',
    hora_inicio         TEXT,
    hora_fim            TEXT,
    id_cliente          TEXT,
    id_franqueado       TEXT,
    whatsapp            TEXT,
    nome                TEXT,
    id_dispositivo      TEXT,
    nome_dispositivo    TEXT
);

CREATE INDEX IF NOT EXISTS idx_ops_arme_janela_hora_inicio ON ops_arme_janela (hora_inicio);
CREATE INDEX IF NOT EXISTS idx_ops_arme_janela_hora_fim ON ops_arme_janela (hora_fim);
CREATE INDEX IF NOT EXISTS idx_ops_arme_janela_dispositivo ON ops_arme_janela (id_dispositivo);

-- ---------------------------------------------------------------------------
-- Autofim — réplica alarm_events + fila + log
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ops_alarm_events (
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_evento       TEXT,
    codigo          TEXT,
    particao        TEXT,
    zona_user       TEXT,
    nivel           TEXT,
    data_entrada    TEXT,
    img             TEXT,
    id_processo     TEXT,
    id_dispositivo  TEXT,
    id_cliente      TEXT,
    nome_cliente    TEXT,
    email_cliente   TEXT,
    cti_grupo       TEXT,
    cti_descricao   TEXT,
    id_franqueado   TEXT,
    codigo_benuvem  TEXT,
    conta           TEXT,
    camera_ativa    TEXT,
    usa_confvision  TEXT,
    provedor_video  TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_alarm_events_id_evento
    ON ops_alarm_events (id_evento) WHERE id_evento IS NOT NULL AND id_evento <> '';

CREATE INDEX IF NOT EXISTS idx_ops_alarm_events_processo
    ON ops_alarm_events (id_processo);

CREATE INDEX IF NOT EXISTS idx_ops_alarm_events_dispositivo_created
    ON ops_alarm_events (id_dispositivo, created_at DESC);

CREATE TABLE IF NOT EXISTS ops_bot_finalizaeventoauto (
    id                  SERIAL PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_processo         TEXT,
    id_dispositivo      TEXT,
    ultimo_evento_ts    BIGINT,
    status              TEXT DEFAULT 'PENDENTE',
    rodar_em            BIGINT NOT NULL DEFAULT 0,
    tentativas          INT NOT NULL DEFAULT 0,
    lock_token          TEXT DEFAULT '',
    lock_at             TIMESTAMPTZ,
    motivo_final        TEXT DEFAULT '',
    acao_final          TEXT DEFAULT '',
    erro                TEXT DEFAULT '',
    payload_resultado   JSONB DEFAULT '{}'::jsonb,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    alarm_events_id     INT
);

CREATE INDEX IF NOT EXISTS idx_ops_bot_autofim_pending
    ON ops_bot_finalizaeventoauto (status, rodar_em);

CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_bot_autofim_processo
    ON ops_bot_finalizaeventoauto (id_processo) WHERE id_processo IS NOT NULL AND id_processo <> '';

CREATE TABLE IF NOT EXISTS ops_alarm_event_autofimlog (
    id                      SERIAL PRIMARY KEY,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_evento               INT,
    id_processo             TEXT,
    motivo                  TEXT,
    acao                    TEXT,
    qtd_ciclos_5m           INT,
    bloqueio_3x5            BOOLEAN DEFAULT FALSE,
    tem_falhas              BOOLEAN DEFAULT FALSE,
    tem_alarme              BOOLEAN DEFAULT FALSE,
    tem_desarme             BOOLEAN DEFAULT FALSE,
    tem_restaure            BOOLEAN DEFAULT FALSE,
    tem_par_alarme_rest_50  BOOLEAN DEFAULT FALSE,
    retorno_processo_end    JSONB DEFAULT '{}'::jsonb,
    id_dispositivo          TEXT,
    regra_versao            TEXT DEFAULT 'v1',
    erro                    TEXT DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_ops_autofimlog_processo ON ops_alarm_event_autofimlog (id_processo);
CREATE INDEX IF NOT EXISTS idx_ops_autofimlog_evento ON ops_alarm_event_autofimlog (id_evento);
