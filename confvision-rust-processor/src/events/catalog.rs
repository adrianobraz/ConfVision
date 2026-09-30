//! Catálogo de tipos de evento **reais** no ConfVision (não inventar novos contratos aqui).
//!
//! O pipeline analítico Rust hoje materializa principalmente `AnalyticPerson` via fila + API Go.
//! Motion gravação, timelapse e sensor usam outros caminhos (ver `docs/FASE-3-PROCESSAMENTO-E-EVENTOS.md`).

/// Classificação lógica para observabilidade interna (não serializada na fila Redis).
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ConfVisionEventKind {
    /// YOLO + regras → `EventJob` → capture → `vis_evento` (tipo_deteccao da câmera).
    AnalyticDetection,
    /// Sensor: evento criado no receptor → poll Python/Rust futuro → capture.
    SensorPendingCapture,
    /// Gravação segmento movimento (`vis_gravacao_segmento` / motion worker) — fora da fila analítica.
    MotionRecordingSegment,
    /// Timelapse segment upload — fora da fila analítica.
    TimelapseSegment,
    /// Operacional câmera (health/sync) — não persiste como vis_evento analítico.
    CameraOperational,
}

impl ConfVisionEventKind {
    pub fn as_str(self) -> &'static str {
        match self {
            Self::AnalyticDetection => "analytic_detection",
            Self::SensorPendingCapture => "sensor_pending_capture",
            Self::MotionRecordingSegment => "motion_recording_segment",
            Self::TimelapseSegment => "timelapse_segment",
            Self::CameraOperational => "camera_operational",
        }
    }
}
