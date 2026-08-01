-- Corrige vinculo multi-tenant:
-- representante.IDCentralUUID / usuarios.IDCentralUUID / pacotes.IDCentralUUID / faturas.IDCentralUUID
-- passam a guardar central.ID_Central (ex.: 'CENTRAL' ou id gerado),
-- NAO o valor de central.IDCentralUUID (hex).
--
-- Rodar no MySQL apos deploy da API/webCentral.

-- 1) Representantes: se o valor atual for o UUID hex da central, troca pelo ID_Central
UPDATE representante r
INNER JOIN central c ON r.IDCentralUUID = c.IDCentralUUID
SET r.IDCentralUUID = c.ID_Central
WHERE NULLIF(TRIM(r.IDCentralUUID), '') IS NOT NULL
  AND r.IDCentralUUID <> c.ID_Central;

-- 2) Usuarios da Central (e demais) no mesmo criterio
UPDATE usuarios u
INNER JOIN central c ON u.IDCentralUUID = c.IDCentralUUID
SET u.IDCentralUUID = c.ID_Central
WHERE NULLIF(TRIM(u.IDCentralUUID), '') IS NOT NULL
  AND u.IDCentralUUID <> c.ID_Central;

-- 3) Pacotes
UPDATE pacotes p
INNER JOIN central c ON p.IDCentralUUID = c.IDCentralUUID
SET p.IDCentralUUID = c.ID_Central
WHERE NULLIF(TRIM(p.IDCentralUUID), '') IS NOT NULL
  AND p.IDCentralUUID <> c.ID_Central;

-- 4) Faturas
UPDATE faturas f
INNER JOIN central c ON f.IDCentralUUID = c.IDCentralUUID
SET f.IDCentralUUID = c.ID_Central
WHERE NULLIF(TRIM(f.IDCentralUUID), '') IS NOT NULL
  AND f.IDCentralUUID <> c.ID_Central;

-- 5) Legado sem UUID: garante CENTRAL
UPDATE representante
SET IDCentralUUID = 'CENTRAL'
WHERE IDCentralUUID IS NULL OR TRIM(IDCentralUUID) = '';

UPDATE usuarios
SET IDCentralUUID = 'CENTRAL'
WHERE ID_Vinculo = 'CENTRAL'
  AND (IDCentralUUID IS NULL OR TRIM(IDCentralUUID) = '');

UPDATE pacotes
SET IDCentralUUID = 'CENTRAL'
WHERE ID_Vinculo = 'CENTRAL'
  AND (IDCentralUUID IS NULL OR TRIM(IDCentralUUID) = '');

UPDATE faturas
SET IDCentralUUID = 'CENTRAL'
WHERE ID_Origem = 'CENTRAL'
  AND (IDCentralUUID IS NULL OR TRIM(IDCentralUUID) = '');

-- Conferir:
-- SELECT ID_Representante, RazaoSocial, IDCentralUUID FROM representante ORDER BY RazaoSocial;
-- SELECT ID_Central, IDCentralUUID, RazaoSocial FROM central;
