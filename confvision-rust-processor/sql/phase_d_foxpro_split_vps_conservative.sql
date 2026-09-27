-- D1 — split conservador foxpro (VPS CPU, 2 processors Rust)
-- Sem Xano. Ajuste ids após SELECT. Pausar cams RTSP 404 antes do COMMIT.

SELECT id, nome, ativo, deteccao_humano, worker_id
FROM vis_camera
WHERE ativo IS TRUE
  AND deteccao_humano IS TRUE
ORDER BY id;

-- Exemplo 2026-09-27 (revisar!): pilot-01 tinha sync [2,5,18,19,21/22]
-- Id 2 = 404 frequente — considere EXCLUIR do analitico ou manter só num lado após fix RTSP

-- BEGIN;
--
-- UPDATE vis_camera SET worker_id = 'rust-processor-pilot-01'
-- WHERE id IN (5, 18, 19);
--
-- UPDATE vis_camera SET worker_id = 'rust-processor-pilot-02'
-- WHERE id IN (21, 22);
--
-- COMMIT;

-- SELECT worker_id, COUNT(*) AS n
-- FROM vis_camera
-- WHERE worker_id IN ('rust-processor-pilot-01', 'rust-processor-pilot-02')
-- GROUP BY 1;
