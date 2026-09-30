use std::collections::HashMap;
use std::sync::Arc;

use tokio::sync::{Mutex, OwnedMutexGuard};

type CameraLock = Arc<Mutex<()>>;

#[derive(Clone, Default)]
pub struct CaptureLocks {
    inner: Arc<Mutex<HashMap<i64, CameraLock>>>,
}

impl CaptureLocks {
    pub fn new() -> Arc<Self> {
        Arc::new(Self::default())
    }

    pub async fn try_acquire(self: &Arc<Self>, camera_id: i64) -> Option<CaptureGuard> {
        let lock = {
            let mut map = self.inner.lock().await;
            map.entry(camera_id)
                .or_insert_with(|| Arc::new(Mutex::new(())))
                .clone()
        };
        let guard = lock.clone().try_lock_owned().ok()?;
        Some(CaptureGuard {
            camera_id,
            _guard: guard,
        })
    }
}

pub struct CaptureGuard {
    pub camera_id: i64,
    _guard: OwnedMutexGuard<()>,
}
