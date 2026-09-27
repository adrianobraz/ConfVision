use std::sync::atomic::{AtomicBool, AtomicU64, Ordering};
use std::sync::Arc;

use serde_json::to_string;
use tokio::sync::Mutex;
use tracing::{debug, warn};

use crate::config::Config;
use crate::error::{AppError, AppResult};

use super::job::{DlqEnvelope, EventJob};
use serde_json::from_str;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum PublishOutcome {
    Ok,
    QueueFull,
    BackendDisabled,
    Error,
}

#[derive(Default)]
pub struct EventQueueStats {
    pub redis_connected: AtomicBool,
    pub main_depth: AtomicU64,
    pub dlq_depth: AtomicU64,
    pub published: AtomicU64,
    pub queue_full: AtomicU64,
    pub errors: AtomicU64,
}

pub struct EventQueueHandle {
    backend: QueueBackend,
    stats: Arc<EventQueueStats>,
}

enum QueueBackend {
    None,
    Memory {
        inner: Mutex<MemoryQueue>,
    },
    Redis {
        client: redis::Client,
        key: String,
        dlq_key: String,
        max_size: usize,
        max_retries: u32,
    },
}

struct MemoryQueue {
    items: std::collections::VecDeque<EventJob>,
    max_size: usize,
}

impl EventQueueHandle {
    pub fn from_config(cfg: &Config) -> AppResult<Self> {
        let stats = Arc::new(EventQueueStats::default());
        let backend = match cfg.queue_backend.as_str() {
            "none" => QueueBackend::None,
            "memory" => QueueBackend::Memory {
                inner: Mutex::new(MemoryQueue {
                    items: std::collections::VecDeque::new(),
                    max_size: cfg.event_queue_max_size,
                }),
            },
            "redis" => {
                let url = cfg.redis_url.as_ref().ok_or_else(|| {
                    AppError::Config(
                        "QUEUE_BACKEND=redis exige REDIS_URL (Fase D2)".into(),
                    )
                })?;
                let client = redis::Client::open(url.as_str()).map_err(|e| {
                    AppError::Config(format!("REDIS_URL inválida: {e}"))
                })?;
                QueueBackend::Redis {
                    client,
                    key: cfg.event_queue_key.clone(),
                    dlq_key: cfg.event_queue_dlq_key.clone(),
                    max_size: cfg.event_queue_max_size,
                    max_retries: cfg.queue_publish_retries,
                }
            }
            other => {
                return Err(AppError::Config(format!(
                    "QUEUE_BACKEND inválido: {other:?} (use none, memory ou redis)"
                )));
            }
        };
        Ok(Self { backend, stats })
    }

    pub fn stats(&self) -> Arc<EventQueueStats> {
        self.stats.clone()
    }

