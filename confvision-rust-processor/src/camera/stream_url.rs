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
        let mut cfg = Config::test_stub();
        cfg.mediamtx_rtsp_base = "rtsp://mtx:8554".into();
        cfg.rtmp_publish_secret = Some("test-secret-salt".into());
        cfg.max_cameras = 1;
        cfg
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
            created_at: None,
            stream_policy_generation: None,
            ultimo_stream_ok_em: None,
            stream_falhas_consecutivas: None,
            stream_tentativas_horarias: None,
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
