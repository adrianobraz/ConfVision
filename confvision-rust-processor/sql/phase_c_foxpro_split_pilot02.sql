-- C4 — dividir carga foxpro: pilot-01 + pilot-02
-- Ajuste ids conforme SELECT abaixo. PYTHON 404: cameras_404_worker_python.sql primeiro.

-- Inventário atual
SELECT id, nome, ativo, worker_id
FROM vis_camera
WHERE worker_id IN ('rust-processor-pilot-01', 'rust-processor-pilot-02')
   OR (ativo IS TRUE AND worker_id LIKE 'rust-processor%')
ORDER BY id;

-- Snapshot 2026-09-26 (10 online no pilot-01): 3,5,15,18,19,20,21,22,26,27
-- Proposta: 6 no 01, 4 no 02

-- BEGIN;
--
-- UPDATE vis_camera SET worker_id = 'rust-processor-pilot-01'
-- WHERE id IN (3, 5, 15, 18, 19, 20);
--
-- UPDATE vis_camera SET worker_id = 'rust-processor-pilot-02'
-- WHERE id IN (21, 22, 26, 27);
--
-- COMMIT;

-- Confirmar
-- SELECT worker_id, COUNT(*) FROM vis_camera WHERE worker_id LIKE 'rust-processor-pilot%' GROUP BY 1;
