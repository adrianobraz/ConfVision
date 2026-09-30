use std::collections::HashSet;
use std::sync::Arc;
use std::time::{Duration, Instant};

use tokio::sync::RwLock;
use tracing::{info, warn};

use crate::capacity::{CapacityEngine, CapacitySnapshot, CapacityState};
use crate::camera::CameraManager;
use crate::config::Config;

struct SheddingRuntime {
    over_since: Option<Instant>,
    under_since: Option<Instant>,
}

pub struct LoadSheddingCoordinator {
    cfg: Config,
    capacity: Arc<CapacityEngine>,
    shed_ids: Arc<RwLock<HashSet<i64>>>,
    runtime: tokio::sync::Mutex<SheddingRuntime>,
}

impl LoadSheddingCoordinator {
    pub fn new(cfg: Config, capacity: Arc<CapacityEngine>) -> Arc<Self> {
        Arc::new(Self {
            cfg,
            capacity,
            shed_ids: Arc::new(RwLock::new(HashSet::new())),
            runtime: tokio::sync::Mutex::new(SheddingRuntime {
                over_since: None,
                under_since: None,
            }),
        })
    }

    pub fn shed_ids_handle(&self) -> Arc<RwLock<HashSet<i64>>> {
        self.shed_ids.clone()
    }

    pub async fn is_shed(&self, camera_id: i64) -> bool {
        self.shed_ids.read().await.contains(&camera_id)
    }

    /// Só CPU/RAM reais — não usar `CapacityState::Critical` (100% do teto planejado ≠ overload).
    fn resource_hot(snap: &CapacitySnapshot, enter_pct: f64) -> bool {
        let cpu = snap.cpu.percent.unwrap_or(0.0);
        let mem = snap.memory.percent.unwrap_or(0.0);
        cpu >= enter_pct || mem >= enter_pct
    }

    fn resource_cool(snap: &CapacitySnapshot, exit_pct: f64) -> bool {
        let cpu = snap.cpu.percent.unwrap_or(100.0);
        let mem = snap.memory.percent.unwrap_or(100.0);
        cpu <= exit_pct && mem <= exit_pct && snap.state != CapacityState::Unknown
    }

    pub async fn tick(&self, manager: &Arc<CameraManager>) {
        if !self.cfg.load_shedding_enabled {
            let mut shed = self.shed_ids.write().await;
            if !shed.is_empty() {
                shed.clear();
                info!("load shedding desligado: lista de bloqueio liberada (reabertura no sync)");
            }
            return;
        }
        let snap = self.capacity.snapshot().await;
        let enter = self.cfg.load_shed_enter_percent;
        let exit = self.cfg.load_shed_exit_percent;
        let mut rt = self.runtime.lock().await;

        if Self::resource_hot(&snap, enter) {
            rt.under_since = None;
            if rt.over_since.is_none() {
                rt.over_since = Some(Instant::now());
            }
            let sustained = rt
                .over_since
                .map(|t| t.elapsed() >= Duration::from_secs(self.cfg.load_shed_enter_secs))
                .unwrap_or(false);
            if sustained && manager.active_camera_count().await > 0 {
                if let Some(id) = manager.pick_camera_to_shed().await {
                    manager.stop_camera(id).await;
                    self.shed_ids.write().await.insert(id);
                    warn!(
                        camera_id = id,
                        cpu = ?snap.cpu.percent,
                        mem = ?snap.memory.percent,
                        "load shedding: analítico suspenso por CPU/RAM"
                    );
                    rt.over_since = Some(Instant::now());
                }
            }
            return;
        }

        rt.over_since = None;
        if Self::resource_cool(&snap, exit) {
            if rt.under_since.is_none() {
                rt.under_since = Some(Instant::now());
            }
            let ok = rt
                .under_since
                .map(|t| t.elapsed() >= Duration::from_secs(self.cfg.load_shed_exit_secs))
                .unwrap_or(false);
            if ok {
                let mut shed = self.shed_ids.write().await;
                if let Some(id) = shed.iter().copied().next() {
                    shed.remove(&id);
                    info!(
                        camera_id = id,
                        "load shedding: câmera elegível para reabrir no próximo sync"
                    );
                }
                rt.under_since = None;
            }
        } else {
            rt.under_since = None;
        }
    }
}

pub fn spawn_load_shedding_loop(
    coordinator: Arc<LoadSheddingCoordinator>,
    manager: Arc<CameraManager>,
    interval: Duration,
) {
    tokio::spawn(async move {
        let mut ticker = tokio::time::interval(interval);
        loop {
            ticker.tick().await;
            coordinator.tick(&manager).await;
        }
    });
}
