-- Gerado por export_xano.ps1 em 2026-08-07T20:03:38.2439309-03:00
BEGIN;

INSERT INTO vis_mediamtx_node (
    id, created_at, nome, rtmp_public, hls_public, rtsp_internal,
    max_cameras, ordem, status, observacao
) VALUES (
    1, '2026-08-01 02:47:44+00', 'servidor1', 'rtmp://srv1.dnsid.com.br:1935',
    'https://hls1.dnsid.com.br', 'rtsp://foxpro_confvision:8554', 200,
    1, 'ativo',
    NULL
) ON CONFLICT (id) DO UPDATE SET
    nome = EXCLUDED.nome,
    rtmp_public = EXCLUDED.rtmp_public,
    hls_public = EXCLUDED.hls_public,
    rtsp_internal = EXCLUDED.rtsp_internal,
    max_cameras = EXCLUDED.max_cameras,
    ordem = EXCLUDED.ordem,
    status = EXCLUDED.status,
    observacao = EXCLUDED.observacao;
INSERT INTO vis_mediamtx_node (
    id, created_at, nome, rtmp_public, hls_public, rtsp_internal,
    max_cameras, ordem, status, observacao
) VALUES (
    2, '2026-08-01 02:52:51+00', 'servidor2', 'rtmp://srv2.dnsid.com.br:1935',
    'https://hls2.dnsid.com.br', 'rtsp://foxpro_confvision:8554', 200,
    2, 'inativo',
    'VPS2 ainda nao provisionada'
) ON CONFLICT (id) DO UPDATE SET
    nome = EXCLUDED.nome,
    rtmp_public = EXCLUDED.rtmp_public,
    hls_public = EXCLUDED.hls_public,
    rtsp_internal = EXCLUDED.rtsp_internal,
    max_cameras = EXCLUDED.max_cameras,
    ordem = EXCLUDED.ordem,
    status = EXCLUDED.status,
    observacao = EXCLUDED.observacao;
INSERT INTO vis_mediamtx_node (
    id, created_at, nome, rtmp_public, hls_public, rtsp_internal,
    max_cameras, ordem, status, observacao
) VALUES (
    3, '2026-08-01 02:58:32+00', 'servidor3', 'rtmp://srv3.dnsid.com.br:1935',
    'https://hls3.dnsid.com.br', 'rtsp://foxpro_confvision:8554', 200,
    3, 'inativo',
    'VPS3 ainda nao provisionada'
) ON CONFLICT (id) DO UPDATE SET
    nome = EXCLUDED.nome,
    rtmp_public = EXCLUDED.rtmp_public,
    hls_public = EXCLUDED.hls_public,
    rtsp_internal = EXCLUDED.rtsp_internal,
    max_cameras = EXCLUDED.max_cameras,
    ordem = EXCLUDED.ordem,
    status = EXCLUDED.status,
    observacao = EXCLUDED.observacao;
INSERT INTO vis_mediamtx_node (
    id, created_at, nome, rtmp_public, hls_public, rtsp_internal,
    max_cameras, ordem, status, observacao
) VALUES (
    4, '2026-08-01 02:58:48+00', 'servidor4', 'rtmp://srv4.dnsid.com.br:1935',
    'https://hls4.dnsid.com.br', 'rtsp://foxpro_confvision:8554', 200,
    4, 'inativo',
    'VPS4 ainda nao provisionada'
) ON CONFLICT (id) DO UPDATE SET
    nome = EXCLUDED.nome,
    rtmp_public = EXCLUDED.rtmp_public,
    hls_public = EXCLUDED.hls_public,
    rtsp_internal = EXCLUDED.rtsp_internal,
    max_cameras = EXCLUDED.max_cameras,
    ordem = EXCLUDED.ordem,
    status = EXCLUDED.status,
    observacao = EXCLUDED.observacao;
INSERT INTO vis_mediamtx_node (
    id, created_at, nome, rtmp_public, hls_public, rtsp_internal,
    max_cameras, ordem, status, observacao
) VALUES (
    5, '2026-08-01 02:59:04+00', 'servidor5', 'rtmp://srv5.dnsid.com.br:1935',
    'https://hls5.dnsid.com.br', 'rtsp://foxpro_confvision:8554', 200,
    5, 'inativo',
    'VPS5 ainda nao provisionada'
) ON CONFLICT (id) DO UPDATE SET
    nome = EXCLUDED.nome,
    rtmp_public = EXCLUDED.rtmp_public,
    hls_public = EXCLUDED.hls_public,
    rtsp_internal = EXCLUDED.rtsp_internal,
    max_cameras = EXCLUDED.max_cameras,
    ordem = EXCLUDED.ordem,
    status = EXCLUDED.status,
    observacao = EXCLUDED.observacao;

