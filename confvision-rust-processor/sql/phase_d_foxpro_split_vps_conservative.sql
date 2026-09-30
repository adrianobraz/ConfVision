-- D1 — split conservador foxpro (2× Rust, VPS CPU)
-- Sem Xano. API Go + Postgres.
-- Referência foxpro 2026-09-27: sync analítico ~ [2,5,18,19,22]; id 2 = RTSP 404 (cam/jdw6yld9mebn).

-- =============================================================================
-- 0) Inventário (rodar sempre antes)
-- =============================================================================
SELECT id, nome, ativo, deteccao_humano, analitico_pausado, worker_id
FROM vis_camera
WHERE ativo IS TRUE
  AND deteccao_humano IS TRUE
ORDER BY id;

SELECT worker_id, COUNT(*) AS n
FROM vis_camera
WHERE ativo IS TRUE
  AND deteccao_humano IS TRUE
  AND COALESCE(analitico_pausado, FALSE) IS FALSE
GROUP BY 1
ORDER BY 1;

-- =============================================================================
-- 1) Câmera 2 — fora do analítico até RTSP ok (404 no MediaMTX)
--    Entrada Principal, path cam/jdw6yld9mebn
-- =============================================================================
-- BEGIN;
-- UPDATE vis_camera
-- SET analitico_pausado = TRUE,
--     stream_motivo_pausa = COALESCE(stream_motivo_pausa, 'manual_d1_404_entrada')
-- WHERE id = 2;
-- COMMIT;

-- =============================================================================
-- 2) Normalizar legado (worker Python / env antigo → Rust pilot-01)
-- =============================================================================
-- BEGIN;
-- UPDATE vis_camera
-- SET worker_id = 'rust-processor-pilot-01'
-- WHERE worker_id IN ('worker-processor-pilot-01', 'worker-docker-21');
-- COMMIT;

-- =============================================================================
-- 3) Split D1 recomendado (5 cams → 3 + 1, SEM id 2)
--    pilot-01: 5, 18, 19  (MAX_CAMERAS=3 no env)
--    pilot-02: 22         (+ 21 se existir no SELECT e RTSP ok)
-- =============================================================================
-- BEGIN;
--
-- UPDATE vis_camera
-- SET worker_id = 'rust-processor-pilot-01',
--     analitico_pausado = FALSE
-- WHERE id IN (5, 18, 19);
--
-- UPDATE vis_camera
-- SET worker_id = 'rust-processor-pilot-02',
--     analitico_pausado = FALSE
-- WHERE id IN (22);
-- -- Opcional se id 21 ativo com stream ok:
-- -- UPDATE vis_camera SET worker_id = 'rust-processor-pilot-02', analitico_pausado = FALSE WHERE id = 21;
--
-- COMMIT;

-- =============================================================================
-- 4) Confirmar pós-split
-- =============================================================================
-- SELECT id, nome, worker_id, analitico_pausado
-- FROM vis_camera
-- WHERE id IN (2, 5, 18, 19, 21, 22)
--    OR worker_id IN ('rust-processor-pilot-01', 'rust-processor-pilot-02')
-- ORDER BY id;
--
-- SELECT worker_id, COUNT(*) AS n
-- FROM vis_camera
-- WHERE worker_id IN ('rust-processor-pilot-01', 'rust-processor-pilot-02')
--   AND ativo IS TRUE
--   AND deteccao_humano IS TRUE
--   AND COALESCE(analitico_pausado, FALSE) IS FALSE
-- GROUP BY 1;

-- Rollback D1: UPDATE vis_camera SET worker_id = 'rust-processor-pilot-01'
--               WHERE worker_id = 'rust-processor-pilot-02';
