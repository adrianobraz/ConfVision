-- MySQL confmonit: coluna Master em receptorEvento (porta padrao FP)
-- Mesmo script do receptor-4/sql/003_receptor_evento_master.sql

ALTER TABLE receptorEvento
  ADD COLUMN Master CHAR(1) NOT NULL DEFAULT 'N'
  COMMENT 'S = porta padrao painel para cadastro alarme';

UPDATE receptorEvento SET Master = 'S' WHERE ID_ReceptorEvento IN (14, 16, 19, 30);
UPDATE receptorEvento SET Master = 'N' WHERE Master <> 'S' OR Master IS NULL;

UPDATE receptorEvento
SET Porta = '2093', Master = 'S', Ativo = 'S', Producao = 'S'
WHERE ID_ReceptorEvento = 30 AND Modulo = 'COMPATEC';
