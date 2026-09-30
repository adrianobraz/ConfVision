//! Captura acionada por evento sensor — Fase 2C.
//!
//! Python: poll `get_eventos_sensor_pendentes` + RTSP dedicado em `event_capture.py`.
//! Meta: frame atual do pipeline Rust quando a câmera já está no processor.

#[derive(Debug, Clone, Copy)]
pub struct SensorEventRef {
    pub evento_id: i64,
    pub camera_id: i64,
}
