-- Exemplo Fase D / Opção A — despausar analítico + RTSP direto
-- SEMPRE revisar IDs e URLs antes de COMMIT.
-- Substitua rtsp_url_sec pelas URLs reais acessíveis da VPS foxpro.

BEGIN;

-- Despausar câmeras de teste (ajuste a lista de id)
UPDATE vis_camera
SET analitico_pausado = FALSE
WHERE id IN (5, 18, 19, 22, 26, 27)
  AND ativo = TRUE
  AND deteccao_humano = TRUE;

-- Uma câmera por vez (exemplo id=5)
-- UPDATE vis_camera
-- SET rtsp_url_sec = 'rtsp://usuario:senha@192.168.0.10:554/Streaming/Channels/501'
-- WHERE id = 5;

-- Conferência
SELECT id, nome, analitico_pausado,
       LEFT(COALESCE(rtsp_url_sec, ''), 50) AS rtsp_preview,
       worker_id, stream_erro_classe
FROM vis_camera
WHERE id IN (5, 18, 19, 22, 26, 27)
ORDER BY id;

-- COMMIT;
-- ROLLBACK;
