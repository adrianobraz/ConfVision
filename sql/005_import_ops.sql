-- Gerado por export_ops_xano.ps1 em 2026-08-08T14:32:22.1317926-03:00
BEGIN;

INSERT INTO ops_arme_janela (id, created_at, whatsappeventocad_id, dias, hora_inicio, hora_fim, id_cliente, id_franqueado, whatsapp, nome, id_dispositivo, nome_dispositivo) VALUES (4, '2026-06-02 00:16:36+00', 1944, ARRAY['0'], '00:00', '20:00', '2025101305062894595132374', '2025101001181399925723224', '5511949294764', 'Thiago Sciammarella', '2026060103444511408868385', 'QUARESMEIRAS CAMPINAS ENGENHARIA - ESTOQUE') ON CONFLICT (id) DO UPDATE SET dias = EXCLUDED.dias, hora_inicio = EXCLUDED.hora_inicio, hora_fim = EXCLUDED.hora_fim;

SELECT setval(pg_get_serial_sequence('vis_cliente_grade_config','id'), COALESCE((SELECT MAX(id) FROM vis_cliente_grade_config), 1));
SELECT setval(pg_get_serial_sequence('vis_cliente_grade_slot','id'), COALESCE((SELECT MAX(id) FROM vis_cliente_grade_slot), 1));
SELECT setval(pg_get_serial_sequence('vis_cliente_grade_escopo','id'), COALESCE((SELECT MAX(id) FROM vis_cliente_grade_escopo), 1));
SELECT setval(pg_get_serial_sequence('ops_arme_janela','id'), COALESCE((SELECT MAX(id) FROM ops_arme_janela), 1));
COMMIT;
