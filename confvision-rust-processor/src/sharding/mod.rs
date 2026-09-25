//! Semântica de sharding alinhada a `ConfVision/sharding.py` (workers Python).

use crate::api::CameraRecord;
use crate::config::{Config, ShardMode};

/// Equivalente a `shard_enabled()` no Python.
pub fn shard_enabled(cfg: &Config) -> bool {
    match cfg.shard_mode {
        ShardMode::WorkerId => !cfg.worker_id.is_empty(),
        ShardMode::Hash => cfg.worker_shard_total > 0 && cfg.worker_shard_index >= 0,
        ShardMode::Auto => {
            !cfg.worker_id.is_empty() || (cfg.worker_shard_total > 0 && cfg.worker_shard_index >= 0)
        }
    }
}

/// Equivalente a `camera_belongs_to_shard(camera_id)`.
pub fn camera_belongs_to_shard(cfg: &Config, camera_id: i64) -> bool {
    if !shard_enabled(cfg) {
        return true;
    }
    if cfg.shard_mode == ShardMode::WorkerId {
        // API já filtra por worker_id; não aplicar hash local.
        return true;
    }
    if cfg.worker_shard_total <= 0 || cfg.worker_shard_index < 0 {
        return true;
    }
    (camera_id.rem_euclid(cfg.worker_shard_total as i64)) == cfg.worker_shard_index as i64
}

/// `worker_id` para query HTTP — somente `SHARD_MODE=worker_id` (como `query_params()`).
pub fn sync_query_worker_id(cfg: &Config) -> Option<&str> {
    if cfg.shard_mode == ShardMode::WorkerId && !cfg.worker_id.is_empty() {
        Some(cfg.worker_id.as_str())
    } else {
        None
    }
}

/// Filtro analítico equivalente a `filter_cameras()` no Python.
pub fn filter_analytic_cameras(cfg: &Config, cameras: Vec<CameraRecord>) -> Vec<CameraRecord> {
    let mut filtered: Vec<CameraRecord> = cameras
        .into_iter()
        .filter(|cam| {
            cam.ativo
                && cam.deteccao_humano.unwrap_or(false)
                && camera_belongs_to_shard(cfg, cam.id)
        })
        .collect();

    let hard_max = cfg.effective_max_cameras();
    if filtered.len() > hard_max {
        tracing::warn!(
            total = filtered.len(),
            max = hard_max,
            "shard: truncating camera list (MAX_CAMERAS hard safety)"
        );
        filtered.truncate(hard_max);
    }
    filtered
}

