use std::path::Path;

use tracing::{info, warn};

use crate::config::Config;
use crate::error::{AppError, AppResult};

/// Upload S3 compatível Contabo (path-style). Retorna URL pública ou key.
pub async fn upload_file(
    cfg: &Config,
    local: &Path,
    object_key: &str,
    _content_type: &str,
) -> AppResult<String> {
    let endpoint = cfg
        .s3_endpoint
        .as_ref()
        .ok_or_else(|| AppError::Config("S3_ENDPOINT não configurado".into()))?;
    let bucket = cfg
        .s3_bucket
        .as_ref()
        .ok_or_else(|| AppError::Config("S3_BUCKET não configurado".into()))?;
    let access = cfg
        .s3_access_key
        .as_ref()
        .ok_or_else(|| AppError::Config("S3_ACCESS_KEY não configurado".into()))?;
    let secret = cfg
        .s3_secret_key
        .as_ref()
        .ok_or_else(|| AppError::Config("S3_SECRET_KEY não configurado".into()))?;

    let bytes = tokio::fs::read(local)
        .await
        .map_err(|e| AppError::Other(e.into()))?;
    let url = format!(
        "{}/{}/{}",
        endpoint.trim_end_matches('/'),
        bucket.trim_matches('/'),
        object_key.trim_start_matches('/')
    );

    let client = reqwest::Client::new();
    let resp = client
        .put(&url)
        .basic_auth(access, Some(secret))
        .header("Content-Length", bytes.len())
        .body(bytes)
        .send()
        .await
        .map_err(|e| AppError::Other(e.into()))?;

    if !resp.status().is_success() {
        let status = resp.status();
        let body = resp.text().await.unwrap_or_default();
        warn!(status = %status, body = %body, "s3 upload failed");
        return Err(AppError::Config(format!("s3 upload HTTP {status}")));
    }
    info!(object_key, "s3 upload ok");
    Ok(url)
}

pub fn evento_snapshot_key(id_franqueado: Option<&str>, evento_id: i64) -> String {
    let fra = id_franqueado.unwrap_or("0");
    format!("eventos/{fra}/{evento_id}/snapshot.jpg")
}

pub fn evento_clip_key(id_franqueado: Option<&str>, evento_id: i64, seq: u32) -> String {
    let fra = id_franqueado.unwrap_or("0");
    format!("eventos/{fra}/{evento_id}/clip_{seq:03}.mp4")
}
