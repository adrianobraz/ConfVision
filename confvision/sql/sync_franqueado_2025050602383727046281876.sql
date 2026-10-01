-- Sync vis_licenca Xano -> Postgres (gerado automaticamente)
BEGIN;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (34, '2026-08-18 21:34:40+00', '2025050602383727046281876', 'analitico_armado_foto', 'camera', 11.19, '2026-08-18 21:35:22+00', '2026-09-17 21:35:22+00', 'disponivel', '87', '77', 'Aguardando pagamento ÔÇö Analitico armado ÔÇö foto')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (33, '2026-08-18 21:34:40+00', '2025050602383727046281876', 'analitico_armado_foto', 'camera', 11.19, '2026-08-18 21:35:22+00', '2026-09-17 21:35:22+00', 'disponivel', '87', '77', 'Aguardando pagamento ÔÇö Analitico armado ÔÇö foto')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (32, '2026-08-18 21:34:40+00', '2025050602383727046281876', 'analitico_armado_foto', 'camera', 11.19, '2026-08-18 21:35:22+00', '2026-09-17 21:35:22+00', 'disponivel', '87', '77', 'Aguardando pagamento ÔÇö Analitico armado ÔÇö foto')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (31, '2026-08-18 21:34:27+00', '2025050602383727046281876', 'analitico_24h_foto_video', 'camera', 15.99, '2026-08-18 21:35:27+00', '2026-09-17 21:35:27+00', 'disponivel', '86', '78', 'Aguardando pagamento ÔÇö Analitico 24h ÔÇö foto + video')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (30, '2026-08-18 21:34:27+00', '2025050602383727046281876', 'analitico_24h_foto_video', 'camera', 15.99, '2026-08-18 21:35:27+00', '2026-09-17 21:35:27+00', 'disponivel', '86', '78', 'Aguardando pagamento ÔÇö Analitico 24h ÔÇö foto + video')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (29, '2026-08-18 21:34:27+00', '2025050602383727046281876', 'analitico_24h_foto_video', 'camera', 15.99, '2026-08-18 21:35:27+00', '2026-09-17 21:35:27+00', 'disponivel', '86', '78', 'Aguardando pagamento ÔÇö Analitico 24h ÔÇö foto + video')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (28, '2026-08-18 21:34:17+00', '2025050602383727046281876', 'analitico_24h_evento', 'camera', 13.59, '2026-08-18 21:35:33+00', '2026-09-17 21:35:33+00', 'disponivel', '85', '79', 'Aguardando pagamento ÔÇö Analitico 24h ÔÇö so evento')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (27, '2026-08-18 21:34:17+00', '2025050602383727046281876', 'analitico_24h_evento', 'camera', 13.59, '2026-08-18 21:35:33+00', '2026-09-17 21:35:33+00', 'disponivel', '85', '79', 'Aguardando pagamento ÔÇö Analitico 24h ÔÇö so evento')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (26, '2026-08-18 21:34:17+00', '2025050602383727046281876', 'analitico_24h_evento', 'camera', 13.59, '2026-08-18 21:35:33+00', '2026-09-17 21:35:33+00', 'disponivel', '85', '79', 'Aguardando pagamento ÔÇö Analitico 24h ÔÇö so evento')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (25, '2026-08-18 21:33:48+00', '2025050602383727046281876', 'analitico_armado_foto_video', 'camera', 11.99, '2026-08-18 21:35:38+00', '2026-09-17 21:35:38+00', 'disponivel', '84', '80', 'Aguardando pagamento ÔÇö Analitico armado ÔÇö foto + video')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (24, '2026-08-18 21:33:48+00', '2025050602383727046281876', 'analitico_armado_foto_video', 'camera', 11.99, '2026-08-18 21:35:38+00', '2026-09-17 21:35:38+00', 'disponivel', '84', '80', 'Aguardando pagamento ÔÇö Analitico armado ÔÇö foto + video')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (23, '2026-08-18 21:33:48+00', '2025050602383727046281876', 'analitico_armado_foto_video', 'camera', 11.99, '2026-08-18 21:35:38+00', '2026-09-17 21:35:38+00', 'disponivel', '84', '80', 'Aguardando pagamento ÔÇö Analitico armado ÔÇö foto + video')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (22, '2026-08-18 21:33:37+00', '2025050602383727046281876', 'analitico_armado_evento', 'camera', 9.59, '2026-08-18 21:35:41+00', '2026-09-17 21:35:41+00', 'disponivel', '83', '81', 'Aguardando pagamento ÔÇö Analitico armado ÔÇö so evento')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (21, '2026-08-18 21:33:37+00', '2025050602383727046281876', 'analitico_armado_evento', 'camera', 9.59, '2026-08-18 21:35:41+00', '2026-09-17 21:35:41+00', 'disponivel', '83', '81', 'Aguardando pagamento ÔÇö Analitico armado ÔÇö so evento')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (20, '2026-08-18 21:33:37+00', '2025050602383727046281876', 'analitico_armado_evento', 'camera', 9.59, '2026-08-18 21:35:41+00', '2026-09-17 21:35:41+00', 'disponivel', '83', '81', 'Aguardando pagamento ÔÇö Analitico armado ÔÇö so evento')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (19, '2026-08-18 21:33:22+00', '2025050602383727046281876', 'sensor_foto_video', 'camera', 7.99, '2026-08-18 21:35:46+00', '2026-09-17 21:35:46+00', 'disponivel', '82', '82', 'Aguardando pagamento ÔÇö Sensor foto + video')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (18, '2026-08-18 21:33:22+00', '2025050602383727046281876', 'sensor_foto_video', 'camera', 7.99, '2026-08-18 21:35:46+00', '2026-09-17 21:35:46+00', 'disponivel', '82', '82', 'Aguardando pagamento ÔÇö Sensor foto + video')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (17, '2026-08-18 21:33:22+00', '2025050602383727046281876', 'sensor_foto_video', 'camera', 7.99, '2026-08-18 21:35:46+00', '2026-09-17 21:35:46+00', 'disponivel', '82', '82', 'Aguardando pagamento ÔÇö Sensor foto + video')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (16, '2026-08-18 21:33:14+00', '2025050602383727046281876', 'sensor_foto', 'camera', 6.39, '2026-08-18 21:35:50+00', '2026-09-17 21:35:50+00', 'disponivel', '81', '83', 'Aguardando pagamento ÔÇö Sensor foto')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (15, '2026-08-18 21:33:14+00', '2025050602383727046281876', 'sensor_foto', 'camera', 6.39, '2026-08-18 21:35:50+00', '2026-09-17 21:35:50+00', 'disponivel', '81', '83', 'Aguardando pagamento ÔÇö Sensor foto')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (14, '2026-08-18 21:33:14+00', '2025050602383727046281876', 'sensor_foto', 'camera', 6.39, '2026-08-18 21:35:50+00', '2026-09-17 21:35:50+00', 'disponivel', '81', '83', 'Aguardando pagamento ÔÇö Sensor foto')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (13, '2026-08-18 21:32:57+00', '2025050602383727046281876', 'online', 'camera', 2.39, '2026-08-18 21:35:55+00', '2026-09-17 21:35:55+00', 'disponivel', '80', '84', 'Aguardando pagamento ÔÇö Camera online')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (12, '2026-08-18 21:32:57+00', '2025050602383727046281876', 'online', 'camera', 2.39, '2026-08-18 21:35:55+00', '2026-09-17 21:35:55+00', 'disponivel', '80', '84', 'Aguardando pagamento ÔÇö Camera online')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (11, '2026-08-18 21:32:57+00', '2025050602383727046281876', 'online', 'camera', 2.39, '2026-08-18 21:35:55+00', '2026-09-17 21:35:55+00', 'disponivel', '80', '84', 'Aguardando pagamento ÔÇö Camera online')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;
SELECT setval(pg_get_serial_sequence('vis_licenca','id'), COALESCE((SELECT MAX(id) FROM vis_licenca), 1));
COMMIT;
