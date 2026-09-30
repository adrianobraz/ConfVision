use tokio::sync::watch;
use tokio::task::JoinHandle;

/// Controle de parada por câmera (independente do shutdown global do processo).
pub struct CameraWorkerControl {
    pub join: JoinHandle<()>,
    pub stop: watch::Sender<bool>,
}

impl CameraWorkerControl {
    pub fn new(join: JoinHandle<()>, stop: watch::Sender<bool>) -> Self {
        Self { join, stop }
    }
}

/// Cancelamento composto: shutdown do processo ou stop só desta câmera.
#[derive(Clone)]
pub struct CameraCancel {
    global: watch::Receiver<bool>,
    local: watch::Receiver<bool>,
}

impl CameraCancel {
    pub fn new(global: watch::Receiver<bool>, local: watch::Receiver<bool>) -> Self {
        Self { global, local }
    }

    pub fn is_cancelled(&self) -> bool {
        *self.global.borrow() || *self.local.borrow()
    }

    pub fn global(&self) -> watch::Receiver<bool> {
        self.global.clone()
    }

    /// Aguarda shutdown global ou stop local (para `select!` no loop RTSP).
    pub async fn wait_until_cancelled(&self) {
        if self.is_cancelled() {
            return;
        }
        let mut global = self.global.clone();
        let mut local = self.local.clone();
        tokio::select! {
            biased;
            res = global.changed() => {
                let _ = res;
            }
            res = local.changed() => {
                let _ = res;
            }
        }
    }
}
