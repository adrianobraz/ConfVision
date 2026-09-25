use std::str::FromStr;
use std::time::Duration;

use serde::Serialize;

use crate::error::{AppError, AppResult};

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum CapacityMode {
    Dynamic,
    Disabled,
}

impl FromStr for CapacityMode {
    type Err = AppError;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.trim().to_lowercase().as_str() {
            "dynamic" => Ok(Self::Dynamic),
            "disabled" => Ok(Self::Disabled),
            other => Err(AppError::Config(format!(
                "CAPACITY_MODE inválido: {other:?} (use dynamic ou disabled)"
            ))),
        }
    }
}

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum ShardMode {
    Auto,
    WorkerId,
    Hash,
}

impl ShardMode {
    pub fn as_str(self) -> &'static str {
        match self {
            Self::Auto => "auto",
            Self::WorkerId => "worker_id",
            Self::Hash => "hash",
        }
    }
}

impl FromStr for ShardMode {
    type Err = AppError;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.trim().to_lowercase().as_str() {
            "auto" => Ok(Self::Auto),
            "worker_id" => Ok(Self::WorkerId),
            "hash" => Ok(Self::Hash),
            other => Err(AppError::Config(format!(
                "SHARD_MODE inválido: {other:?} (use auto, worker_id ou hash)"
            ))),
        }
    }
}

#[derive(Clone, Debug)]
pub struct Config {
    pub confvision_api_url: String,
    pub vis_worker_api_key: String,
    pub mediamtx_rtsp_base: String,
    pub rtmp_publish_secret: Option<String>,
    pub processor_id: String,
    pub processor_hostname: String,
    pub processor_version: String,
    /// Identidade lógica do worker (`WORKER_ID`) — usada em ping e sync worker_id.
    pub worker_id: String,
    pub worker_tipo: String,
    pub shard_mode: ShardMode,
    pub worker_shard_index: i32,
    pub worker_shard_total: u32,
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
    pub capacity_mode: CapacityMode,
    pub capacity_cpu_target_percent: f64,
    pub capacity_memory_target_percent: f64,
    pub capacity_gpu_target_percent: f64,
    pub capacity_vram_target_percent: f64,
    pub capacity_min_sample_sec: u64,
    pub capacity_safety_factor: f64,
    pub capacity_history_size: usize,
    pub capacity_sample_interval_sec: u64,
}

impl Config {
    /// Teto local de câmeras (hard safety). `0` = sem limite.
    pub fn effective_max_cameras(&self) -> usize {
        if self.max_cameras == 0 {
            usize::MAX
        } else {
            self.max_cameras
        }
    }

    pub fn max_cameras_hard_limit(&self) -> Option<usize> {
        if self.max_cameras == 0 {
            None
        } else {
            Some(self.max_cameras)
        }
    }

    pub fn max_cameras_for_ping(&self) -> Option<i32> {
        self.max_cameras_hard_limit()
            .and_then(|v| i32::try_from(v).ok())
    }
}

/// Preenche campos de capacidade (Fase 6) — uso em testes e stubs.
pub fn fill_capacity_defaults(cfg: &mut Config) {
    cfg.capacity_mode = CapacityMode::Dynamic;
    cfg.capacity_cpu_target_percent = 80.0;
    cfg.capacity_memory_target_percent = 80.0;
    cfg.capacity_gpu_target_percent = 80.0;
    cfg.capacity_vram_target_percent = 80.0;
    cfg.capacity_min_sample_sec = 30;
    cfg.capacity_safety_factor = 0.80;
    cfg.capacity_history_size = 120;
    cfg.capacity_sample_interval_sec = 5;
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
            worker_id: env_or("WORKER_ID", "worker-01"),
            worker_tipo: env_or("WORKER_TIPO", "rust_processor"),
            shard_mode: parse_shard_mode(&env_or("SHARD_MODE", "auto"))?,
            worker_shard_index: env_i32("WORKER_SHARD_INDEX", -1),
            worker_shard_total: env_u32("WORKER_SHARD_TOTAL", 0),
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
            capacity_mode: parse_capacity_mode(&env_or("CAPACITY_MODE", "dynamic"))?,
            capacity_cpu_target_percent: env_f64("CAPACITY_CPU_TARGET_PERCENT", 80.0),
            capacity_memory_target_percent: env_f64("CAPACITY_MEMORY_TARGET_PERCENT", 80.0),
            capacity_gpu_target_percent: env_f64("CAPACITY_GPU_TARGET_PERCENT", 80.0),
            capacity_vram_target_percent: env_f64("CAPACITY_VRAM_TARGET_PERCENT", 80.0),
            capacity_min_sample_sec: env_u64("CAPACITY_MIN_SAMPLE_SEC", 30),
            capacity_safety_factor: env_f64("CAPACITY_SAFETY_FACTOR", 0.80),
            capacity_history_size: env_usize("CAPACITY_HISTORY_SIZE", 120),
            capacity_sample_interval_sec: env_u64("CAPACITY_SAMPLE_INTERVAL_SEC", 5),
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

fn env_i32(key: &str, default: i32) -> i32 {
    std::env::var(key)
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(default)
}

fn parse_shard_mode(raw: &str) -> AppResult<ShardMode> {
    raw.parse()
}

fn parse_capacity_mode(raw: &str) -> AppResult<CapacityMode> {
    raw.parse()
}

fn env_f64(key: &str, default: f64) -> f64 {
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

    #[test]
    fn parses_shard_mode() {
        assert_eq!("hash".parse::<ShardMode>().unwrap(), ShardMode::Hash);
        assert!("invalid".parse::<ShardMode>().is_err());
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
