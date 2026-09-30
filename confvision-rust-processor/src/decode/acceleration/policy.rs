use crate::error::{AppError, AppResult};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum VideoAccelerationMode {
    Auto,
    Cpu,
    Gpu,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum VideoGpuBackend {
    Auto,
    Cuda,
    Vaapi,
    Qsv,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct AccelerationPolicy {
    pub mode: VideoAccelerationMode,
    pub gpu_backend: VideoGpuBackend,
}

impl AccelerationPolicy {
    pub fn from_env() -> AppResult<Self> {
        let mode = parse_video_acceleration(&env_or("VIDEO_ACCELERATION", "auto"))?;
        let gpu_backend = parse_video_gpu_backend(&env_or("VIDEO_GPU_BACKEND", "auto"))?;
        Ok(Self { mode, gpu_backend })
    }

    pub fn from_parts(mode: VideoAccelerationMode, gpu_backend: VideoGpuBackend) -> Self {
        Self { mode, gpu_backend }
    }
}

pub fn parse_video_acceleration(raw: &str) -> AppResult<VideoAccelerationMode> {
    match raw.trim().to_lowercase().as_str() {
        "auto" => Ok(VideoAccelerationMode::Auto),
        "cpu" => Ok(VideoAccelerationMode::Cpu),
        "gpu" => Ok(VideoAccelerationMode::Gpu),
        other => Err(AppError::Config(format!(
            "VIDEO_ACCELERATION inválido: {other:?} (use auto, cpu ou gpu)"
        ))),
    }
}

pub fn parse_video_gpu_backend(raw: &str) -> AppResult<VideoGpuBackend> {
    match raw.trim().to_lowercase().as_str() {
        "auto" => Ok(VideoGpuBackend::Auto),
        "cuda" => Ok(VideoGpuBackend::Cuda),
        "vaapi" => Ok(VideoGpuBackend::Vaapi),
        "qsv" => Ok(VideoGpuBackend::Qsv),
        other => Err(AppError::Config(format!(
            "VIDEO_GPU_BACKEND inválido: {other:?} (use auto, cuda, vaapi ou qsv)"
        ))),
    }
}

pub fn mode_label(mode: VideoAccelerationMode) -> &'static str {
    match mode {
        VideoAccelerationMode::Auto => "auto",
        VideoAccelerationMode::Cpu => "cpu",
        VideoAccelerationMode::Gpu => "gpu",
    }
}

pub fn backend_label(backend: VideoGpuBackend) -> &'static str {
    match backend {
        VideoGpuBackend::Auto => "auto",
        VideoGpuBackend::Cuda => "cuda",
        VideoGpuBackend::Vaapi => "vaapi",
        VideoGpuBackend::Qsv => "qsv",
    }
}

fn env_or(key: &str, default: &str) -> String {
    std::env::var(key).unwrap_or_else(|_| default.to_string())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn default_auto_when_env_missing() {
        std::env::remove_var("VIDEO_ACCELERATION");
        std::env::remove_var("VIDEO_GPU_BACKEND");
        let p = AccelerationPolicy::from_env().unwrap();
        assert_eq!(p.mode, VideoAccelerationMode::Auto);
        assert_eq!(p.gpu_backend, VideoGpuBackend::Auto);
    }

    #[test]
    fn cpu_mode() {
        let mode = parse_video_acceleration("cpu").unwrap();
        assert_eq!(mode, VideoAccelerationMode::Cpu);
    }

    #[test]
    fn auto_mode() {
        let mode = parse_video_acceleration("auto").unwrap();
        assert_eq!(mode, VideoAccelerationMode::Auto);
    }

    #[test]
    fn gpu_mode_parses() {
        let mode = parse_video_acceleration("gpu").unwrap();
        assert_eq!(mode, VideoAccelerationMode::Gpu);
    }

    #[test]
    fn gpu_backend_auto() {
        let b = parse_video_gpu_backend("auto").unwrap();
        assert_eq!(b, VideoGpuBackend::Auto);
    }

    #[test]
    fn invalid_acceleration_rejected() {
        assert!(parse_video_acceleration("nvidia").is_err());
    }

    #[test]
    fn invalid_gpu_backend_rejected() {
        assert!(parse_video_gpu_backend("opencl").is_err());
    }
}
