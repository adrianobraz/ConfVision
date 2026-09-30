use serde::Serialize;

use crate::capacity::CapacitySnapshot;
use crate::decode::{
    decode_backend_effective, evaluate_hw_stack_grade, gpu_decode_state, AccelerationRuntime,
    DecodePolicyCoordinator,
};
use crate::load::{evaluate_load, LoadAdvisory, LoadPolicyConfig};

#[derive(Debug, Clone, Serialize)]
pub struct Phase62RuntimeView {
    pub gpu_decode_state: crate::decode::GpuDecodeState,
    pub decode_backend_effective: String,
    pub hw_stack_grade: crate::decode::HwStackGrade,
    pub load_advisory: LoadAdvisory,
    pub load_advisory_reason: String,
}

pub fn build_phase62_view(
    acceleration: &AccelerationRuntime,
    decode_policy: &DecodePolicyCoordinator,
    capacity: &CapacitySnapshot,
    load_config: &LoadPolicyConfig,
) -> Phase62RuntimeView {
    let load = evaluate_load(capacity, load_config);
    Phase62RuntimeView {
        gpu_decode_state: gpu_decode_state(acceleration, decode_policy, Some(capacity)),
        decode_backend_effective: decode_backend_effective(acceleration, decode_policy)
            .as_str()
            .into(),
        hw_stack_grade: evaluate_hw_stack_grade(acceleration),
        load_advisory: load.advisory,
        load_advisory_reason: load.reason,
    }
}
