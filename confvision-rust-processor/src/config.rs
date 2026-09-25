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
    /// Fase 6.2 — fallback runtime NVDEC→CPU (somente VIDEO_ACCELERATION=auto).
    pub decode_runtime_fallback: bool,
    pub decode_hw_error_threshold: u32,
    pub load_policy_mode: crate::load::LoadPolicyMode,
    pub load_admission_enabled: bool,
    /// Fase 6.3 — teto de decode+motion por câmera (0 = sem limite).
    pub motion_analysis_max_fps: f64,
    /// 1 AU a cada N (Python `MOTION_FRAME_SKIP`, default 1 = todos elegíveis).
    pub motion_frame_stride: usize,
    /// Decode+motion 1 a cada N AUs elegíveis (Python `FRAME_SKIP`, default 1).
    pub decode_frame_stride: usize,
    /// Só decode+motion frequente após movimento (Python `YOLO_ONLY_ON_MOTION`).
    pub analysis_only_on_motion: bool,
    /// FPS máximo de probe decode+motion com cena parada (`analysis_only_on_motion`).
    pub motion_gate_probe_max_fps: f64,
    /// Análises consecutivas sem movimento antes de desarmar (Python `YOLO_MOTION_MISS_FRAMES` × amostras).
    pub motion_gate_miss_frames: u32,
    /// Probe idle: enfileira decode só em IDR (menos CPU entre movimentos).
    pub motion_probe_keyframe_only: bool,
    /// Desconecta RTSP entre probes quando `analysis_only_on_motion`.
    pub rtsp_idle_suspend: bool,
    pub motion_pixel_diff_threshold: u8,
    pub motion_percent_threshold: u32,
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

/// Defaults Fase 6.2 — uso em stubs de teste.
pub fn fill_phase62_defaults(cfg: &mut Config) {
    cfg.decode_runtime_fallback = true;
    cfg.decode_hw_error_threshold = 10;
    cfg.load_policy_mode = crate::load::LoadPolicyMode::Advisory;
    cfg.load_admission_enabled = false;
    cfg.motion_analysis_max_fps = 0.0;
    cfg.motion_frame_stride = 1;
    cfg.decode_frame_stride = 1;
    cfg.analysis_only_on_motion = false;
    cfg.motion_gate_probe_max_fps = 0.5;
    cfg.motion_gate_miss_frames = 10;
    cfg.motion_probe_keyframe_only = false;
    cfg.rtsp_idle_suspend = false;
    cfg.motion_pixel_diff_threshold = 8;
    cfg.motion_percent_threshold = 5;
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
            rtsp_reconnect_base: Duration::from_secs(env_u64("RTSP_RECONNECT_SECONDS", 10).max(1)),
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
            decode_runtime_fallback: env_bool("DECODE_RUNTIME_FALLBACK", true),
            decode_hw_error_threshold: env_u32("DECODE_HW_ERROR_THRESHOLD", 10).max(1),
            load_policy_mode: parse_load_policy_mode(&env_or("LOAD_POLICY_MODE", "advisory"))?,
            load_admission_enabled: env_bool("LOAD_ADMISSION_ENABLED", false),
            motion_analysis_max_fps: analysis_max_fps_from_env(),
            motion_frame_stride: analysis_stride_from_env("MOTION_FRAME_STRIDE", 3, 1),
            decode_frame_stride: analysis_stride_from_env("DECODE_FRAME_STRIDE", 5, 1),
            analysis_only_on_motion: analysis_only_on_motion_from_env(),
            motion_gate_probe_max_fps: motion_gate_probe_max_fps_from_env(),
            motion_gate_miss_frames: motion_gate_miss_frames_from_env(),
            motion_probe_keyframe_only: motion_probe_keyframe_only_from_env(),
            rtsp_idle_suspend: rtsp_idle_suspend_from_env(),
            motion_pixel_diff_threshold: motion_pixel_diff_from_env(),
            motion_percent_threshold: motion_percent_from_env(),
        })
    }

    pub fn load_policy_config(&self) -> crate::load::LoadPolicyConfig {
        crate::load::LoadPolicyConfig {
            mode: self.load_policy_mode,
            admission_enabled: self.load_admission_enabled,
        }
    }

    pub fn decode_fallback_config(&self) -> crate::decode::DecodeFallbackConfig {
        crate::decode::DecodeFallbackConfig::from_env_parts(
            self.decode_runtime_fallback,
            self.decode_hw_error_threshold,
        )
    }
}

/// Perfil VPS alinhado ao worker Python (`FRAME_SKIP`, `MOTION_FRAME_SKIP`, FPS menor).
fn analysis_legacy_vps_enabled() -> bool {
    env_bool("ANALYSIS_LEGACY_VPS", false)
}

fn analysis_max_fps_from_env() -> f64 {
    let default = if analysis_legacy_vps_enabled() {
        2.0
    } else {
        5.0
    };
    env_f64("MOTION_ANALYSIS_MAX_FPS", default)
}

fn analysis_stride_from_env(key: &str, legacy_default: usize, normal_default: usize) -> usize {
    let default = if analysis_legacy_vps_enabled() {
        legacy_default
    } else {
        normal_default
    };
    env_usize(key, default).max(1)
}

fn analysis_only_on_motion_from_env() -> bool {
    if std::env::var("ANALYSIS_ONLY_ON_MOTION").is_ok() {
        return env_bool("ANALYSIS_ONLY_ON_MOTION", true);
    }
    analysis_legacy_vps_enabled()
}

fn motion_gate_probe_max_fps_from_env() -> f64 {
    let default = if analysis_legacy_vps_enabled() {
        0.5
    } else {
        1.0
    };
    env_f64("MOTION_GATE_PROBE_MAX_FPS", default).max(0.01)
}

fn motion_gate_miss_frames_from_env() -> u32 {
    let legacy_default = env_u32("YOLO_MOTION_MISS_FRAMES", 2).max(1).saturating_mul(5);
    let default = if analysis_legacy_vps_enabled() {
        legacy_default
    } else {
        10
    };
    env_u32("MOTION_GATE_MISS_FRAMES", default).max(1)
}

fn motion_probe_keyframe_only_from_env() -> bool {
    if std::env::var("MOTION_PROBE_KEYFRAME_ONLY").is_ok() {
        return env_bool("MOTION_PROBE_KEYFRAME_ONLY", true);
    }
    analysis_legacy_vps_enabled() || analysis_only_on_motion_from_env()
}

fn rtsp_idle_suspend_from_env() -> bool {
    if std::env::var("RTSP_IDLE_SUSPEND").is_ok() {
        return env_bool("RTSP_IDLE_SUSPEND", true);
    }
    analysis_legacy_vps_enabled() || analysis_only_on_motion_from_env()
}

fn motion_pixel_diff_from_env() -> u8 {
    env_u32("MOTION_PIXEL_DIFF_THRESHOLD", 8).min(255) as u8
}

fn motion_percent_from_env() -> u32 {
    env_u32("MOTION_PERCENT_THRESHOLD", 5).min(100)
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

fn env_bool(key: &str, default: bool) -> bool {
    match std::env::var(key) {
        Ok(v) => matches!(
            v.trim().to_lowercase().as_str(),
            "1" | "true" | "yes" | "on"
        ),
        Err(_) => default,
    }
}

fn parse_load_policy_mode(raw: &str) -> AppResult<crate::load::LoadPolicyMode> {
    crate::load::LoadPolicyMode::parse(raw).map_err(AppError::Config)
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
