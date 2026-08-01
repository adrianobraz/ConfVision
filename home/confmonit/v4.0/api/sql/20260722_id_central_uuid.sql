-- Multi-tenant hibrido: mantem literal "CENTRAL" e adiciona IDCentralUUID
-- para identificar qual empresa Central (hoje so existe 1).
-- Rode no MySQL legado (phpMyAdmin ou mysql CLI) ANTES do deploy da API/webCentral.

-- 1) Colunas
ALTER TABLE central
  ADD COLUMN IDCentralUUID VARCHAR(64) NULL AFTER ID_Central;

ALTER TABLE usuarios
  ADD COLUMN IDCentralUUID VARCHAR(64) NULL AFTER ID_Vinculo;

ALTER TABLE representante
  ADD COLUMN IDCentralUUID VARCHAR(64) NULL AFTER ID_Representante;

ALTER TABLE faturas
  ADD COLUMN IDCentralUUID VARCHAR(64) NULL AFTER ID_Origem;

ALTER TABLE pacotes
  ADD COLUMN IDCentralUUID VARCHAR(64) NULL AFTER ID_Vinculo;

-- 2) UUID da Central atual (unica linha ID_Central = 'CENTRAL')
UPDATE central
SET IDCentralUUID = REPLACE(UUID(), '-', '')
WHERE ID_Central = 'CENTRAL'
  AND (IDCentralUUID IS NULL OR IDCentralUUID = '');

-- 3) Backfill legado: coluna IDCentralUUID nas filhas guarda central.ID_Central
UPDATE usuarios u
SET u.IDCentralUUID = 'CENTRAL'
WHERE u.ID_Vinculo = 'CENTRAL'
  AND (u.IDCentralUUID IS NULL OR u.IDCentralUUID = '');

UPDATE representante r
SET r.IDCentralUUID = 'CENTRAL'
WHERE (r.IDCentralUUID IS NULL OR r.IDCentralUUID = '');

UPDATE faturas f
SET f.IDCentralUUID = 'CENTRAL'
WHERE f.ID_Origem = 'CENTRAL'
  AND (f.IDCentralUUID IS NULL OR f.IDCentralUUID = '');

UPDATE pacotes p
SET p.IDCentralUUID = 'CENTRAL'
WHERE p.ID_Vinculo = 'CENTRAL'
  AND (p.IDCentralUUID IS NULL OR p.IDCentralUUID = '');

-- 4) Indices (opcional, ajuda filtros)
CREATE INDEX idx_usuarios_idcentraluuid ON usuarios (IDCentralUUID);
CREATE INDEX idx_representante_idcentraluuid ON representante (IDCentralUUID);
CREATE INDEX idx_faturas_idcentraluuid ON faturas (IDCentralUUID);
CREATE INDEX idx_pacotes_idcentraluuid ON pacotes (IDCentralUUID);
