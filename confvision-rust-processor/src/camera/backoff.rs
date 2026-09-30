use std::time::Duration;

/// Intervalo mínimo entre tentativas RTSP (evita busy-loop se env/base = 0).
pub const MIN_RECONNECT_INTERVAL: Duration = Duration::from_secs(1);

/// Backoff exponencial com teto (evita reconexão agressiva).
#[derive(Debug, Clone)]
pub struct ReconnectBackoff {
    base: Duration,
    max: Duration,
    attempt: u32,
}

impl ReconnectBackoff {
    pub fn new(base: Duration) -> Self {
        let base = base.max(MIN_RECONNECT_INTERVAL);
        Self {
            base,
            max: Duration::from_secs(120),
            attempt: 0,
        }
    }

    pub fn base_interval(&self) -> Duration {
        self.base
    }

    pub fn next_delay(&mut self) -> Duration {
        let exp = self.attempt.min(6);
        let mult = 1u64 << exp;
        let mut delay = self.base.saturating_mul(mult as u32);
        if delay > self.max {
            delay = self.max;
        }
        self.attempt = self.attempt.saturating_add(1);
        delay.max(MIN_RECONNECT_INTERVAL)
    }

    pub fn reset(&mut self) {
        self.attempt = 0;
    }

    pub fn attempts(&self) -> u32 {
        self.attempt
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn backoff_grows_and_caps() {
        let mut b = ReconnectBackoff::new(Duration::from_secs(2));
        assert_eq!(b.next_delay(), Duration::from_secs(2));
        assert_eq!(b.next_delay(), Duration::from_secs(4));
        b.reset();
        assert_eq!(b.attempts(), 0);
    }

    #[test]
    fn zero_base_is_clamped_to_minimum_interval() {
        let mut b = ReconnectBackoff::new(Duration::ZERO);
        assert_eq!(b.base_interval(), MIN_RECONNECT_INTERVAL);
        assert_eq!(b.next_delay(), MIN_RECONNECT_INTERVAL);
    }
}
