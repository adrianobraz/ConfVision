use crate::api::CameraStreamHealthReport;
use crate::stream_policy::StreamHealthAction;
use tokio::sync::RwLock;

pub async fn enqueue_stream_action(
    queue: &RwLock<Vec<CameraStreamHealthReport>>,
    camera_id: i64,
    action: StreamHealthAction,
    last_error: Option<String>,
    error_class: Option<String>,
) {
    let report = match action {
        StreamHealthAction::ReportStreamOk => CameraStreamHealthReport {
            camera_id,
            event: "stream_ok".into(),
            failures_consecutive: Some(0),
            hourly_attempts: Some(0),
            last_error: None,
            error_class: None,
            pause_reason: None,
        },
        StreamHealthAction::PauseAnalytic { reason } => CameraStreamHealthReport {
            camera_id,
            event: "pause_analytic".into(),
            failures_consecutive: None,
            hourly_attempts: None,
            last_error,
            error_class,
            pause_reason: Some(reason),
        },
        StreamHealthAction::RetryAfter(_) | StreamHealthAction::ReportFailure { .. } => {
            return;
        }
    };
    queue.write().await.push(report);
}

pub async fn enqueue_stream_failure(
    queue: &RwLock<Vec<CameraStreamHealthReport>>,
    camera_id: i64,
    failures: u32,
    hourly: u32,
    last_error: Option<String>,
    error_class: Option<String>,
) {
    queue.write().await.push(CameraStreamHealthReport {
        camera_id,
        event: "stream_failure".into(),
        failures_consecutive: Some(failures),
        hourly_attempts: Some(hourly),
        last_error,
        error_class,
        pause_reason: None,
    });
}

/// Falha transient (ex. FU-A): diagnóstico no Postgres sem alterar contadores de path 404.
pub async fn enqueue_stream_incident(
    queue: &RwLock<Vec<CameraStreamHealthReport>>,
    camera_id: i64,
    last_error: Option<String>,
    error_class: Option<String>,
) {
    queue.write().await.push(CameraStreamHealthReport {
        camera_id,
        event: "stream_incident".into(),
        failures_consecutive: None,
        hourly_attempts: None,
        last_error,
        error_class,
        pause_reason: None,
    });
}
