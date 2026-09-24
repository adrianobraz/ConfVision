mod api;
mod camera;
mod config;
mod decode;
mod error;
mod events;
mod health;
mod logging;
mod media;
mod metrics;
mod pipeline;
mod redis;
mod rtsp;
mod worker;

use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;

use axum::routing::get;
use axum::Router;
use chrono::Utc;
use tokio::signal;
use tracing::{error, info, warn};

use crate::api::{ConfVisionClient, WorkerPingRequest};
use crate::camera::CameraManager;
use crate::config::Config;
use crate::health::{health_handler, metrics_handler, ready_handler, AppState};
use crate::metrics::ProcessorMetrics;

#[tokio::main]
async fn main() {
    dotenvy::dotenv().ok();
    let cfg = match Config::from_env() {
        Ok(c) => c,
        Err(e) => {
            eprintln!("config error: {e}");
            std::process::exit(1);
        }
    };

    logging::init_logging(&cfg.log_level);
    redis::log_redis_status(&cfg);
    media::log_media_status(&cfg);

    info!(
        processor_id = %cfg.processor_id,
        api = %cfg.confvision_api_url,
        "confvision-rust-processor starting"
    );

    let metrics = Arc::new(ProcessorMetrics::new(cfg.processor_id.clone()));
    let manager = Arc::new(CameraManager::new(cfg.clone(), metrics.clone()));
    let api_ready = Arc::new(AtomicBool::new(false));

    let app_state = AppState {
        metrics: metrics.clone(),
        camera_states: manager.states_handle(),
        api_ready: api_ready.clone(),
    };

    let app = Router::new()
        .route("/health", get(health_handler))
        .route("/ready", get(ready_handler))
        .route("/metrics", get(metrics_handler))
        .with_state(app_state);

    let bind = format!("{}:{}", cfg.http_host, cfg.http_port);
    let listener = tokio::net::TcpListener::bind(&bind)
        .await
        .unwrap_or_else(|e| {
            error!(bind = %bind, error = %e, "failed to bind HTTP");
            std::process::exit(1);
        });

    let http = axum::serve(listener, app);
    let http_handle = tokio::spawn(async move {
        if let Err(e) = http.await {
            error!(error = %e, "HTTP server error");
        }
    });

    let client = match ConfVisionClient::new(&cfg) {
        Ok(c) => {
            api_ready.store(true, Ordering::Relaxed);
            c
        }
        Err(e) => {
            warn!(error = %e, "API client init failed");
            std::process::exit(1);
        }
    };

    let sync_cfg = cfg.clone();
    let sync_client = client.clone();
    let sync_manager = manager.clone();
    let sync_handle = tokio::spawn(async move {
        run_sync_loop(sync_client, sync_manager, sync_cfg).await;
    });

    let ping_cfg = cfg.clone();
    let ping_client = client.clone();
    let ping_manager = manager.clone();
    let ping_metrics = metrics.clone();
    let ping_handle = tokio::spawn(async move {
        run_ping_loop(ping_client, ping_manager, ping_metrics, ping_cfg).await;
    });

    shutdown_signal().await;
    info!(processor_id = %cfg.processor_id, "shutdown initiated");

    let _ = manager.shutdown_signal().send(true);
    manager.stop_all().await;

    sync_handle.abort();
    ping_handle.abort();
    http_handle.abort();

    send_final_ping(&client, &cfg, &manager, &metrics).await;
    info!(processor_id = %cfg.processor_id, "shutdown complete");
}

async fn shutdown_signal() {
    let ctrl_c = async {
        signal::ctrl_c()
            .await
            .expect("failed to install Ctrl+C handler");
    };

    #[cfg(unix)]
    let terminate = async {
        signal::unix::signal(signal::unix::SignalKind::terminate())
            .expect("failed to install SIGTERM handler")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => {},
        _ = terminate => {},
    }
}

async fn run_sync_loop(client: ConfVisionClient, manager: Arc<CameraManager>, cfg: Config) {
    let mut config_version: Option<String> = None;
    loop {
        if manager.is_shutdown() {
            break;
        }

        let worker_filter = if cfg.sync_filter_worker_id {
            Some(cfg.processor_id.as_str())
        } else {
            None
        };

        match client
            .sync_cameras_ativas(worker_filter, cfg.mediamtx_node_id, config_version.as_deref())
            .await
        {
            Ok(resp) => {
                if let Some(v) = resp.config_version {
                    config_version = Some(v);
                }
                if !resp.unchanged {
                    manager.sync_cameras(resp.cameras).await;
                }
            }
            Err(e) => {
                warn!(processor_id = %cfg.processor_id, error = %e, "camera sync failed");
            }
        }

        tokio::time::sleep(cfg.sync_interval).await;
    }
}

async fn run_ping_loop(
    client: ConfVisionClient,
    manager: Arc<CameraManager>,
    metrics: Arc<ProcessorMetrics>,
    cfg: Config,
) {
    loop {
        if manager.is_shutdown() {
            break;
        }

        let cameras_ativas = manager.active_camera_count().await as i32;
        let node_id = if cfg.mediamtx_node_id > 0 {
            Some(cfg.mediamtx_node_id as i32)
        } else {
            None
        };

        let ping = WorkerPingRequest {
            worker_id: cfg.processor_id.clone(),
            worker_tipo: cfg.worker_tipo.clone(),
            hostname: cfg.processor_hostname.clone(),
            versao: cfg.processor_version.clone(),
            cameras_ativas,
            ultimo_ping_em: Utc::now().to_rfc3339(),
            ativo: true,
            yolo_device: "none".to_string(),
            queue_backend: cfg.queue_backend.clone(),
            vis_mediamtx_node_id: node_id,
            max_cameras: Some(cfg.max_cameras as i32),
        };

        if let Err(e) = client.worker_ping(&ping).await {
            warn!(processor_id = %cfg.processor_id, error = %e, "worker ping failed");
            metrics.errors.fetch_add(1, Ordering::Relaxed);
        }

        tokio::time::sleep(cfg.ping_interval).await;
    }
}

async fn send_final_ping(
    client: &ConfVisionClient,
    cfg: &Config,
    manager: &CameraManager,
    metrics: &ProcessorMetrics,
) {
    let ping = WorkerPingRequest {
        worker_id: cfg.processor_id.clone(),
        worker_tipo: cfg.worker_tipo.clone(),
        hostname: cfg.processor_hostname.clone(),
        versao: cfg.processor_version.clone(),
        cameras_ativas: 0,
        ultimo_ping_em: Utc::now().to_rfc3339(),
        ativo: false,
        yolo_device: "none".into(),
        queue_backend: cfg.queue_backend.clone(),
        vis_mediamtx_node_id: None,
        max_cameras: Some(cfg.max_cameras as i32),
    };
    if let Err(e) = client.worker_ping(&ping).await {
        warn!(error = %e, "final ping failed");
        metrics.errors.fetch_add(1, Ordering::Relaxed);
    }
    let _ = manager;
}
