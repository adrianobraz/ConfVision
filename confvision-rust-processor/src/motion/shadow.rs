//! Shadow mode Fase 2A — JSONL para comparar Luma (Rust) vs MOG2 (Python).

use std::fs::{create_dir_all, OpenOptions};
use std::io::Write;
use std::path::PathBuf;
use std::sync::Mutex;

use serde::Serialize;

use super::session::{MotionSessionEvent, MotionSessionEventKind};

#[derive(Debug, Clone)]
pub struct MotionShadowConfig {
    pub enabled: bool,
    pub log_dir: PathBuf,
    pub processor_id: String,
}

#[derive(Debug, Serialize)]
struct ShadowSampleLine<'a> {
    source: &'static str,
    algorithm: &'static str,
    processor_id: &'a str,
    camera_id: i64,
    ts_unix_ms: i64,
    detected: bool,
    score_percent: u32,
    #[serde(skip_serializing_if = "Option::is_none")]
    session_event: Option<&'static str>,
}

pub struct MotionShadowWriter {
    path: PathBuf,
    processor_id: String,
    camera_id: i64,
    file: Mutex<Option<std::fs::File>>,
}

impl MotionShadowWriter {
    pub fn new(cfg: &MotionShadowConfig, camera_id: i64) -> Option<Self> {
        if !cfg.enabled {
            return None;
        }
        let _ = create_dir_all(&cfg.log_dir);
        let path = cfg
            .log_dir
            .join(format!("rust_cam_{camera_id}.jsonl"));
        Some(Self {
            path,
            processor_id: cfg.processor_id.clone(),
            camera_id,
            file: Mutex::new(None),
        })
    }

    pub fn write_sample(&self, detected: bool, score_percent: u32, ts_unix_ms: i64) {
        let line = ShadowSampleLine {
            source: "rust",
            algorithm: "luma",
            processor_id: &self.processor_id,
            camera_id: self.camera_id,
            ts_unix_ms,
            detected,
            score_percent,
            session_event: None,
        };
        self.append(&line);
    }

    pub fn write_session_event(&self, ev: &MotionSessionEvent) {
        let session_event = Some(match ev.kind {
            MotionSessionEventKind::MotionStarted => "motion_started",
            MotionSessionEventKind::MotionUpdated => "motion_updated",
            MotionSessionEventKind::MotionEnded => "motion_ended",
        });
        let line = ShadowSampleLine {
            source: "rust",
            algorithm: "luma",
            processor_id: &self.processor_id,
            camera_id: ev.camera_id,
            ts_unix_ms: ev.ts_unix_ms,
            detected: !matches!(ev.kind, MotionSessionEventKind::MotionEnded),
            score_percent: ev.score_percent,
            session_event,
        };
        self.append(&line);
    }

    fn append<T: Serialize>(&self, value: &T) {
        let Ok(json) = serde_json::to_string(value) else {
            return;
        };
        let Ok(mut guard) = self.file.lock() else {
            return;
        };
        if guard.is_none() {
            *guard = OpenOptions::new()
                .create(true)
                .append(true)
                .open(&self.path)
                .ok();
        }
        let Some(f) = guard.as_mut() else {
            return;
        };
        let _ = writeln!(f, "{json}");
        let _ = f.flush();
    }
}

impl MotionShadowConfig {
    pub fn from_config(cfg: &crate::config::Config) -> Self {
        Self {
            enabled: cfg.motion_shadow_compare,
            log_dir: cfg.motion_shadow_log_dir.clone(),
            processor_id: cfg.processor_id.clone(),
        }
    }
}

pub fn default_log_dir() -> PathBuf {
    PathBuf::from("/tmp/confvision/motion-shadow")
}
