use std::path::Path;
use std::process::Stdio;
use std::time::Duration;

use tokio::process::Command;
use tracing::warn;

use crate::error::{AppError, AppResult};

pub async fn capture_clip_rtsp(
    rtsp_url: &str,
    dest: &Path,
    duration_sec: u32,
) -> AppResult<()> {
    if let Some(parent) = dest.parent() {
        std::fs::create_dir_all(parent).map_err(|e| AppError::Other(e.into()))?;
    }
    let dur = duration_sec.max(1);
    let output = Command::new("ffmpeg")
        .args([
            "-hide_banner",
            "-loglevel",
            "error",
            "-rtsp_transport",
            "tcp",
            "-i",
            rtsp_url,
            "-t",
            &dur.to_string(),
            "-c",
            "copy",
            "-y",
            dest.to_str().unwrap_or("clip.mp4"),
        ])
        .stdout(Stdio::null())
        .stderr(Stdio::piped())
        .kill_on_drop(true)
        .output()
        .await
        .map_err(|e| AppError::Config(format!("ffmpeg spawn: {e}")))?;

    if !output.status.success() {
        let err = String::from_utf8_lossy(&output.stderr);
        warn!(stderr = %err, "ffmpeg clip failed");
        return Err(AppError::Config(format!("ffmpeg clip: {err}")));
    }
    if !dest.exists() || dest.metadata().map(|m| m.len()).unwrap_or(0) == 0 {
        return Err(AppError::Config("clip vazio".into()));
    }
    Ok(())
}

pub async fn capture_snapshot_rtsp(rtsp_url: &str, dest: &Path) -> AppResult<()> {
    if let Some(parent) = dest.parent() {
        std::fs::create_dir_all(parent).map_err(|e| AppError::Other(e.into()))?;
    }
    let output = Command::new("ffmpeg")
        .args([
            "-hide_banner",
            "-loglevel",
            "error",
            "-rtsp_transport",
            "tcp",
            "-i",
            rtsp_url,
            "-frames:v",
            "1",
            "-q:v",
            "2",
            "-y",
            dest.to_str().unwrap_or("snap.jpg"),
        ])
        .stdout(Stdio::null())
        .stderr(Stdio::piped())
        .kill_on_drop(true)
        .output()
        .await
        .map_err(|e| AppError::Config(format!("ffmpeg spawn: {e}")))?;

    if !output.status.success() {
        let err = String::from_utf8_lossy(&output.stderr);
        return Err(AppError::Config(format!("ffmpeg snapshot: {err}")));
    }
    Ok(())
}

pub fn clip_timeout(duration_sec: u32) -> Duration {
    Duration::from_secs(duration_sec as u64 + 30)
}
