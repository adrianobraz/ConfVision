use crate::api::CameraRecord;
use crate::config::Config;
use crate::error::{AppError, AppResult};

/// Monta URL RTSP alinhada ao Python `urls.rtsp_url_for_camera` (sem credenciais no log).
pub fn resolve_rtsp_url(cam: &CameraRecord, cfg: &Config) -> AppResult<String> {
    if let Some(sec) = cam
        .rtsp_url_sec
        .as_ref()
        .map(|s| s.trim())
        .filter(|s| !s.is_empty())
    {
        let lower = sec.to_lowercase();
        if lower.starts_with("rtsp://") || lower.starts_with("rtsps://") {
            if lower.contains("/live/") {
                // legado — ignorar e usar cam/{hash}
            } else {
                return Ok(sec.to_string());
            }
        }
    }

    let base = cam
        .mediamtx_rtsp_base
        .as_ref()
        .map(|s| s.trim())
        .filter(|s| !s.is_empty())
        .map(|s| s.trim_end_matches('/').to_string())
        .unwrap_or_else(|| cfg.mediamtx_rtsp_base.clone());

    let path = stream_path(cam.id, cfg)?;
    Ok(format!("{base}/{path}"))
}

pub fn stream_path(camera_id: i64, cfg: &Config) -> AppResult<String> {
    let secret = cfg.rtmp_publish_secret.as_ref().ok_or_else(|| {
        AppError::Config("RTMP_PUBLISH_SECRET necessário para path cam/{hash}".into())
    })?;
    let h = hashids::HashIds::builder()
        .with_salt(secret.as_str())
        .with_min_length(12)
        .with_alphabet("0123456789abcdefghijklmnopqrstuvwxyz")
        .finish()
        .map_err(|_| AppError::Config("alfabeto hashids inválido".into()))?;
    let hash = h.encode(&[camera_id as u64]);
    if hash.is_empty() {
        return Err(AppError::Config("falha ao gerar hash da camera".into()));
    }
    Ok(format!("cam/{hash}"))
}

pub fn redact_rtsp_url(url: &str) -> String {
    if let Ok(parsed) = url::Url::parse(url) {
        let host = parsed.host_str().unwrap_or("?");
        let port = parsed.port().map(|p| format!(":{p}")).unwrap_or_default();
        let path = parsed.path();
        return format!("rtsp://{host}{port}{path}");
    }
    "rtsp://?".to_string()
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::api::CameraRecord;
    use serde_json::json;
    use std::time::Duration;

    fn cfg_with_secret() -> Config {
        Config {
            confvision_api_url: "http://localhost".into(),
            vis_worker_api_key: String::new(),
            mediamtx_rtsp_base: "rtsp://mtx:8554".into(),
            rtmp_publish_secret: Some("test-secret-salt".into()),
            processor_id: "p".into(),
            processor_hostname: "h".into(),
            processor_version: "0.1.0".into(),
            worker_tipo: "rust".into(),
            worker_id: "worker-01".into(),
            shard_mode: crate::config::ShardMode::Auto,
            worker_shard_index: -1,
            worker_shard_total: 0,
            mediamtx_node_id: 0,
            redis_url: None,
            s3_endpoint: None,
            s3_bucket: None,
            http_host: "0.0.0.0".into(),
            http_port: 8090,
            log_level: "info".into(),
            max_cameras: 1,
            sync_interval: Duration::from_secs(60),
            ping_interval: Duration::from_secs(30),
            rtsp_connect_timeout: Duration::from_secs(5),
            rtsp_reconnect_base: Duration::from_secs(10),
            rtsp_frame_timeout: Duration::from_secs(30),
            frame_buffer_max: 2,
            queue_backend: "none".into(),
            capacity_mode: crate::config::CapacityMode::Dynamic,
            capacity_cpu_target_percent: 80.0,
            capacity_memory_target_percent: 80.0,
            capacity_gpu_target_percent: 80.0,
            capacity_vram_target_percent: 80.0,
            capacity_min_sample_sec: 30,
            capacity_safety_factor: 0.80,
            capacity_history_size: 120,
            capacity_sample_interval_sec: 5,
            decode_runtime_fallback: true,
            decode_hw_error_threshold: 10,
            load_policy_mode: crate::load::LoadPolicyMode::Advisory,
            load_admission_enabled: false,
        }
    }

    #[test]
    fn builds_mediamtx_url() {
        let cfg = cfg_with_secret();
        let cam = CameraRecord {
            id: 100,
            ativo: true,
            nome: None,
            rtsp_url_sec: None,
            mediamtx_rtsp_base: None,
            analitico_pausado: None,
            deteccao_humano: Some(true),
            worker_id: None,
            extra: json!({}),
        };
        let url = resolve_rtsp_url(&cam, &cfg).unwrap();
        assert!(url.starts_with("rtsp://mtx:8554/cam/"));
    }

    #[test]
    fn redact_hides_userinfo() {
        let r = redact_rtsp_url("rtsp://user:pass@host:8554/cam/abc");
        assert!(!r.contains("pass"));
        assert!(r.contains("host"));
    }
}
