use thiserror::Error;

#[derive(Error, Debug)]
pub enum AppError {
    #[error("configuração: {0}")]
    Config(String),
    #[error("API ConfVision: {0}")]
    Api(String),
    #[error("RTSP: {0}")]
    Rtsp(String),
    #[error("YOLO ocupado: {0}")]
    YoloBusy(String),
    #[error("{0}")]
    Other(#[from] anyhow::Error),
}

impl AppError {
    pub fn is_yolo_busy(&self) -> bool {
        matches!(self, AppError::YoloBusy(_))
    }
}

pub type AppResult<T> = Result<T, AppError>;
