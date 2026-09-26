-- Diagnóstico de último erro RTSP/decode (ping stream_incident / stream_failure)
ALTER TABLE vis_camera ADD COLUMN IF NOT EXISTS stream_ultimo_erro TEXT;
ALTER TABLE vis_camera ADD COLUMN IF NOT EXISTS stream_erro_classe VARCHAR(64);
ALTER TABLE vis_camera ADD COLUMN IF NOT EXISTS stream_ultimo_erro_em TIMESTAMPTZ;
