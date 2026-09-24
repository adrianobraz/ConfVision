use thiserror::Error;

#[derive(Error, Debug)]
pub enum AppError {
    #[error("configuração: {0}")]
    Config(String),
    #[error("API ConfVision: {0}")]
    Api(String),
    #[error("RTSP: {0}")]
    Rtsp(String),
    #[error("{0}")]
    Other(#[from] anyhow::Error),
}

pub type AppResult<T> = Result<T, AppError>;
