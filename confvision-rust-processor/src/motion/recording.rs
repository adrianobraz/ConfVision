//! Coordenação de clip por movimento (FFmpeg externo) — Fase 2A.
//!
//! Python (`motion_worker.py`): REC START/STOP + `process_segment_file(tipo=movimento)`.
//! Implementação pendente; usar `MotionSessionEvent` como gatilho.

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum MotionRecordingState {
    Idle,
    Recording,
}
