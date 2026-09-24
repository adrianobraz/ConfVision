use std::time::Duration;

use crate::error::{AppError, AppResult};

#[derive(Clone, Debug)]
pub struct Config {
    pub confvision_api_url: String,
    pub vis_worker_api_key: String,
    pub mediamtx_rtsp_base: String,
    pub rtmp_publish_secret: Option<String>,
    pub processor_id: String,
    pub processor_hostname: String,
    pub processor_version: String,
    pub worker_tipo: String,
    pub sync_filter_worker_id: bool,
    pub mediamtx_node_id: u32,
    pub redis_url: Option<String>,
    pub s3_endpoint: Option<String>,
    pub s3_bucket: Option<String>,
    pub http_host: String,
    pub http_port: u16,
    pub log_level: String,
    pub max_cameras: usize,
    pub sync_interval: Duration,
    pub ping_interval: Duration,
    pub rtsp_connect_timeout: Duration,
    pub rtsp_reconnect_base: Duration,
    pub rtsp_frame_timeout: Duration,
    pub frame_buffer_max: usize,
    pub queue_backend: String,
}

impl Config {
    pub fn from_env() -> AppResult<Self> {
        let confvision_api_url = require_env("CONFVISION_API_URL")?;
        if confvision_api_url.contains("xano.io") {
            return Err(AppError::Config(
                "CONFVISION_API_URL não pode apontar para Xano (*.xano.io)".into(),
            ));
        }
        forbid_legacy_xano_env()?;

        let processor_id = env_or("PROCESSOR_ID", "rust-processor-01");
        let processor_hostname = env_or(
            "PROCESSOR_HOSTNAME",
            hostname::get()
                .ok()
                .and_then(|h| h.into_string().ok())
                .unwrap_or_else(|| processor_id.clone())
                .as_str(),
        );

        Ok(Self {
            confvision_api_url: confvision_api_url.trim_end_matches('/').to_string(),
            vis_worker_api_key: std::env::var("VIS_WORKER_API_KEY").unwrap_or_default(),
            mediamtx_rtsp_base: env_or("MEDIAMTX_RTSP_BASE", "rtsp://127.0.0.1:8554")
                .trim_end_matches('/')
                .to_string(),
            rtmp_publish_secret: non_empty_opt("RTMP_PUBLISH_SECRET"),
            processor_id,
            processor_hostname,
            processor_version: env_or("PROCESSOR_VERSION", "0.1.0"),
            worker_tipo: env_or("WORKER_TIPO", "rust_processor"),
            sync_filter_worker_id: env_bool("SYNC_FILTER_WORKER_ID", true),
            mediamtx_node_id: env_u32("MEDIAMTX_NODE_ID", 0),
            redis_url: non_empty_opt("REDIS_URL"),
            s3_endpoint: non_empty_opt("S3_ENDPOINT"),
            s3_bucket: non_empty_opt("S3_BUCKET"),
            http_host: env_or("HTTP_HOST", "0.0.0.0"),
            http_port: env_u16("HTTP_PORT", 8090),
            log_level: env_or("LOG_LEVEL", "info"),
            max_cameras: env_usize("MAX_CAMERAS", 10),
            sync_interval: Duration::from_secs(env_u64("SYNC_INTERVAL_SEC", 60)),
            ping_interval: Duration::from_secs(env_u64("PING_INTERVAL_SEC", 30)),
            rtsp_connect_timeout: Duration::from_secs(env_u64("RTSP_CONNECT_TIMEOUT_SEC", 15)),
            rtsp_reconnect_base: Duration::from_secs(env_u64("RTSP_RECONNECT_SECONDS", 10)),
            rtsp_frame_timeout: Duration::from_secs(env_u64("RTSP_FRAME_TIMEOUT_SEC", 30)),
            frame_buffer_max: env_usize("FRAME_BUFFER_MAX", 2).max(1),
            queue_backend: env_or("QUEUE_BACKEND", "none"),
        })
    }
}

fn forbid_legacy_xano_env() -> AppResult<()> {
    for key in ["XANO_BASE_URL", "XANO_API_FRANQUEADO_PRO"] {
        if non_empty_opt(key).is_some() {
            return Err(AppError::Config(format!(
                "variável legada {key} não deve ser usada no Rust processor"
            )));
        }
    }
    Ok(())
}

fn require_env(key: &str) -> AppResult<String> {
    std::env::var(key).map_err(|_| AppError::Config(format!("{key} obrigatório")))
}

fn env_or(key: &str, default: &str) -> String {
    std::env::var(key).unwrap_or_else(|_| default.to_string())
}

fn non_empty_opt(key: &str) -> Option<String> {
    std::env::var(key)
        .ok()
        .map(|s| s.trim().to_string())
        .filter(|s| !s.is_empty())
}

fn env_bool(key: &str, default: bool) -> bool {
    match std::env::var(key) {
        Ok(v) => matches!(
            v.trim().to_lowercase().as_str(),
            "1" | "true" | "yes" | "on"
        ),
        Err(_) => default,
    }
}

fn env_u64(key: &str, default: u64) -> u64 {
    std::env::var(key)
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(default)
}

fn env_u32(key: &str, default: u32) -> u32 {
    std::env::var(key)
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(default)
}

fn env_u16(key: &str, default: u16) -> u16 {
    std::env::var(key)
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(default)
}

fn env_usize(key: &str, default: usize) -> usize {
    std::env::var(key)
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(default)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn rejects_xano_api_url() {
        std::env::set_var("CONFVISION_API_URL", "https://foo.xano.io/api:abc");
        std::env::remove_var("XANO_BASE_URL");
        let err = Config::from_env().unwrap_err();
        assert!(err.to_string().contains("Xano"));
        std::env::remove_var("CONFVISION_API_URL");
    }

    #[test]
    fn rejects_xano_base_env() {
        std::env::set_var("CONFVISION_API_URL", "https://vision.example.com");
        std::env::set_var("XANO_BASE_URL", "https://legacy.xano.io/x");
        let err = Config::from_env().unwrap_err();
        assert!(err.to_string().contains("XANO_BASE_URL"));
        std::env::remove_var("CONFVISION_API_URL");
        std::env::remove_var("XANO_BASE_URL");
    }
}

// Small helper without extra crate dependency for hostname default
mod hostname {
    pub fn get() -> Result<std::ffi::OsString, ()> {
        Ok(std::env::var("COMPUTERNAME")
            .or_else(|_| std::env::var("HOSTNAME"))
            .map(std::ffi::OsString::from)
            .unwrap_or_else(|_| std::ffi::OsString::from("localhost")))
    }
}
