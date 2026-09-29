use std::sync::atomic::{AtomicU64, AtomicUsize, Ordering};
use std::sync::{Mutex, OnceLock};

/// Pool global de `Vec<u8>` reutilizáveis (reduz churn do allocator em muitas câmeras).
pub struct BufferPool {
    max_idle: usize,
    idle: Mutex<Vec<Vec<u8>>>,
    pub hits: AtomicU64,
    pub misses: AtomicU64,
    pub returns: AtomicU64,
    pub drops: AtomicU64,
    pub idle_count: AtomicUsize,
}

impl BufferPool {
    pub fn new(max_idle: usize) -> Self {
        Self {
            max_idle: max_idle.max(8),
            idle: Mutex::new(Vec::new()),
            hits: AtomicU64::new(0),
            misses: AtomicU64::new(0),
            returns: AtomicU64::new(0),
            drops: AtomicU64::new(0),
            idle_count: AtomicUsize::new(0),
        }
    }

    pub fn acquire(&self, min_capacity: usize) -> Vec<u8> {
        let min_capacity = min_capacity.max(64);
        if let Ok(mut idle) = self.idle.lock() {
            while let Some(mut buf) = idle.pop() {
                self.idle_count.store(idle.len(), Ordering::Relaxed);
                if buf.capacity() >= min_capacity {
                    buf.clear();
                    self.hits.fetch_add(1, Ordering::Relaxed);
                    return buf;
                }
            }
            self.idle_count.store(0, Ordering::Relaxed);
        }
        self.misses.fetch_add(1, Ordering::Relaxed);
        Vec::with_capacity(min_capacity)
    }

    pub fn release(&self, mut buf: Vec<u8>, min_capacity: usize) {
        let min_capacity = min_capacity.max(64);
        if buf.capacity() < min_capacity {
            return;
        }
        buf.clear();
        self.returns.fetch_add(1, Ordering::Relaxed);
        if let Ok(mut idle) = self.idle.lock() {
            if idle.len() < self.max_idle {
                idle.push(buf);
                self.idle_count.store(idle.len(), Ordering::Relaxed);
                return;
            }
        }
        self.drops.fetch_add(1, Ordering::Relaxed);
    }
}

static GLOBAL_POOL: OnceLock<BufferPool> = OnceLock::new();

pub fn init_global_pool(max_idle: usize) {
    let _ = GLOBAL_POOL.set(BufferPool::new(max_idle));
}

pub fn global_pool() -> &'static BufferPool {
    GLOBAL_POOL.get_or_init(|| BufferPool::new(128))
}

pub fn acquire(min_capacity: usize) -> Vec<u8> {
    global_pool().acquire(min_capacity)
}

pub fn release(buf: Vec<u8>, min_capacity: usize) {
    global_pool().release(buf, min_capacity);
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn reuses_buffer_when_capacity_matches() {
        let pool = BufferPool::new(4);
        let v = pool.acquire(1024);
        assert!(v.capacity() >= 1024);
        pool.release(v, 1024);
        let v2 = pool.acquire(1024);
        assert!(v2.capacity() >= 1024);
        assert_eq!(pool.hits.load(Ordering::Relaxed), 1);
    }
}
