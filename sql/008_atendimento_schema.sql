-- Configuracao de atendimento FranqueadoPro (PostgreSQL confmonit)
-- Aplicar: POSTGRES_URL=... go run ./cmd/apply ../../008_atendimento_schema.sql

-- ---------------------------------------------------------------------------
-- Politica padrao do franqueado (modo todos / todos_menos / somente)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ops_franqueado_atendimento_politica (
    id_franqueado               TEXT PRIMARY KEY,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    inteligencia_artificial     BOOLEAN NOT NULL DEFAULT FALSE,
    ia_modo                     TEXT NOT NULL DEFAULT 'todos',
    finalizacao_automatica      BOOLEAN NOT NULL DEFAULT FALSE,
    autofim_modo                TEXT NOT NULL DEFAULT 'todos',
    parceiro_monitoramento      BOOLEAN NOT NULL DEFAULT FALSE,
    parceiro_modo               TEXT NOT NULL DEFAULT 'todos',
    email_ativo                 BOOLEAN NOT NULL DEFAULT FALSE,
    email_modo                  TEXT NOT NULL DEFAULT 'todos',
    cobertura_horaria           TEXT NOT NULL DEFAULT '24h',
    id_parceiro                 TEXT,
    grupos_evento_parceiro      TEXT[] NOT NULL DEFAULT '{}'
);

-- ---------------------------------------------------------------------------
-- Lista de clientes (excecoes ou inclusoes por recurso)
-- recurso: inteligencia_artificial | finalizacao_automatica | parceiro | email
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ops_franqueado_atendimento_lista (
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado   TEXT NOT NULL,
    recurso         TEXT NOT NULL,
    id_cliente      TEXT NOT NULL,
    nome_cliente    TEXT DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_atend_lista
    ON ops_franqueado_atendimento_lista (id_franqueado, recurso, id_cliente);

CREATE INDEX IF NOT EXISTS idx_ops_atend_lista_fra_rec
    ON ops_franqueado_atendimento_lista (id_franqueado, recurso);

-- ---------------------------------------------------------------------------
-- Config por cliente (override ou registro explicito pos-migracao)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ops_cliente_atendimento_config (
    id                          SERIAL PRIMARY KEY,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado               TEXT NOT NULL,
    id_cliente                  TEXT NOT NULL,
    inteligencia_artificial     BOOLEAN NOT NULL DEFAULT FALSE,
    finalizacao_automatica      BOOLEAN NOT NULL DEFAULT FALSE,
    parceiro_monitoramento      BOOLEAN NOT NULL DEFAULT FALSE,
    email_ativo                 BOOLEAN NOT NULL DEFAULT FALSE,
    cobertura_horaria           TEXT NOT NULL DEFAULT '24h',
    id_parceiro                 TEXT,
    grupos_evento_parceiro      TEXT[] NOT NULL DEFAULT '{}',
    ativo                       BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_cliente_atend_config
    ON ops_cliente_atendimento_config (id_franqueado, id_cliente);

-- Staging para migracao MySQL -> Postgres (liberacao inicial)
CREATE TABLE IF NOT EXISTS ops_stg_cliente_ativo (
    id_franqueado TEXT NOT NULL,
    id_cliente    TEXT NOT NULL,
    nome_cliente  TEXT DEFAULT '',
    PRIMARY KEY (id_franqueado, id_cliente)
);

-- ---------------------------------------------------------------------------
-- Grade horaria (inteligencia_artificial | parceiro)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ops_cliente_atendimento_grade (
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado   TEXT NOT NULL,
    id_cliente      TEXT NOT NULL DEFAULT '',
    responsavel     TEXT NOT NULL,
    dias_semana     INT[] NOT NULL DEFAULT '{}',
    hora_inicio     TIME NOT NULL,
    hora_fim        TIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_ops_atend_grade_fra_cli
    ON ops_cliente_atendimento_grade (id_franqueado, id_cliente);

-- ---------------------------------------------------------------------------
-- Bloqueio IA (Excecoes da IA) — regra maior; substitui WhatsEventCadBloq
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ops_ia_bloqueio_cliente (
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado   TEXT NOT NULL,
    id_cliente      TEXT NOT NULL,
    nome_cliente    TEXT DEFAULT '',
    motivo          TEXT DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_ia_bloqueio
    ON ops_ia_bloqueio_cliente (id_franqueado, id_cliente);

CREATE INDEX IF NOT EXISTS idx_ops_ia_bloqueio_fra
    ON ops_ia_bloqueio_cliente (id_franqueado);

-- ---------------------------------------------------------------------------
-- Creditos pre-pagos por canal
-- canal: ligacao | sms | whatsapp | email
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ops_credito_saldo (
    id_franqueado   TEXT NOT NULL,
    canal           TEXT NOT NULL,
    saldo           NUMERIC(14, 4) NOT NULL DEFAULT 0,
    saldo_inicial   NUMERIC(14, 4) NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id_franqueado, canal)
);

CREATE TABLE IF NOT EXISTS ops_credito_movimento (
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    id_franqueado   TEXT NOT NULL,
    canal           TEXT NOT NULL,
    tipo            TEXT NOT NULL,
    quantidade      NUMERIC(14, 4) NOT NULL,
    valor_unitario  NUMERIC(12, 4),
    valor_total     NUMERIC(12, 2),
    id_fatura       BIGINT,
    id_evento       TEXT,
    id_processo     TEXT,
    observacao      TEXT,
    criado_por      TEXT DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_ops_credito_mov_fra_data
    ON ops_credito_movimento (id_franqueado, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ops_credito_mov_canal
    ON ops_credito_movimento (id_franqueado, canal, created_at DESC);
