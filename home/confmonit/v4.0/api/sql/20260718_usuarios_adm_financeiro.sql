-- Flag de acesso ao admConfmonit (financeiro hierárquico CEN/REP).
-- Rode no MySQL legado antes de usar o login por usuarios.

ALTER TABLE usuarios
  ADD COLUMN AdmFinanceiro CHAR(1) NOT NULL DEFAULT 'N'
  AFTER Master;

-- Exemplo: liberar um master da Central
-- UPDATE usuarios SET AdmFinanceiro = 'S' WHERE ID_Vinculo = 'CENTRAL' AND Master = 'S' AND Email1 = 'seu@email.com';

-- Exemplo: liberar um master Representante
-- UPDATE usuarios SET AdmFinanceiro = 'S' WHERE Master = 'S' AND AdmFinanceiro = 'N' AND ID_Vinculo = '<ID_REPRESENTANTE>';
