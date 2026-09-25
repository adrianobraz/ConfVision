use std::collections::HashMap;
use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::Arc;
use std::time::Duration;

use tokio::sync::{watch, RwLock};
use tokio::task::JoinHandle;
use tracing::{info, warn};

use crate::api::CameraRecord;
use crate::camera::stream_url::{redact_rtsp_url, resolve_rtsp_url};
use crate::camera::types::{CameraRuntimeState, SharedCameraState};
use crate::camera::worker_control::CameraWorkerControl;
use crate::config::Config;
use crate::decode::{AccelerationRuntime, DecodePolicyCoordinator};
use crate::load::LoadAdmissionGate;
use crate::metrics::ProcessorMetrics;
use crate::worker::camera_worker::run_camera_worker;

pub struct CameraManager {
    cfg: Config,
    metrics: Arc<ProcessorMetrics>,
    acceleration: Arc<AccelerationRuntime>,
    decode_policy: Arc<DecodePolicyCoordinator>,
    load_admission: Arc<LoadAdmissionGate>,
    /// Mapa só para listagem (health/metrics); cada câmera tem lock próprio em `SharedCameraState`.
    states: Arc<RwLock<HashMap<i64, SharedCameraState>>>,
    shutdown: watch::Sender<bool>,
    shutdown_rx: watch::Receiver<bool>,
    handles: Arc<RwLock<HashMap<i64, CameraWorkerControl>>>,
    global_frames: Arc<AtomicU64>,
}

impl CameraManager {
    pub fn new(
        cfg: Config,
        metrics: Arc<ProcessorMetrics>,
        acceleration: Arc<AccelerationRuntime>,
        decode_policy: Arc<DecodePolicyCoordinator>,
        load_admission: Arc<LoadAdmissionGate>,
    ) -> Self {
        let (shutdown, shutdown_rx) = watch::channel(false);
        Self {
            cfg,
            metrics,
            acceleration,
            decode_policy,
            load_admission,
            states: Arc::new(RwLock::new(HashMap::new())),
            shutdown,
            shutdown_rx,
            handles: Arc::new(RwLock::new(HashMap::new())),
            global_frames: Arc::new(AtomicU64::new(0)),
        }
    }

    pub fn states_handle(&self) -> Arc<RwLock<HashMap<i64, SharedCameraState>>> {
        self.states.clone()
    }

    pub fn shutdown_signal(&self) -> watch::Sender<bool> {
        self.shutdown.clone()
    }

    pub fn is_shutdown(&self) -> bool {
        *self.shutdown_rx.borrow()
    }

    pub async fn sync_cameras(&self, desired: Vec<CameraRecord>) {
        let mut limited = desired;
        let hard_max = self.cfg.effective_max_cameras();
        if limited.len() > hard_max {
            warn!(
                processor_id = %self.cfg.processor_id,
                total = limited.len(),
                max = hard_max,
                "truncating camera list (hard safety limit)"
            );
            limited.truncate(hard_max);
        }

        let desired_ids: Vec<i64> = limited.iter().map(|c| c.id).collect();
        let running: Vec<i64> = self.handles.read().await.keys().copied().collect();

        for id in running {
            if !desired_ids.contains(&id) {
                self.stop_camera(id).await;
            }
        }

        for cam in limited {
            if self.handles.read().await.contains_key(&cam.id) {
                continue;
            }
            if !self.load_admission.allow_new_camera().await {
                warn!(
                    camera_id = cam.id,
                    processor_id = %self.cfg.processor_id,
                    "load admission rejected new camera (existing sessions unchanged)"
                );
                continue;
            }
            let camera_id = cam.id;
            if let Err(e) = self.start_camera(cam).await {
                warn!(camera_id, error = %e, "failed to start camera");
            }
        }
    }

    async fn start_camera(&self, cam: CameraRecord) -> Result<(), String> {
        let rtsp_url = resolve_rtsp_url(&cam, &self.cfg).map_err(|e| e.to_string())?;
        let redacted = redact_rtsp_url(&rtsp_url);

        let state: SharedCameraState = Arc::new(RwLock::new(CameraRuntimeState::new(
            cam.id,
            redacted.clone(),
        )));
        {
            let mut states = self.states.write().await;
            states.insert(cam.id, state.clone());
        }

        info!(
            camera_id = cam.id,
            processor_id = %self.cfg.processor_id,
            url = %redacted,
            "worker started"
        );

        let cfg = self.cfg.clone();
        let metrics = self.metrics.clone();
        let acceleration = self.acceleration.clone();
        let decode_policy = self.decode_policy.clone();
        let mut shutdown_rx = self.shutdown_rx.clone();
        let global_frames = self.global_frames.clone();
        let (camera_stop_tx, camera_stop_rx) = watch::channel(false);

        let camera_id = cam.id;
        let join: JoinHandle<()> = tokio::spawn(async move {
            run_camera_worker(
                camera_id,
                rtsp_url,
                cfg,
                metrics,
                acceleration,
                decode_policy,
                state,
                &mut shutdown_rx,
                camera_stop_rx,
                global_frames,
            )
            .await;
        });

        self.handles
            .write()
            .await
            .insert(cam.id, CameraWorkerControl::new(join, camera_stop_tx));
        Ok(())
    }

