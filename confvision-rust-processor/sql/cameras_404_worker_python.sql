-- Câmeras com RTSP 404 no piloto Rust — mover para worker Python ou revisar ativo.
-- Banco: confmonit, tabela: vis_camera (sem DDL).
-- Guia: docs/RUNBOOK_CAMERAS_404_ATIVO.md
--
-- Substituir:
--   PYTHON_WORKER_ID = worker-docker-21   (WORKER_ID do confvision-worker no EasyPanel)
--   RUST_WORKER_ID   = rust-processor-pilot-01

-- ---------------------------------------------------------------------------
-- 1) Inventário: ativas no Rust + último contexto operacional
-- ---------------------------------------------------------------------------
SELECT id,
       nome,
       ativo,
       worker_id,
       deteccao_humano
FROM vis_camera
WHERE worker_id = 'rust-processor-pilot-01'
ORDER BY id;

-- Candidatas conhecidas em teste de carga (404 RTSP) — conferir antes de UPDATE:
SELECT id, nome, ativo, worker_id
FROM vis_camera
WHERE id IN (2, 4, 8, 9)
ORDER BY id;

-- Ativas no Rust que ainda consomem sync (devem ser só as desejadas):
SELECT id, nome, worker_id
FROM vis_camera
WHERE worker_id = 'rust-processor-pilot-01'
  AND COALESCE(ativo, true) = true;

-- ---------------------------------------------------------------------------
-- 2) Mover ids 404 típicos de volta ao Python (revise lista de ids)
-- ---------------------------------------------------------------------------
BEGIN;

UPDATE vis_camera
SET worker_id = 'worker-docker-21'   -- <<< PYTHON_WORKER_ID
WHERE worker_id = 'rust-processor-pilot-01'
  AND id IN (2, 4, 8, 9);            -- <<< ajuste conforme /capacity-report

SELECT id, nome, worker_id
FROM vis_camera
WHERE id IN (2, 4, 8, 9)
ORDER BY id;

COMMIT;
-- ROLLBACK;

-- ---------------------------------------------------------------------------
-- 3) Opcional: inativar câmera sem stream (negócio)
-- ---------------------------------------------------------------------------
-- UPDATE vis_camera SET ativo = false WHERE id = ? AND worker_id = 'rust-processor-pilot-01';
