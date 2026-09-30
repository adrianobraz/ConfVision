//! Timelapse inteligente — Fase 2B.
//!
//! Scheduler sobre frames do pipeline (sem segundo RTSP).
//! Referência Python: `timelapse_worker.py`.

#[derive(Debug, Clone, Copy)]
pub struct TimelapseSchedulerConfig {
    pub frame_intervalo_seg: u64,
    pub frames_por_segmento: u32,
}

impl TimelapseSchedulerConfig {
    pub fn from_env_defaults() -> Self {
        Self {
            frame_intervalo_seg: 720,
            frames_por_segmento: 60,
        }
    }
}
