-- Governanca financeira cascata REP->Central / Central->Break-glass (runtime Go)
-- Embutido no binario apifunction (go:embed). Espelho: sql/004_fp_governanca_repasse.sql

CREATE TABLE IF NOT EXISTS fp_gov_repasse (
    fatura_id       INTEGER PRIMARY KEY,
    tipo            TEXT NOT NULL,
    id_representante TEXT NOT NULL DEFAULT '',
    id_central      TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'aberta',
    vencimento_em   TIMESTAMPTZ,
    valor_total     NUMERIC(12, 2) NOT NULL DEFAULT 0,
    synced_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fp_gov_repasse_tipo ON fp_gov_repasse (tipo);
CREATE INDEX IF NOT EXISTS idx_fp_gov_repasse_rep ON fp_gov_repasse (id_representante);
CREATE INDEX IF NOT EXISTS idx_fp_gov_repasse_cen ON fp_gov_repasse (id_central);
CREATE INDEX IF NOT EXISTS idx_fp_gov_repasse_status ON fp_gov_repasse (status);

CREATE TABLE IF NOT EXISTS fp_gov_restricao (
    entidade_tipo   TEXT NOT NULL,
    entidade_id     TEXT NOT NULL,
    fatura_id       INTEGER NOT NULL DEFAULT 0,
    nivel           TEXT NOT NULL DEFAULT 'alerta',
    vencimento_em   TIMESTAMPTZ,
    dias_restantes  INTEGER NOT NULL DEFAULT 0,
    ativo           BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (entidade_tipo, entidade_id)
);

CREATE INDEX IF NOT EXISTS idx_fp_gov_restricao_ativo ON fp_gov_restricao (ativo);

CREATE TABLE IF NOT EXISTS fp_gov_suspensao_cascata (
    id                  SERIAL PRIMARY KEY,
    id_franqueado       TEXT NOT NULL,
    assinatura_id       INTEGER NOT NULL,
    id_representante    TEXT NOT NULL DEFAULT '',
    fatura_repasse_id   INTEGER NOT NULL,
    motivo_interno      TEXT NOT NULL DEFAULT '',
    suspenso_em         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reativado_em        TIMESTAMPTZ,
    ativo               BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_fp_gov_susp_cascata_uniq
    ON fp_gov_suspensao_cascata (assinatura_id, fatura_repasse_id);

CREATE INDEX IF NOT EXISTS idx_fp_gov_susp_cascata_fra ON fp_gov_suspensao_cascata (id_franqueado, ativo);
