-- Fase 0 — ops 5 passos (Postgres ConfVision). Revisar antes de COMMIT.
-- Cam 15: piloto no Rust B, analítico ativo.
-- Cam 4: não corrige RTMP no MediaMTX; só metadados / diagnóstico.

BEGIN;

-- Passo 4: câmera 15 no processor B, despausar analítico
UPDATE vis_camera
SET worker_id = 'rust-processor-pilot-b-02',
    analitico_pausado = FALSE,
    ativo = TRUE,
    deteccao_humano = TRUE
WHERE id = 15;

-- Passo 1 (carga): pausar analítico nas demais do B durante fechamento Fase 0 (opcional — descomente se CPU crítica)
-- UPDATE vis_camera SET analitico_pausado = TRUE
-- WHERE worker_id = 'rust-processor-pilot-b-02' AND id IN (3, 5) AND id <> 15;

-- Passo 3: cam 4 — conferir RTSP (404 = path não publicado no MediaMTX; operador deve republicar RTMP em cam/boezyjmydlk4)
UPDATE vis_camera
SET ativo = TRUE,
    deteccao_humano = TRUE
WHERE id = 4;

COMMIT;

-- Verificação pós-apply
SELECT id, nome, worker_id, analitico_pausado, LEFT(COALESCE(rtsp_url_sec, ''), 55) AS rtsp
FROM vis_camera
WHERE id IN (3, 4, 5, 15)
ORDER BY id;