    pub async fn stop_camera(&self, camera_id: i64) {
        if let Some(ctrl) = self.handles.write().await.remove(&camera_id) {
            let _ = ctrl.stop.send(true);
            let mut join = ctrl.join;
            if tokio::time::timeout(Duration::from_secs(15), &mut join)
                .await
                .is_err()
            {
                warn!(camera_id, "camera worker stop timed out — aborting task");
                join.abort();
            }
        }
        self.states.write().await.remove(&camera_id);
        info!(camera_id, processor_id = %self.cfg.processor_id, "worker stopped");
    }

    pub async fn stop_all(&self) {
        let _ = self.shutdown.send(true);
        let ids: Vec<i64> = self.handles.read().await.keys().copied().collect();
        for id in ids {
            self.stop_camera(id).await;
        }
    }

    pub async fn active_camera_count(&self) -> usize {
        self.handles.read().await.len()
    }

    pub fn total_frames(&self) -> u64 {
        self.global_frames.load(Ordering::Relaxed)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::api::CameraRecord;
    use crate::decode::{
        AccelerationPolicy, AccelerationRuntime, VideoAccelerationMode, VideoGpuBackend,
    };
    use serde_json::json;
    use std::time::Duration;

    fn test_cfg() -> Config {
        Config {
            confvision_api_url: "http://localhost".into(),
            vis_worker_api_key: String::new(),
            mediamtx_rtsp_base: "rtsp://localhost:8554".into(),
            rtmp_publish_secret: Some("secret".into()),
            processor_id: "test-proc".into(),
            processor_hostname: "host".into(),
            processor_version: "0.1.0".into(),
            worker_tipo: "rust_processor".into(),
            worker_id: "worker-01".into(),
            shard_mode: crate::config::ShardMode::Auto,
            worker_shard_index: -1,
            worker_shard_total: 0,
            mediamtx_node_id: 0,
            redis_url: None,
            s3_endpoint: None,
            s3_bucket: None,
            http_host: "127.0.0.1".into(),
            http_port: 8090,
            log_level: "info".into(),
            max_cameras: 2,
            sync_interval: Duration::from_secs(60),
            ping_interval: Duration::from_secs(30),
            rtsp_connect_timeout: Duration::from_secs(1),
            rtsp_reconnect_base: Duration::from_secs(1),
            rtsp_frame_timeout: Duration::from_secs(1),
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
            decode_runtime_fallback: true,
            decode_hw_error_threshold: 10,
            load_policy_mode: crate::load::LoadPolicyMode::Advisory,
            load_admission_enabled: false,
            motion_analysis_max_fps: 0.0,
            motion_frame_stride: 1,
            decode_frame_stride: 1,
        }
    }

    #[tokio::test]
    async fn manager_respects_max_cameras() {
        std::env::set_var("RTSP_SIMULATE", "1");
        let metrics = Arc::new(ProcessorMetrics::new("test"));
        let acceleration = AccelerationRuntime::bootstrap(AccelerationPolicy::from_parts(
            VideoAccelerationMode::Cpu,
            VideoGpuBackend::Auto,
        ))
        .unwrap();
        let decode_policy =
            crate::decode::DecodePolicyCoordinator::new(test_cfg().decode_fallback_config());
        let capacity = crate::capacity::CapacityEngine::new(&test_cfg());
        let admission =
            crate::load::LoadAdmissionGate::new(capacity, test_cfg().load_policy_config());
        let mgr = CameraManager::new(test_cfg(), metrics, acceleration, decode_policy, admission);
        let cams: Vec<CameraRecord> = (1..=5)
            .map(|id| CameraRecord {
                id,
                ativo: true,
                nome: None,
                rtsp_url_sec: Some(format!("rtsp://127.0.0.1:1/{id}")),
                mediamtx_rtsp_base: None,
                analitico_pausado: None,
                deteccao_humano: Some(true),
                worker_id: None,
                extra: json!({}),
            })
            .collect();
        mgr.sync_cameras(cams).await;
        tokio::time::sleep(Duration::from_millis(50)).await;
        assert!(mgr.active_camera_count().await <= 2);
        mgr.stop_all().await;
        std::env::remove_var("RTSP_SIMULATE");
    }

    #[tokio::test]
    async fn admission_reject_blocks_only_new_cameras() {
        use crate::capacity::{
            CapacitySnapshot, CapacityState, LimitingResource, NetworkSnapshot, ResourceMetric,
            ResourceScope, StorageSnapshot,
        };
        use crate::load::LoadPolicyMode;

        std::env::set_var("RTSP_SIMULATE", "1");
        let mut cfg = test_cfg();
        cfg.load_policy_mode = LoadPolicyMode::Admission;
        cfg.load_admission_enabled = true;
        cfg.max_cameras = 10;

        let metrics = Arc::new(ProcessorMetrics::new("test"));
        let acceleration = AccelerationRuntime::bootstrap(AccelerationPolicy::from_parts(
            VideoAccelerationMode::Cpu,
            VideoGpuBackend::Auto,
        ))
        .unwrap();
        let decode_policy = DecodePolicyCoordinator::new(cfg.decode_fallback_config());

        let m = ResourceMetric {
            percent: None,
            target_percent: 80.0,
            headroom_percent: None,
            scope: ResourceScope::Unavailable,
        };
        let healthy = CapacitySnapshot {
            mode: cfg.capacity_mode,
            state: CapacityState::Healthy,
            worker_id: cfg.worker_id.clone(),
            processor_id: cfg.processor_id.clone(),
            cpu: m.clone(),
            memory: m.clone(),
            gpu: m.clone(),
            vram: m.clone(),
            network: NetworkSnapshot {
                rx_bytes_total: None,
                tx_bytes_total: None,
                rx_bps: None,
                tx_bps: None,
                scope: ResourceScope::Unavailable,
            },
            storage: StorageSnapshot {
                disk_total_bytes: None,
                disk_used_bytes: None,
                disk_free_bytes: None,
                disk_used_percent: None,
                scope: ResourceScope::Unavailable,
            },
            current_cameras: 0,
            current_online_cameras: 0,
            current_fps: 0.0,
            average_fps_per_camera: None,
            current_frames_received: 0,
            current_frames_processed: 0,
            current_frames_dropped: 0,
            drop_rate_percent: None,
            queue_depth: 0,
            processing_latency_ms: 0,
            estimated_capacity_cameras: None,
            estimated_capacity_fps: None,
            estimated_available_cameras: Some(5),
            estimated_available_fps: None,
            capacity_used_percent: Some(50.0),
            limiting_resource: LimitingResource::None,
            max_cameras_safety_limit: None,
            cameras: vec![],
            observation_ready: true,
            samples_in_window: 1,
        };
        let critical = CapacitySnapshot {
            mode: cfg.capacity_mode,
            state: CapacityState::Critical,
            worker_id: cfg.worker_id.clone(),
            processor_id: cfg.processor_id.clone(),
            cpu: m.clone(),
            memory: m.clone(),
            gpu: m.clone(),
            vram: m,
            network: NetworkSnapshot {
                rx_bytes_total: None,
                tx_bytes_total: None,
                rx_bps: None,
                tx_bps: None,
                scope: ResourceScope::Unavailable,
            },
            storage: StorageSnapshot {
                disk_total_bytes: None,
                disk_used_bytes: None,
                disk_free_bytes: None,
                disk_used_percent: None,
                scope: ResourceScope::Unavailable,
            },
            current_cameras: 1,
            current_online_cameras: 1,
            current_fps: 0.0,
            average_fps_per_camera: None,
            current_frames_received: 0,
            current_frames_processed: 0,
            current_frames_dropped: 0,
            drop_rate_percent: None,
            queue_depth: 0,
            processing_latency_ms: 0,
            estimated_capacity_cameras: None,
            estimated_capacity_fps: None,
            estimated_available_cameras: Some(0),
            estimated_available_fps: None,
            capacity_used_percent: Some(100.0),
            limiting_resource: LimitingResource::Cpu,
            max_cameras_safety_limit: None,
            cameras: vec![],
            observation_ready: true,
            samples_in_window: 1,
        };
        let capacity = crate::capacity::CapacityEngine::new_with_initial_snapshot(&cfg, healthy);
        let admission =
            crate::load::LoadAdmissionGate::new(capacity.clone(), cfg.load_policy_config());
        let mgr = CameraManager::new(cfg, metrics, acceleration, decode_policy, admission);

        let cam1 = CameraRecord {
            id: 101,
            ativo: true,
            nome: None,
            rtsp_url_sec: Some("rtsp://127.0.0.1:1/101".into()),
            mediamtx_rtsp_base: None,
            analitico_pausado: None,
            deteccao_humano: Some(true),
            worker_id: None,
            extra: json!({}),
        };
        mgr.sync_cameras(vec![cam1.clone()]).await;
        tokio::time::sleep(Duration::from_millis(80)).await;
        assert_eq!(mgr.active_camera_count().await, 1);

        capacity.replace_snapshot_for_test(critical).await;

        let cam2 = CameraRecord {
            id: 102,
            ..cam1.clone()
        };
        mgr.sync_cameras(vec![cam1, cam2]).await;
        tokio::time::sleep(Duration::from_millis(50)).await;
        assert_eq!(mgr.active_camera_count().await, 1);

        mgr.stop_all().await;
        std::env::remove_var("RTSP_SIMULATE");
    }
}
