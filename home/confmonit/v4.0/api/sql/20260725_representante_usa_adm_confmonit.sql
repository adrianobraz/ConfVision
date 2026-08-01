-- Flag: representante opera o admConfmonit (carteira propria) ou Central libera.
-- Padrao N = Central destaca/libera na Central de Liberacao.

ALTER TABLE representante
  ADD COLUMN UsaAdmConfmonit CHAR(1) NOT NULL DEFAULT 'N';

-- Se a coluna ja existir sem default / valores vazios:
-- UPDATE representante SET UsaAdmConfmonit = 'N'
-- WHERE UsaAdmConfmonit IS NULL OR TRIM(UsaAdmConfmonit) = '';