    pub fn backend_label(&self) -> &'static str {
        match self.backend {
            QueueBackend::None => "none",
            QueueBackend::Memory { .. } => "memory",
            QueueBackend::Redis { .. } => "redis",
        }
    }

    pub async fn ping_redis(&self) -> bool {
        let QueueBackend::Redis { client, .. } = &self.backend else {
            return false;
        };
        let ok = redis_ping(client).await.is_ok();
        self.stats.redis_connected.store(ok, Ordering::Relaxed);
        ok
    }

    pub async fn refresh_depths(&self) {
        let QueueBackend::Redis {
            client,
            key,
            dlq_key,
            ..
        } = &self.backend
        else {
            return;
        };
        if let Ok((main, dlq)) = redis_llen_pair(client, key, dlq_key).await {
            self.stats.main_depth.store(main, Ordering::Relaxed);
            self.stats.dlq_depth.store(dlq, Ordering::Relaxed);
            self.stats.redis_connected.store(true, Ordering::Relaxed);
        } else {
            self.stats.redis_connected.store(false, Ordering::Relaxed);
        }
    }

    pub async fn pop(&self, timeout_sec: f64) -> AppResult<Option<EventJob>> {
        match &self.backend {
            QueueBackend::None => Ok(None),
            QueueBackend::Memory { inner } => {
                let mut q = inner.lock().await;
                if let Some(job) = q.items.pop_front() {
                    self.stats
                        .main_depth
                        .store(q.items.len() as u64, Ordering::Relaxed);
                    return Ok(Some(job));
                }
                drop(q);
                tokio::time::sleep(std::time::Duration::from_secs_f64(timeout_sec.max(0.1)))
                    .await;
                Ok(None)
            }
            QueueBackend::Redis { client, key, .. } => {
                redis_brpop(client, key, timeout_sec).await
            }
        }
    }

    pub async fn publish(&self, job: EventJob) -> PublishOutcome {
        match &self.backend {
            QueueBackend::None => PublishOutcome::BackendDisabled,
            QueueBackend::Memory { inner } => {
                let mut q = inner.lock().await;
                if q.items.len() >= q.max_size {
                    self.stats.queue_full.fetch_add(1, Ordering::Relaxed);
                    return PublishOutcome::QueueFull;
                }
                q.items.push_back(job);
                self.stats.published.fetch_add(1, Ordering::Relaxed);
                self.stats
                    .main_depth
                    .store(q.items.len() as u64, Ordering::Relaxed);
                PublishOutcome::Ok
            }
            QueueBackend::Redis {
                client,
                key,
                dlq_key,
                max_size,
                max_retries,
            } => {
                match redis_publish_with_retry(
                    client,
                    key,
                    dlq_key,
                    *max_size,
                    *max_retries,
                    &job,
                )
                .await
                {
                    Ok(PublishOutcome::Ok) => {
                        self.stats.published.fetch_add(1, Ordering::Relaxed);
                        let _ = self.refresh_depths().await;
                        PublishOutcome::Ok
                    }
                    Ok(PublishOutcome::QueueFull) => {
                        self.stats.queue_full.fetch_add(1, Ordering::Relaxed);
                        let _ = push_dlq(
                            client,
                            dlq_key,
                            DlqEnvelope {
                                reason: "queue_full".into(),
                                job,
                                at: chrono::Utc::now().timestamp_millis() as f64 / 1000.0,
                            },
                        )
                        .await;
                        PublishOutcome::QueueFull
                    }
                    Ok(other) => other,
                    Err(e) => {
                        warn!(error = %e, "event queue redis publish failed");
                        self.stats.errors.fetch_add(1, Ordering::Relaxed);
                        let _ = push_dlq(
                            client,
                            dlq_key,
                            DlqEnvelope {
                                reason: format!("redis_error:{e}"),
                                job,
                                at: chrono::Utc::now().timestamp_millis() as f64 / 1000.0,
                            },
                        )
                        .await;
                        PublishOutcome::Error
                    }
                }
            }
        }
    }
}

async fn redis_publish_with_retry(
    client: &redis::Client,
    key: &str,
    _dlq_key: &str,
    max_size: usize,
    max_retries: u32,
    job: &EventJob,
) -> AppResult<PublishOutcome> {
    let payload = to_string(job).map_err(|e| AppError::Other(e.into()))?;
    let mut last_err = None;
    for attempt in 0..max_retries.max(1) {
        match redis_lpush_if_room(client, key, max_size, &payload).await {
            Ok(PublishOutcome::Ok) => return Ok(PublishOutcome::Ok),
            Ok(PublishOutcome::QueueFull) => return Ok(PublishOutcome::QueueFull),
            Ok(other) => return Ok(other),
            Err(e) => {
                last_err = Some(e);
                if attempt + 1 < max_retries.max(1) {
                    tokio::time::sleep(std::time::Duration::from_millis(50 * (attempt as u64 + 1)))
                        .await;
                }
            }
        }
    }
    Err(last_err.unwrap_or_else(|| AppError::Config("redis publish failed".into())))
}

