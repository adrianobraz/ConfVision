use std::time::Duration;

/// Backoff exponencial com teto (evita reconexão agressiva).
#[derive(Debug, Clone)]
pub struct ReconnectBackoff {
    base: Duration,
    max: Duration,
    attempt: u32,
}

impl ReconnectBackoff {
    pub fn new(base: Duration) -> Self {
        Self {
            base,
            max: Duration::from_secs(120),
            attempt: 0,
        }
    }

    pub fn next_delay(&mut self) -> Duration {
        let exp = self.attempt.min(6);
        let mult = 1u64 << exp;
        let mut delay = self.base.saturating_mul(mult as u32);
        if delay > self.max {
            delay = self.max;
        }
        self.attempt = self.attempt.saturating_add(1);
        delay
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
}
