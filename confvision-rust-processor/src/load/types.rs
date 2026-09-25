use serde::Serialize;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum LoadPolicyMode {
    Advisory,
    Admission,
    Disabled,
}

impl LoadPolicyMode {
    pub fn parse(raw: &str) -> Result<Self, String> {
        match raw.trim().to_lowercase().as_str() {
            "advisory" => Ok(Self::Advisory),
            "admission" => Ok(Self::Admission),
            "disabled" => Ok(Self::Disabled),
            other => Err(format!(
                "LOAD_POLICY_MODE inválido: {other:?} (use advisory, admission ou disabled)"
            )),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum LoadAdvisory {
    Normal,
    Caution,
    Saturated,
    RejectAdmission,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct LoadAdvisoryResult {
    pub advisory: LoadAdvisory,
    pub reason: String,
}

#[derive(Debug, Clone, Copy)]
pub struct LoadPolicyConfig {
    pub mode: LoadPolicyMode,
    pub admission_enabled: bool,
}

impl LoadPolicyConfig {
    /// Rejeição explícita em métricas (`reject_admission`) — modo admission + flag.
    pub fn admission_active(&self) -> bool {
        matches!(self.mode, LoadPolicyMode::Admission) && self.admission_enabled
    }

    /// Bloqueio de novas câmeras em `critical` (Fase 6.2): basta `LOAD_ADMISSION_ENABLED=1`.
    pub fn admission_blocks_new_cameras(&self, capacity: &crate::capacity::CapacitySnapshot) -> bool {
        if matches!(self.mode, LoadPolicyMode::Disabled) || !self.admission_enabled {
            return false;
        }
        capacity.state == crate::capacity::CapacityState::Critical
    }
}