INSERT INTO vis_camera (
    id, created_at, ativo, bloqueado, nome, id_franqueado, id_cliente, id_dispositivo,
    conta, particao, canal, setor, protocolo, rtsp_url_sec, onvif_host, onvif_porta,
    onvif_usuario, onvif_senha, confianca_min, cooldown_seg, somente_armado,
    deteccao_humano, deteccao_veiculo, status, ultimo_evento_em, worker_id,
    ultimo_ping_em, zonauser, captura_sensor, captura_analitico, analitico_pausado,
    id_setor, snapshot_url, vis_licenca_id, plano, ativado_em, evento_grava_foto,
    evento_grava_video, vis_licenca_gravacao_id, grava_continua, retencao_dias,
    gravacao_ativada_em, gravacao_status, grava_movimento, grava_timelapse,
    gravacao_flush_pedido, modo_deteccao, vis_mediamtx_node_id
) VALUES (
    1, '2026-08-07 22:49:28+00', FALSE, FALSE,
    'Entrada Principal', '2025021002461524550353102', '2025120201391858163117728', '2025120201412129271953361',
    '0025', '1', '01', 'BARRERIA FRENTE',
    'wifi', NULL, NULL, 0,
    NULL, NULL, 0.5, 30,
    FALSE, FALSE, FALSE,
    'offline', NULL, NULL,
    NULL, '001', FALSE,
    FALSE, FALSE, '2025120505123887788999922',
    NULL, 1, 'online', NULL,
    FALSE, FALSE, NULL,
    FALSE, NULL, NULL,
    NULL, FALSE, FALSE,
    FALSE, 'dentro', 1
) ON CONFLICT (id) DO NOTHING;


INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    60, '2026-08-07 22:47:35+00', '2025021002461524550353102', 'gravacao_timelapse_30d',
    'gravacao', 11.19,
    '2026-08-07 22:47:53+00', '2026-09-06 22:47:53+00', 'disponivel',
    NULL, '55', '27',
    'Aguardando pagamento — Gravacao timelapse inteligente 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    59, '2026-08-07 22:47:35+00', '2025021002461524550353102', 'gravacao_timelapse_30d',
    'gravacao', 11.19,
    '2026-08-07 22:47:53+00', '2026-09-06 22:47:53+00', 'disponivel',
    NULL, '55', '27',
    'Aguardando pagamento — Gravacao timelapse inteligente 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    58, '2026-08-07 22:47:35+00', '2025021002461524550353102', 'gravacao_timelapse_30d',
    'gravacao', 11.19,
    '2026-08-07 22:47:53+00', '2026-09-06 22:47:53+00', 'disponivel',
    NULL, '55', '27',
    'Aguardando pagamento — Gravacao timelapse inteligente 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    57, '2026-08-07 22:47:35+00', '2025021002461524550353102', 'gravacao_timelapse_30d',
    'gravacao', 11.19,
    '2026-08-07 22:47:53+00', '2026-09-06 22:47:53+00', 'disponivel',
    NULL, '55', '27',
    'Aguardando pagamento — Gravacao timelapse inteligente 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    56, '2026-08-07 22:47:35+00', '2025021002461524550353102', 'gravacao_timelapse_30d',
    'gravacao', 11.19,
    '2026-08-07 22:47:53+00', '2026-09-06 22:47:53+00', 'disponivel',
    NULL, '55', '27',
    'Aguardando pagamento — Gravacao timelapse inteligente 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    55, '2026-08-07 22:47:35+00', '2025021002461524550353102', 'gravacao_timelapse_30d',
    'gravacao', 11.19,
    '2026-08-07 22:47:53+00', '2026-09-06 22:47:53+00', 'disponivel',
    NULL, '55', '27',
    'Aguardando pagamento — Gravacao timelapse inteligente 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    54, '2026-08-07 22:47:35+00', '2025021002461524550353102', 'gravacao_timelapse_30d',
    'gravacao', 11.19,
    '2026-08-07 22:47:53+00', '2026-09-06 22:47:53+00', 'disponivel',
    NULL, '55', '27',
    'Aguardando pagamento — Gravacao timelapse inteligente 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    53, '2026-08-07 22:47:35+00', '2025021002461524550353102', 'gravacao_timelapse_30d',
    'gravacao', 11.19,
    '2026-08-07 22:47:53+00', '2026-09-06 22:47:53+00', 'disponivel',
    NULL, '55', '27',
    'Aguardando pagamento — Gravacao timelapse inteligente 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    52, '2026-08-07 22:47:35+00', '2025021002461524550353102', 'gravacao_timelapse_30d',
    'gravacao', 11.19,
    '2026-08-07 22:47:53+00', '2026-09-06 22:47:53+00', 'disponivel',
    NULL, '55', '27',
    'Aguardando pagamento — Gravacao timelapse inteligente 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    51, '2026-08-07 22:47:35+00', '2025021002461524550353102', 'gravacao_timelapse_30d',
    'gravacao', 11.19,
    '2026-08-07 22:47:53+00', '2026-09-06 22:47:53+00', 'disponivel',
    NULL, '55', '27',
    'Aguardando pagamento — Gravacao timelapse inteligente 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    50, '2026-08-07 22:47:25+00', '2025021002461524550353102', 'gravacao_movimento_30d',
    'gravacao', 15.99,
    '2026-08-07 22:48:00+00', '2026-09-06 22:48:01+00', 'disponivel',
    NULL, '54', '28',
    'Aguardando pagamento — Gravacao por movimento 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    49, '2026-08-07 22:47:25+00', '2025021002461524550353102', 'gravacao_movimento_30d',
    'gravacao', 15.99,
    '2026-08-07 22:48:00+00', '2026-09-06 22:48:01+00', 'disponivel',
    NULL, '54', '28',
    'Aguardando pagamento — Gravacao por movimento 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    48, '2026-08-07 22:47:25+00', '2025021002461524550353102', 'gravacao_movimento_30d',
    'gravacao', 15.99,
    '2026-08-07 22:48:00+00', '2026-09-06 22:48:01+00', 'disponivel',
    NULL, '54', '28',
    'Aguardando pagamento — Gravacao por movimento 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    47, '2026-08-07 22:47:25+00', '2025021002461524550353102', 'gravacao_movimento_30d',
    'gravacao', 15.99,
    '2026-08-07 22:48:00+00', '2026-09-06 22:48:01+00', 'disponivel',
    NULL, '54', '28',
    'Aguardando pagamento — Gravacao por movimento 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    46, '2026-08-07 22:47:25+00', '2025021002461524550353102', 'gravacao_movimento_30d',
    'gravacao', 15.99,
    '2026-08-07 22:48:00+00', '2026-09-06 22:48:01+00', 'disponivel',
    NULL, '54', '28',
    'Aguardando pagamento — Gravacao por movimento 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    45, '2026-08-07 22:47:25+00', '2025021002461524550353102', 'gravacao_movimento_30d',
    'gravacao', 15.99,
    '2026-08-07 22:48:00+00', '2026-09-06 22:48:01+00', 'disponivel',
    NULL, '54', '28',
    'Aguardando pagamento — Gravacao por movimento 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    44, '2026-08-07 22:47:25+00', '2025021002461524550353102', 'gravacao_movimento_30d',
    'gravacao', 15.99,
    '2026-08-07 22:48:00+00', '2026-09-06 22:48:01+00', 'disponivel',
    NULL, '54', '28',
    'Aguardando pagamento — Gravacao por movimento 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    43, '2026-08-07 22:47:25+00', '2025021002461524550353102', 'gravacao_movimento_30d',
    'gravacao', 15.99,
    '2026-08-07 22:48:00+00', '2026-09-06 22:48:01+00', 'disponivel',
    NULL, '54', '28',
    'Aguardando pagamento — Gravacao por movimento 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    42, '2026-08-07 22:47:25+00', '2025021002461524550353102', 'gravacao_movimento_30d',
    'gravacao', 15.99,
    '2026-08-07 22:48:00+00', '2026-09-06 22:48:01+00', 'disponivel',
    NULL, '54', '28',
    'Aguardando pagamento — Gravacao por movimento 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    41, '2026-08-07 22:47:25+00', '2025021002461524550353102', 'gravacao_movimento_30d',
    'gravacao', 15.99,
    '2026-08-07 22:48:00+00', '2026-09-06 22:48:01+00', 'disponivel',
    NULL, '54', '28',
    'Aguardando pagamento — Gravacao por movimento 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    40, '2026-08-07 22:47:14+00', '2025021002461524550353102', 'gravacao_30d',
    'gravacao', 19.99,
    '2026-08-07 22:48:08+00', '2026-09-06 22:48:09+00', 'disponivel',
    NULL, '53', '29',
    'Aguardando pagamento — Gravacao continua 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    39, '2026-08-07 22:47:14+00', '2025021002461524550353102', 'gravacao_30d',
    'gravacao', 19.99,
    '2026-08-07 22:48:08+00', '2026-09-06 22:48:09+00', 'disponivel',
    NULL, '53', '29',
    'Aguardando pagamento — Gravacao continua 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    38, '2026-08-07 22:47:14+00', '2025021002461524550353102', 'gravacao_30d',
    'gravacao', 19.99,
    '2026-08-07 22:48:08+00', '2026-09-06 22:48:09+00', 'disponivel',
    NULL, '53', '29',
    'Aguardando pagamento — Gravacao continua 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    37, '2026-08-07 22:47:14+00', '2025021002461524550353102', 'gravacao_30d',
    'gravacao', 19.99,
    '2026-08-07 22:48:08+00', '2026-09-06 22:48:09+00', 'disponivel',
    NULL, '53', '29',
    'Aguardando pagamento — Gravacao continua 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    36, '2026-08-07 22:47:14+00', '2025021002461524550353102', 'gravacao_30d',
    'gravacao', 19.99,
    '2026-08-07 22:48:08+00', '2026-09-06 22:48:09+00', 'disponivel',
    NULL, '53', '29',
    'Aguardando pagamento — Gravacao continua 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    35, '2026-08-07 22:47:14+00', '2025021002461524550353102', 'gravacao_30d',
    'gravacao', 19.99,
    '2026-08-07 22:48:08+00', '2026-09-06 22:48:09+00', 'disponivel',
    NULL, '53', '29',
    'Aguardando pagamento — Gravacao continua 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    34, '2026-08-07 22:47:14+00', '2025021002461524550353102', 'gravacao_30d',
    'gravacao', 19.99,
    '2026-08-07 22:48:08+00', '2026-09-06 22:48:09+00', 'disponivel',
    NULL, '53', '29',
    'Aguardando pagamento — Gravacao continua 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    33, '2026-08-07 22:47:14+00', '2025021002461524550353102', 'gravacao_30d',
    'gravacao', 19.99,
    '2026-08-07 22:48:08+00', '2026-09-06 22:48:09+00', 'disponivel',
    NULL, '53', '29',
    'Aguardando pagamento — Gravacao continua 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    32, '2026-08-07 22:47:14+00', '2025021002461524550353102', 'gravacao_30d',
    'gravacao', 19.99,
    '2026-08-07 22:48:08+00', '2026-09-06 22:48:09+00', 'disponivel',
    NULL, '53', '29',
    'Aguardando pagamento — Gravacao continua 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    31, '2026-08-07 22:47:14+00', '2025021002461524550353102', 'gravacao_30d',
    'gravacao', 19.99,
    '2026-08-07 22:48:08+00', '2026-09-06 22:48:08+00', 'disponivel',
    NULL, '53', '29',
    'Aguardando pagamento — Gravacao continua 30 dias', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    30, '2026-08-07 22:47:01+00', '2025021002461524550353102', 'analitico_24h_foto_video',
    'camera', 15.99,
    '2026-08-07 22:48:17+00', '2026-09-06 22:48:18+00', 'disponivel',
    NULL, '52', '30',
    'Aguardando pagamento — Analitico 24h — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    29, '2026-08-07 22:47:01+00', '2025021002461524550353102', 'analitico_24h_foto_video',
    'camera', 15.99,
    '2026-08-07 22:48:17+00', '2026-09-06 22:48:18+00', 'disponivel',
    NULL, '52', '30',
    'Aguardando pagamento — Analitico 24h — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    28, '2026-08-07 22:47:01+00', '2025021002461524550353102', 'analitico_24h_foto_video',
    'camera', 15.99,
    '2026-08-07 22:48:17+00', '2026-09-06 22:48:18+00', 'disponivel',
    NULL, '52', '30',
    'Aguardando pagamento — Analitico 24h — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    27, '2026-08-07 22:47:01+00', '2025021002461524550353102', 'analitico_24h_foto_video',
    'camera', 15.99,
    '2026-08-07 22:48:17+00', '2026-09-06 22:48:18+00', 'disponivel',
    NULL, '52', '30',
    'Aguardando pagamento — Analitico 24h — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    26, '2026-08-07 22:47:01+00', '2025021002461524550353102', 'analitico_24h_foto_video',
    'camera', 15.99,
    '2026-08-07 22:48:17+00', '2026-09-06 22:48:18+00', 'disponivel',
    NULL, '52', '30',
    'Aguardando pagamento — Analitico 24h — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    25, '2026-08-07 22:47:01+00', '2025021002461524550353102', 'analitico_24h_foto_video',
    'camera', 15.99,
    '2026-08-07 22:48:17+00', '2026-09-06 22:48:18+00', 'disponivel',
    NULL, '52', '30',
    'Aguardando pagamento — Analitico 24h — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    24, '2026-08-07 22:47:01+00', '2025021002461524550353102', 'analitico_24h_foto_video',
    'camera', 15.99,
    '2026-08-07 22:48:17+00', '2026-09-06 22:48:18+00', 'disponivel',
    NULL, '52', '30',
    'Aguardando pagamento — Analitico 24h — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    23, '2026-08-07 22:47:01+00', '2025021002461524550353102', 'analitico_24h_foto_video',
    'camera', 15.99,
    '2026-08-07 22:48:17+00', '2026-09-06 22:48:18+00', 'disponivel',
    NULL, '52', '30',
    'Aguardando pagamento — Analitico 24h — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    22, '2026-08-07 22:47:01+00', '2025021002461524550353102', 'analitico_24h_foto_video',
    'camera', 15.99,
    '2026-08-07 22:48:17+00', '2026-09-06 22:48:17+00', 'disponivel',
    NULL, '52', '30',
    'Aguardando pagamento — Analitico 24h — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    21, '2026-08-07 22:47:01+00', '2025021002461524550353102', 'analitico_24h_foto_video',
    'camera', 15.99,
    '2026-08-07 22:48:17+00', '2026-09-06 22:48:17+00', 'disponivel',
    NULL, '52', '30',
    'Aguardando pagamento — Analitico 24h — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    20, '2026-08-07 22:46:49+00', '2025021002461524550353102', 'analitico_armado_foto_video',
    'camera', 11.99,
    '2026-08-07 22:48:26+00', '2026-09-06 22:48:26+00', 'disponivel',
    NULL, '51', '31',
    'Aguardando pagamento — Analitico armado — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    19, '2026-08-07 22:46:49+00', '2025021002461524550353102', 'analitico_armado_foto_video',
    'camera', 11.99,
    '2026-08-07 22:48:26+00', '2026-09-06 22:48:26+00', 'disponivel',
    NULL, '51', '31',
    'Aguardando pagamento — Analitico armado — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    18, '2026-08-07 22:46:49+00', '2025021002461524550353102', 'analitico_armado_foto_video',
    'camera', 11.99,
    '2026-08-07 22:48:26+00', '2026-09-06 22:48:26+00', 'disponivel',
    NULL, '51', '31',
    'Aguardando pagamento — Analitico armado — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    17, '2026-08-07 22:46:49+00', '2025021002461524550353102', 'analitico_armado_foto_video',
    'camera', 11.99,
    '2026-08-07 22:48:26+00', '2026-09-06 22:48:26+00', 'disponivel',
    NULL, '51', '31',
    'Aguardando pagamento — Analitico armado — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    16, '2026-08-07 22:46:49+00', '2025021002461524550353102', 'analitico_armado_foto_video',
    'camera', 11.99,
    '2026-08-07 22:48:26+00', '2026-09-06 22:48:26+00', 'disponivel',
    NULL, '51', '31',
    'Aguardando pagamento — Analitico armado — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    15, '2026-08-07 22:46:49+00', '2025021002461524550353102', 'analitico_armado_foto_video',
    'camera', 11.99,
    '2026-08-07 22:48:26+00', '2026-09-06 22:48:26+00', 'disponivel',
    NULL, '51', '31',
    'Aguardando pagamento — Analitico armado — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    14, '2026-08-07 22:46:49+00', '2025021002461524550353102', 'analitico_armado_foto_video',
    'camera', 11.99,
    '2026-08-07 22:48:26+00', '2026-09-06 22:48:26+00', 'disponivel',
    NULL, '51', '31',
    'Aguardando pagamento — Analitico armado — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    13, '2026-08-07 22:46:49+00', '2025021002461524550353102', 'analitico_armado_foto_video',
    'camera', 11.99,
    '2026-08-07 22:48:26+00', '2026-09-06 22:48:26+00', 'disponivel',
    NULL, '51', '31',
    'Aguardando pagamento — Analitico armado — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    12, '2026-08-07 22:46:49+00', '2025021002461524550353102', 'analitico_armado_foto_video',
    'camera', 11.99,
    '2026-08-07 22:48:26+00', '2026-09-06 22:48:26+00', 'disponivel',
    NULL, '51', '31',
    'Aguardando pagamento — Analitico armado — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    11, '2026-08-07 22:46:49+00', '2025021002461524550353102', 'analitico_armado_foto_video',
    'camera', 11.99,
    '2026-08-07 22:48:26+00', '2026-09-06 22:48:26+00', 'disponivel',
    NULL, '51', '31',
    'Aguardando pagamento — Analitico armado — foto + video', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    10, '2026-08-07 22:46:36+00', '2025021002461524550353102', 'online',
    'camera', 2.39,
    '2026-08-07 22:48:33+00', '2026-09-06 22:48:33+00', 'disponivel',
    NULL, '50', '32',
    'Aguardando pagamento — Camera online', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    9, '2026-08-07 22:46:36+00', '2025021002461524550353102', 'online',
    'camera', 2.39,
    '2026-08-07 22:48:33+00', '2026-09-06 22:48:33+00', 'disponivel',
    NULL, '50', '32',
    'Aguardando pagamento — Camera online', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    8, '2026-08-07 22:46:36+00', '2025021002461524550353102', 'online',
    'camera', 2.39,
    '2026-08-07 22:48:33+00', '2026-09-06 22:48:33+00', 'disponivel',
    NULL, '50', '32',
    'Aguardando pagamento — Camera online', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    7, '2026-08-07 22:46:36+00', '2025021002461524550353102', 'online',
    'camera', 2.39,
    '2026-08-07 22:48:33+00', '2026-09-06 22:48:33+00', 'disponivel',
    NULL, '50', '32',
    'Aguardando pagamento — Camera online', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    6, '2026-08-07 22:46:36+00', '2025021002461524550353102', 'online',
    'camera', 2.39,
    '2026-08-07 22:48:33+00', '2026-09-06 22:48:33+00', 'disponivel',
    NULL, '50', '32',
    'Aguardando pagamento — Camera online', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    5, '2026-08-07 22:46:36+00', '2025021002461524550353102', 'online',
    'camera', 2.39,
    '2026-08-07 22:48:33+00', '2026-09-06 22:48:33+00', 'disponivel',
    NULL, '50', '32',
    'Aguardando pagamento — Camera online', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    4, '2026-08-07 22:46:36+00', '2025021002461524550353102', 'online',
    'camera', 2.39,
    '2026-08-07 22:48:33+00', '2026-09-06 22:48:33+00', 'disponivel',
    NULL, '50', '32',
    'Aguardando pagamento — Camera online', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    3, '2026-08-07 22:46:36+00', '2025021002461524550353102', 'online',
    'camera', 2.39,
    '2026-08-07 22:48:33+00', '2026-09-06 22:48:33+00', 'disponivel',
    NULL, '50', '32',
    'Aguardando pagamento — Camera online', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    2, '2026-08-07 22:46:36+00', '2025021002461524550353102', 'online',
    'camera', 2.39,
    '2026-08-07 22:48:33+00', '2026-09-06 22:48:33+00', 'disponivel',
    NULL, '50', '32',
    'Aguardando pagamento — Camera online', NULL
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    1, '2026-08-07 22:46:36+00', '2025021002461524550353102', 'online',
    'camera', 2.39,
    '2026-08-07 22:48:33+00', '2026-09-06 22:48:33+00', 'em_uso',
    '2025120201412129271953361', '50', '32',
    'Aguardando pagamento — Camera online', 1
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;

SELECT setval(pg_get_serial_sequence('vis_mediamtx_node','id'), COALESCE((SELECT MAX(id) FROM vis_mediamtx_node), 1));
SELECT setval(pg_get_serial_sequence('vis_camera','id'), COALESCE((SELECT MAX(id) FROM vis_camera), 1));
SELECT setval(pg_get_serial_sequence('vis_camera_area','id'), COALESCE((SELECT MAX(id) FROM vis_camera_area), 1));
SELECT setval(pg_get_serial_sequence('vis_licenca','id'), COALESCE((SELECT MAX(id) FROM vis_licenca), 1));
SELECT vis_mediamtx_node_recalc_pontos(id) FROM vis_mediamtx_node;
COMMIT;
