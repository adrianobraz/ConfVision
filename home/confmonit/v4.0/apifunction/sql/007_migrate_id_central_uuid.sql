-- Fase 5: migrar ID_Central legado 'CENTRAL' para UUID da central matriz
-- Rodar apos 001_fp_contrato_schema.sql e 005_servicos_credito_schema.sql
-- MySQL confmonitV4 — substitui literal 'CENTRAL' por central.IDCentralUUID

-- fp_catalogo_produto
UPDATE fp_catalogo_produto t
INNER JOIN central c ON c.ID_Central = 'CENTRAL'
SET t.ID_Central = c.IDCentralUUID
WHERE t.ID_Central = 'CENTRAL'
  AND c.IDCentralUUID IS NOT NULL
  AND c.IDCentralUUID <> '';

-- fp_pacote_cota
UPDATE fp_pacote_cota t
INNER JOIN central c ON c.ID_Central = 'CENTRAL'
SET t.ID_Central = c.IDCentralUUID
WHERE t.ID_Central = 'CENTRAL'
  AND c.IDCentralUUID IS NOT NULL
  AND c.IDCentralUUID <> '';

-- fp_catalogo_cv_licenca
UPDATE fp_catalogo_cv_licenca t
INNER JOIN central c ON c.ID_Central = 'CENTRAL'
SET t.ID_Central = c.IDCentralUUID
WHERE t.ID_Central = 'CENTRAL'
  AND c.IDCentralUUID IS NOT NULL
  AND c.IDCentralUUID <> '';

-- fp_contrato
UPDATE fp_contrato t
INNER JOIN central c ON c.ID_Central = 'CENTRAL'
SET t.ID_Central = c.IDCentralUUID
WHERE t.ID_Central = 'CENTRAL'
  AND c.IDCentralUUID IS NOT NULL
  AND c.IDCentralUUID <> '';

-- fp_fatura_contrato
UPDATE fp_fatura_contrato t
INNER JOIN central c ON c.ID_Central = 'CENTRAL'
SET t.ID_Central = c.IDCentralUUID
WHERE t.ID_Central = 'CENTRAL'
  AND c.IDCentralUUID IS NOT NULL
  AND c.IDCentralUUID <> '';

-- fp_tarifa_operacional
UPDATE fp_tarifa_operacional t
INNER JOIN central c ON c.ID_Central = 'CENTRAL'
SET t.ID_Central = c.IDCentralUUID
WHERE t.ID_Central = 'CENTRAL'
  AND c.IDCentralUUID IS NOT NULL
  AND c.IDCentralUUID <> '';

-- fp_credito_recarga
UPDATE fp_credito_recarga t
INNER JOIN central c ON c.ID_Central = 'CENTRAL'
SET t.ID_Central = c.IDCentralUUID
WHERE t.ID_Central = 'CENTRAL'
  AND c.IDCentralUUID IS NOT NULL
  AND c.IDCentralUUID <> '';

-- ---------------------------------------------------------------------------
-- PostgreSQL (ConfVision / atendimento): ops_tarifa_operacional
-- Rodar separadamente no banco Postgres do ConfVision
-- ---------------------------------------------------------------------------

-- ops_tarifa_operacional: mesmo padrao, sintaxe Postgres (UPDATE ... FROM)
UPDATE ops_tarifa_operacional t
SET id_central = c.IDCentralUUID
FROM central c
WHERE t.id_central = 'CENTRAL'
  AND c.ID_Central = 'CENTRAL'
  AND c.IDCentralUUID IS NOT NULL
  AND TRIM(c.IDCentralUUID) <> '';
