-- Rodar no banco confservice se a tabela ja existir (indice do relatorio)
ALTER TABLE cs_webhook_fila
  ADD INDEX idx_cs_fila_parceiro_created (id_parceiro, created_at);
