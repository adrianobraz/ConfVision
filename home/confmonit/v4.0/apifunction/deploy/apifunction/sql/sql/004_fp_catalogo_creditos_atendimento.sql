-- Produtos de credito pre-pago (MySQL confmonitV4 / fp_catalogo_produto)
-- Valores iniciais zero; central ajusta no admConfmonit.

INSERT INTO fp_catalogo_produto (ID_Central, Produto, Plano, NomeExibicao, GrupoUI, TipoSelecao, ValorMensal, Ordem, Ativo)
VALUES
  ('CENTRAL', 'credito_ligacao', 'pacote', 'Crédito — Ligação', 'creditos', 'checkbox', 0, 100, 'S'),
  ('CENTRAL', 'credito_sms', 'pacote', 'Crédito — SMS', 'creditos', 'checkbox', 0, 101, 'S'),
  ('CENTRAL', 'credito_whatsapp', 'pacote', 'Crédito — WhatsApp', 'creditos', 'checkbox', 0, 102, 'S'),
  ('CENTRAL', 'credito_email', 'pacote', 'Crédito — E-mail', 'creditos', 'checkbox', 0, 103, 'S'),
  ('CENTRAL', 'finalizacao_auto_evento', 'modulo', 'Finalização automática (avulso)', 'atendimento', 'checkbox', 0, 110, 'S')
ON DUPLICATE KEY UPDATE
  NomeExibicao = VALUES(NomeExibicao),
  GrupoUI = VALUES(GrupoUI),
  Ativo = VALUES(Ativo);
