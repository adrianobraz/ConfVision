-- Fase C4 / A3 — repartir câmeras entre Rust e Python
-- REVISAR ids e worker_id reais antes de COMMIT

-- Diagnóstico
SELECT id, worker_id, ativo, analitico_pausado, stream_motivo_pausa
FROM vis_camera
WHERE ativo IS TRUE
  AND (worker_id LIKE 'rust-processor%' OR id IN (2, 4, 8, 9))
ORDER BY id;

-- Câmeras 404 conhecidas → worker Python (ajuste worker-docker-21)
-- BEGIN;
-- UPDATE vis_camera SET worker_id = 'worker-docker-21'
-- WHERE id IN (2, 4, 8, 9);
-- COMMIT;

-- Exemplo split 10 → 6 no pilot-01 + 4 no pilot-02 (ids fictícios — ajuste)
-- BEGIN;
-- UPDATE vis_camera SET worker_id = 'rust-processor-pilot-01' WHERE id IN (3,5,6,18,19,20);
-- UPDATE vis_camera SET worker_id = 'rust-processor-pilot-02' WHERE id IN (21,22,26,27);
-- COMMIT;

-- Rollback piloto → Python
-- UPDATE vis_camera SET worker_id = 'worker-docker-21' WHERE worker_id = 'rust-processor-pilot-01';