async fn redis_lpush_if_room(
    client: &redis::Client,
    key: &str,
    max_size: usize,
    payload: &str,
) -> AppResult<PublishOutcome> {
    let mut conn = client
        .get_multiplexed_async_connection()
        .await
        .map_err(|e| AppError::Other(e.into()))?;
    let len: usize = redis::cmd("LLEN")
        .arg(key)
        .query_async(&mut conn)
        .await
        .map_err(|e| AppError::Other(e.into()))?;
    if len >= max_size {
        return Ok(PublishOutcome::QueueFull);
    }
    redis::cmd("LPUSH")
        .arg(key)
        .arg(payload)
        .query_async::<()>(&mut conn)
        .await
        .map_err(|e| AppError::Other(e.into()))?;
    Ok(PublishOutcome::Ok)
}

async fn push_dlq(client: &redis::Client, dlq_key: &str, envelope: DlqEnvelope) -> AppResult<()> {
    let payload = to_string(&envelope).map_err(|e| AppError::Other(e.into()))?;
    let mut conn = client
        .get_multiplexed_async_connection()
        .await
        .map_err(|e| AppError::Other(e.into()))?;
    redis::cmd("LPUSH")
        .arg(dlq_key)
        .arg(payload)
        .query_async::<()>(&mut conn)
        .await
        .map_err(|e| AppError::Other(e.into()))?;
    debug!(dlq_key = %dlq_key, reason = %envelope.reason, "event pushed to dlq");
    Ok(())
}

async fn redis_ping(client: &redis::Client) -> AppResult<()> {
    let mut conn = client
        .get_multiplexed_async_connection()
        .await
        .map_err(|e| AppError::Other(e.into()))?;
    redis::cmd("PING")
        .query_async::<String>(&mut conn)
        .await
        .map_err(|e| AppError::Other(e.into()))?;
    Ok(())
}

async fn redis_brpop(client: &redis::Client, key: &str, timeout_sec: f64) -> AppResult<Option<EventJob>> {
    let mut conn = client
        .get_multiplexed_async_connection()
        .await
        .map_err(|e| AppError::Other(e.into()))?;
    let timeout = timeout_sec.max(1.0) as usize;
    let item: Option<(String, String)> = redis::cmd("BRPOP")
        .arg(key)
        .arg(timeout)
        .query_async(&mut conn)
        .await
        .map_err(|e| AppError::Other(e.into()))?;
    let Some((_key, payload)) = item else {
        return Ok(None);
    };
    let job: EventJob = from_str(&payload).map_err(|e| AppError::Other(e.into()))?;
    Ok(Some(job))
}

async fn redis_llen_pair(
    client: &redis::Client,
    key: &str,
    dlq_key: &str,
) -> AppResult<(u64, u64)> {
    let mut conn = client
        .get_multiplexed_async_connection()
        .await
        .map_err(|e| AppError::Other(e.into()))?;
    let main: u64 = redis::cmd("LLEN")
        .arg(key)
        .query_async(&mut conn)
        .await
        .map_err(|e| AppError::Other(e.into()))?;
    let dlq: u64 = redis::cmd("LLEN")
        .arg(dlq_key)
        .query_async(&mut conn)
        .await
        .map_err(|e| AppError::Other(e.into()))?;
    Ok((main, dlq))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn memory_queue_respects_max_size() {
        let mut cfg = Config::test_stub();
        cfg.queue_backend = "memory".into();
        cfg.event_queue_max_size = 2;
        let q = EventQueueHandle::from_config(&cfg).unwrap();
        assert_eq!(
            q.publish(EventJob::new(1, 0.9)).await,
            PublishOutcome::Ok
        );
        assert_eq!(
            q.publish(EventJob::new(2, 0.9)).await,
            PublishOutcome::Ok
        );
        assert_eq!(
            q.publish(EventJob::new(3, 0.9)).await,
            PublishOutcome::QueueFull
        );
    }

    #[tokio::test]
    async fn none_backend_disabled() {
        let cfg = Config::test_stub();
        let q = EventQueueHandle::from_config(&cfg).unwrap();
        assert_eq!(
            q.publish(EventJob::new(1, 0.5)).await,
            PublishOutcome::BackendDisabled
        );
    }
}
