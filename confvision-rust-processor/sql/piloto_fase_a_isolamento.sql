-- Piloto Fase A: 1 câmera no Rust (rust-processor-pilot-01), demais no worker Python.
-- NÃO cria tabela — apenas UPDATE em vis_camera.worker_id (tabela já existe).
-- Executar no Postgres ConfVision (banco confmonit).
-- Caminho: confvision-rust-processor/sql/piloto_fase_a_isolamento.sql
-- Guia: confvision-rust-processor/docs/PILOTO_FASE_A_ABRIR_AQUI.md
--
-- ANTES DE UPDATE: substituir placeholders e conferir SELECTs.
--
-- PILOTO_CAMERA_ID     = 5   -- câmera com RTSP ok (alternativa: 3)
-- RUST_WORKER_ID       = 'rust-processor-pilot-01'
-- PYTHON_WORKER_ID     = 'worker-docker-21'  -- IGUAL ao WORKER_ID do confvision-worker no EasyPanel

-- ---------------------------------------------------------------------------
-- 1) Inventário atual
-- ---------------------------------------------------------------------------
SELECT id,
       nome,
       ativo,
       worker_id,
       deteccao_humano,
       analitico_pausado
FROM vis_camera
WHERE worker_id = 'rust-processor-pilot-01'
   OR id IN (3, 4, 5)
ORDER BY id;

-- Câmeras que ainda seriam syncadas para o Rust (deve ficar só 1 após UPDATE):
SELECT id, nome, worker_id
FROM vis_camera
WHERE worker_id = 'rust-processor-pilot-01'
  AND COALESCE(ativo, true) = true;

-- ---------------------------------------------------------------------------
-- 2) Aplicar isolamento (revise IDs e PYTHON_WORKER_ID antes de executar)
-- ---------------------------------------------------------------------------
BEGIN;

-- Devolver todas as câmeras do piloto Rust para o Python, exceto a câmera oficial.
UPDATE vis_camera
SET worker_id = 'worker-docker-21'   -- <<< PYTHON_WORKER_ID do EasyPanel worker
WHERE worker_id = 'rust-processor-pilot-01'
  AND id <> 5;                         -- <<< PILOTO_CAMERA_ID (5 ou 3)

-- Garantir que a câmera piloto está no Rust
UPDATE vis_camera
SET worker_id = 'rust-processor-pilot-01'
WHERE id = 5;                          -- <<< PILOTO_CAMERA_ID

-- Confirmar
SELECT id, nome, worker_id
FROM vis_camera
WHERE id IN (3, 4, 5)
ORDER BY id;

COMMIT;
-- ROLLBACK;  -- use em teste se preferir transação abortada
