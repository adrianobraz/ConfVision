//! Deduplicação / cooldown entre publicações analíticas (paridade Python worker).

/// Gate por câmera: impede vários `EventJob` dentro do `cooldown_seg` da câmera.
#[derive(Debug, Clone, Copy, Default)]
pub struct EmitCooldownGate {
    last_emit_epoch: f64,
}

impl EmitCooldownGate {
    pub fn new() -> Self {
        Self::default()
    }

    /// Retorna `true` se a emissão é permitida e atualiza o último instante.
    pub fn try_emit(&mut self, now_epoch: f64, cooldown_sec: u64) -> bool {
        if cooldown_sec == 0 {
            self.last_emit_epoch = now_epoch;
            return true;
        }
        if now_epoch - self.last_emit_epoch < cooldown_sec as f64 {
            return false;
        }
        self.last_emit_epoch = now_epoch;
        true
    }

    pub fn last_emit_epoch(&self) -> f64 {
        self.last_emit_epoch
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn blocks_within_cooldown() {
        let mut g = EmitCooldownGate::new();
        assert!(g.try_emit(100.0, 30));
        assert!(!g.try_emit(110.0, 30));
        assert!(g.try_emit(131.0, 30));
    }

    #[test]
    fn zero_cooldown_always_emits() {
        let mut g = EmitCooldownGate::new();
        assert!(g.try_emit(1.0, 0));
        assert!(g.try_emit(1.0, 0));
    }
}
