use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::{Arc, RwLock};

/// Extradata H.264 (avcC) compartilhada entre RTSP (producer) e consumer da sessão.
#[derive(Default)]
pub struct SessionDecodeContext {
    extradata: RwLock<Option<Arc<[u8]>>>,
    generation: AtomicU64,
}

impl SessionDecodeContext {
    pub fn new() -> Arc<Self> {
        Arc::new(Self::default())
    }

    pub fn update_extradata(&self, data: &[u8]) {
        if data.is_empty() {
            return;
        }
        let mut guard = self.extradata.write().expect("decode extradata lock");
        let changed = guard.as_deref() != Some(data);
        if changed {
            *guard = Some(Arc::from(data));
            self.generation.fetch_add(1, Ordering::Relaxed);
        }
    }

    pub fn extradata(&self) -> Option<Arc<[u8]>> {
        self.extradata.read().ok()?.clone()
    }

    pub fn generation(&self) -> u64 {
        self.generation.load(Ordering::Relaxed)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn extradata_update_bumps_generation() {
        let ctx = SessionDecodeContext::new();
        assert_eq!(ctx.generation(), 0);
        ctx.update_extradata(&[1, 2, 3]);
        assert_eq!(ctx.generation(), 1);
        assert_eq!(ctx.extradata().unwrap().as_ref(), &[1, 2, 3]);
        ctx.update_extradata(&[1, 2, 3]);
        assert_eq!(ctx.generation(), 1);
        ctx.update_extradata(&[4]);
        assert_eq!(ctx.generation(), 2);
    }
}
