-- Remove tabela Postgres duplicada; codigo interno passa a viver em cliente.CodigoInterno (MySQL).
-- Antes de aplicar: migrar manualmente dados de vis_cliente_ext se houver registros em producao.

DROP TABLE IF EXISTS vis_cliente_ext;
