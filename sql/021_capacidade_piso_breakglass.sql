-- Piso Break-glass para preco base de capacidade ConfVision
ALTER TABLE vis_capacidade_config
    ADD COLUMN IF NOT EXISTS preco_piso_breakglass NUMERIC(12, 2) NOT NULL DEFAULT 0;

UPDATE vis_capacidade_config
SET preco_piso_breakglass = preco_base_camera
WHERE preco_piso_breakglass = 0 AND preco_base_camera > 0;
