//! Control Plane vs Data Plane (Fase 4–6) — mapeamento no código atual.
//!
//! **Control Plane (decisão):** API Go + PostgreSQL (`vis_camera.worker_id`, `vis_worker`, D5 assign).
//! **Data Plane (execução):** MediaMTX + `confvision-rust-processor` (`CameraManager` sync pull).
//!
//! Não existe binário `node_agent` separado: o processor Rust faz sync, ping e reconciliação local.
//!
//! ## Unidade de escala (Fase 6)
//!
//! Expansão horizontal = **nova instância Rust** com `WORKER_ID` único + registro em
//! `RUST_PROCESSOR_BASE_URLS` (Go D5). Não usar hostname como chave lógica.
//!
//! ## Node Pool (conceito, sem tabela dedicada)
//!
//! - **Pool analítico Rust:** entradas `servidor_id|base_url` em `RUST_PROCESSOR_BASE_URLS`.
//! - **Pool MediaMTX / ingest:** `vis_mediamtx_node` + `vis_camera.vis_mediamtx_node_id`.
//! - **Pool DVR / motion / timelapse:** workers Python distintos (`worker_tipo` em `vis_worker`).
//!
//! ## Scheduler v1 (já implementado)
//!
//! `visdata/rust_processor_d5.go`: GET `/capacity-report` → `scoreProcessorForAssign` →
//! `PickBestProcessor` → UPDATE `worker_id`. Ver `docs/FASE-6-SCALE-AND-OPERATIONS.md`.

/// Identidade estável do nó de vídeo no modelo atual (env + DB).
#[derive(Debug, Clone)]
pub struct VideoNodeIdentity {
    pub worker_id: String,
    pub processor_id: String,
    pub worker_tipo: String,
    pub hostname: String,
}

impl VideoNodeIdentity {
    pub fn from_config(cfg: &crate::config::Config) -> Self {
        Self {
            worker_id: cfg.worker_id.clone(),
            processor_id: cfg.processor_id.clone(),
            worker_tipo: cfg.worker_tipo.clone(),
            hostname: cfg.processor_hostname.clone(),
        }
    }
}
