//! Abstração de storage de mídia (Fase 5).
//!
//! Objetivo: concentrar upload/leitura futura fora de paths fixos nos call sites.
//! Implementação atual: S3 compatível (Contabo) via `upload_file`.

use std::path::Path;

use tracing::warn;

use crate::config::Config;
use crate::error::{AppError, AppResult};

use super::upload::upload_file;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum MediaKind {
    EventSnapshot,
    EventClip,
}

/// Backend configurável a partir de `Config` (endpoint/bucket/credenciais).
#[derive(Clone)]
pub struct MediaStorage {
    cfg: Config,
}

impl MediaStorage {
    pub fn from_config(cfg: &Config) -> Self {
        Self { cfg: cfg.clone() }
    }

    pub fn is_ready(&self) -> bool {
        self.cfg.s3_ready()
    }

    /// Grava objeto remoto; falha se S3 não configurado ou HTTP erro.
    pub async fn put(
        &self,
        local: &Path,
        object_key: &str,
        content_type: &str,
    ) -> AppResult<String> {
        if !self.is_ready() {
            return Err(AppError::Config(
                "media storage (S3) não configurado".into(),
            ));
        }
        upload_file(&self.cfg, local, object_key, content_type).await
    }

    /// Upload tolerante — usado no pipeline de captura para não bloquear `finalizar_evento`.
    pub async fn put_or_log(
        &self,
        kind: MediaKind,
        camera_id: i64,
        local: &Path,
        object_key: &str,
        content_type: &str,
    ) -> Option<String> {
        if !local.exists() {
            return None;
        }
        if !self.is_ready() {
            warn!(camera_id, ?kind, "media storage indisponível — upload omitido");
            return None;
        }
        match self.put(local, object_key, content_type).await {
            Ok(url) => Some(url),
            Err(e) => {
                warn!(camera_id, ?kind, error = %e, "media upload falhou");
                None
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn storage_not_ready_without_s3_env() {
        let cfg = Config::test_stub();
        let st = MediaStorage::from_config(&cfg);
        assert!(!st.is_ready());
    }
}
