-- DNS receptor alarmes (FranqueadoPro)
-- Aplicado automaticamente pelo apifunction na subida (pgreceptordns.Migrate)

CREATE TABLE IF NOT EXISTS fp_receptor_dns (
    id_franqueado     TEXT PRIMARY KEY,
    subdominio        TEXT NOT NULL,
    fqdn              TEXT NOT NULL,
    zona              TEXT NOT NULL DEFAULT 'dnsid.com.br',
    ip_destino        TEXT NOT NULL,
    cf_record_id      TEXT,
    status            TEXT NOT NULL DEFAULT 'ativo',
    erro              TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_fp_receptor_dns_fqdn ON fp_receptor_dns (LOWER(fqdn));
CREATE UNIQUE INDEX IF NOT EXISTS uq_fp_receptor_dns_sub ON fp_receptor_dns (LOWER(subdominio), LOWER(zona));
