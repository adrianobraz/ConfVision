-- Exemplo: liberar sync Go (analitico_pausado) — REVISAR ids com negócio antes de executar.
-- Elegíveis no export 2026-09-27 já com pausa=false: 5,18,19,22,26,27
-- Câmeras ativas frequentemente pausadas no export: 2,3,4,6,8,9,15,20,21,25,28,29

-- Ver estado atual:
-- SELECT id, nome, analitico_pausado, deteccao_humano, worker_id FROM vis_camera WHERE ativo ORDER BY id;

-- UPDATE vis_camera SET analitico_pausado = FALSE WHERE id IN (/* ids */);