pub fn shard_label(cfg: &Config) -> String {
    let mut parts = vec![
        format!("mode={}", cfg.shard_mode.as_str()),
        format!("worker_id={}", cfg.worker_id),
        format!("max={}", cfg.max_cameras),
    ];
    if cfg.mediamtx_node_id > 0 {
        parts.push(format!("mtx_node={}", cfg.mediamtx_node_id));
    }
    if cfg.worker_shard_total > 0 && cfg.worker_shard_index >= 0 {
        parts.push(format!(
            "shard={}/{}",
            cfg.worker_shard_index, cfg.worker_shard_total
        ));
    }
    parts.join(" ")
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::config::ShardMode;

    fn base_cfg() -> Config {
        Config {
            confvision_api_url: "http://test".into(),
            vis_worker_api_key: String::new(),
            mediamtx_rtsp_base: "rtsp://localhost".into(),
            rtmp_publish_secret: None,
            processor_id: "proc-1".into(),
            processor_hostname: "host".into(),
            processor_version: "0.1.0".into(),
            worker_id: String::new(),
            worker_tipo: "rust_processor".into(),
            shard_mode: ShardMode::Auto,
            worker_shard_index: -1,
            worker_shard_total: 0,
            mediamtx_node_id: 0,
            redis_url: None,
            s3_endpoint: None,
            s3_bucket: None,
            http_host: "127.0.0.1".into(),
            http_port: 8090,
            log_level: "info".into(),
            max_cameras: 50,
            sync_interval: std::time::Duration::from_secs(60),
            ping_interval: std::time::Duration::from_secs(30),
            rtsp_connect_timeout: std::time::Duration::from_secs(5),
            rtsp_reconnect_base: std::time::Duration::from_secs(10),
            rtsp_frame_timeout: std::time::Duration::from_secs(30),
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
        }
    }

    fn cam(id: i64, ativo: bool, deteccao: bool) -> CameraRecord {
        CameraRecord {
            id,
            ativo,
            nome: None,
            rtsp_url_sec: None,
            mediamtx_rtsp_base: None,
            analitico_pausado: None,
            deteccao_humano: Some(deteccao),
            worker_id: None,
            extra: serde_json::Value::Null,
        }
    }

    #[test]
    fn sync_query_worker_id_only_in_worker_id_mode() {
        let mut cfg = base_cfg();
        cfg.worker_id = "w-1".into();
        cfg.shard_mode = ShardMode::Hash;
        assert!(sync_query_worker_id(&cfg).is_none());

        cfg.shard_mode = ShardMode::WorkerId;
        assert_eq!(sync_query_worker_id(&cfg), Some("w-1"));
    }

    #[test]
    fn hash_shard_selects_by_modulo() {
        let mut cfg = base_cfg();
        cfg.shard_mode = ShardMode::Hash;
        cfg.worker_shard_index = 1;
        cfg.worker_shard_total = 4;

        assert!(camera_belongs_to_shard(&cfg, 5)); // 5 % 4 == 1
        assert!(!camera_belongs_to_shard(&cfg, 4));
    }

    #[test]
    fn worker_id_mode_does_not_hash_locally() {
        let mut cfg = base_cfg();
        cfg.shard_mode = ShardMode::WorkerId;
        cfg.worker_shard_index = 1;
        cfg.worker_shard_total = 4;
        assert!(camera_belongs_to_shard(&cfg, 999));
    }

    #[test]
    fn auto_enables_shard_with_worker_id_or_index() {
        let mut cfg = base_cfg();
        cfg.shard_mode = ShardMode::Auto;
        assert!(!shard_enabled(&cfg));

        cfg.worker_id = "w".into();
        assert!(shard_enabled(&cfg));

        cfg.worker_id.clear();
        cfg.worker_shard_index = 0;
        cfg.worker_shard_total = 2;
        assert!(shard_enabled(&cfg));
    }

    #[test]
    fn max_cameras_truncates_after_filters() {
        let mut cfg = base_cfg();
        cfg.max_cameras = 2;
        let cameras = vec![cam(1, true, true), cam(2, true, true), cam(3, true, true)];
        let out = filter_analytic_cameras(&cfg, cameras);
        assert_eq!(out.len(), 2);
        assert_eq!(out[0].id, 1);
        assert_eq!(out[1].id, 2);
    }

    #[test]
    fn filter_respects_ativo_and_deteccao() {
        let cfg = base_cfg();
        let cameras = vec![cam(1, false, true), cam(2, true, false), cam(3, true, true)];
        let out = filter_analytic_cameras(&cfg, cameras);
        assert_eq!(out.len(), 1);
        assert_eq!(out[0].id, 3);
    }

    #[test]
    fn hash_filter_combined_with_max_cameras() {
        let mut cfg = base_cfg();
        cfg.shard_mode = ShardMode::Hash;
        cfg.worker_shard_index = 0;
        cfg.worker_shard_total = 2;
        cfg.max_cameras = 1;
        let cameras = vec![cam(2, true, true), cam(4, true, true), cam(6, true, true)];
        let out = filter_analytic_cameras(&cfg, cameras);
        assert_eq!(out.len(), 1);
        assert_eq!(out[0].id, 2);
    }

    #[test]
    fn no_shard_config_accepts_all_ids_for_hash_path() {
        let cfg = base_cfg();
        assert!(camera_belongs_to_shard(&cfg, 123));
    }

    #[test]
    fn processor_id_not_used_in_sync_query() {
        let mut cfg = base_cfg();
        cfg.processor_id = "processor-x".into();
        cfg.worker_id = "worker-y".into();
        cfg.shard_mode = ShardMode::WorkerId;
        assert_eq!(sync_query_worker_id(&cfg), Some("worker-y"));
    }

    #[test]
    fn auto_with_worker_id_does_not_send_worker_id_query() {
        let mut cfg = base_cfg();
        cfg.shard_mode = ShardMode::Auto;
        cfg.worker_id = "w-auto".into();
        assert!(sync_query_worker_id(&cfg).is_none());
        assert!(shard_enabled(&cfg));
        assert!(camera_belongs_to_shard(&cfg, 99));
    }

    #[test]
    fn auto_with_hash_applies_modulo_filter() {
        let mut cfg = base_cfg();
        cfg.shard_mode = ShardMode::Auto;
        cfg.worker_shard_index = 1;
        cfg.worker_shard_total = 3;
        assert!(!camera_belongs_to_shard(&cfg, 2)); // 2 % 3 != 1
        assert!(camera_belongs_to_shard(&cfg, 4)); // 4 % 3 == 1
    }

    #[test]
    fn shard_label_includes_mediamtx_node() {
        let mut cfg = base_cfg();
        cfg.mediamtx_node_id = 7;
        let label = shard_label(&cfg);
        assert!(label.contains("mtx_node=7"));
    }

    #[test]
    fn worker_id_mode_skips_hash_even_if_shard_numbers_set() {
        let mut cfg = base_cfg();
        cfg.shard_mode = ShardMode::WorkerId;
        cfg.worker_shard_index = 0;
        cfg.worker_shard_total = 10;
        let cameras = vec![cam(11, true, true), cam(21, true, true)];
        let out = filter_analytic_cameras(&cfg, cameras);
        assert_eq!(out.len(), 2);
    }
}
