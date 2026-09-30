-- Diagnostico receptor — presenca IP/sinal por central (PostgreSQL)
CREATE TABLE IF NOT EXISTS fp_receptor_presenca (
    id_franqueado   TEXT NOT NULL DEFAULT '',
    fabricante      TEXT NOT NULL,
    modulo          TEXT NOT NULL,
    conta           TEXT NOT NULL,
    id_dispositivo  TEXT NOT NULL DEFAULT '',
    ip_remoto       TEXT NOT NULL DEFAULT '',
    ultimo_evento   TEXT NOT NULL DEFAULT '',
    ultimo_sinal_at TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (fabricante, conta)
);

CREATE INDEX IF NOT EXISTS idx_fp_receptor_presenca_fra
    ON fp_receptor_presenca (id_franqueado, fabricante);

CREATE TABLE IF NOT EXISTS fp_receptor_diag_log (
    id              BIGSERIAL PRIMARY KEY,
    id_franqueado   TEXT NOT NULL DEFAULT '',
    fabricante      TEXT NOT NULL,
    conta           TEXT NOT NULL,
    ip_tecnico      TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT '',
    conclusao       TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fp_receptor_diag_log_fra
    ON fp_receptor_diag_log (id_franqueado, created_at DESC);
