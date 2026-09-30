-- Política automática de retry RTSP (Rust processor) — executar no banco confmonit
ALTER TABLE vis_camera ADD COLUMN IF NOT EXISTS ultimo_stream_ok_em TIMESTAMPTZ;
ALTER TABLE vis_camera ADD COLUMN IF NOT EXISTS stream_falhas_consecutivas INT NOT NULL DEFAULT 0;
ALTER TABLE vis_camera ADD COLUMN IF NOT EXISTS stream_tentativas_horarias INT NOT NULL DEFAULT 0;
ALTER TABLE vis_camera ADD COLUMN IF NOT EXISTS stream_policy_generation BIGINT NOT NULL DEFAULT 0;
ALTER TABLE vis_camera ADD COLUMN IF NOT EXISTS stream_motivo_pausa TEXT;
