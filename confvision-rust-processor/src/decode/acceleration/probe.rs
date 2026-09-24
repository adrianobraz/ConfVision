use std::path::Path;

use super::policy::VideoGpuBackend;

/// Resultado do probe — separa presença de dispositivo vs capacidade FFmpeg HW.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ProbeResult {
    pub gpu_detected: bool,
    pub ffmpeg_hw_decode_available: bool,
    /// Backend inferido para métricas (`none`, `cuda`, `vaapi`, `qsv`).
    pub inferred_hw_backend: &'static str,
}

/// Abstração para testes sem GPU física.
pub trait SystemProbe: Send + Sync {
    fn nvidia_device_present(&self) -> bool;
    fn vaapi_device_present(&self) -> bool;
    fn intel_qsv_likely(&self) -> bool;
    fn ffmpeg_reports_hw_decode(&self) -> bool;
}

#[derive(Debug, Default)]
pub struct DefaultSystemProbe;

impl SystemProbe for DefaultSystemProbe {
    fn nvidia_device_present(&self) -> bool {
        nvidia_device_present()
    }

    fn vaapi_device_present(&self) -> bool {
        vaapi_device_present()
    }

    fn intel_qsv_likely(&self) -> bool {
        intel_qsv_likely()
    }

    fn ffmpeg_reports_hw_decode(&self) -> bool {
        ffmpeg_reports_hw_decode()
    }
}

pub fn run_probe(probe: &dyn SystemProbe, requested_backend: VideoGpuBackend) -> ProbeResult {
    let nvidia = probe.nvidia_device_present();
    let vaapi_dev = probe.vaapi_device_present();
    let qsv = probe.intel_qsv_likely();
    let ffmpeg_hw = probe.ffmpeg_reports_hw_decode();

    let gpu_detected = nvidia || vaapi_dev || qsv;

    let inferred = resolve_inferred_backend(requested_backend, nvidia, vaapi_dev, qsv);

    ProbeResult {
        gpu_detected,
        ffmpeg_hw_decode_available: ffmpeg_hw,
        inferred_hw_backend: inferred,
    }
}

fn resolve_inferred_backend(
    requested: VideoGpuBackend,
    nvidia: bool,
    vaapi: bool,
    qsv: bool,
) -> &'static str {
    match requested {
        VideoGpuBackend::Cuda => {
            if nvidia {
                "cuda"
            } else {
                "none"
            }
        }
        VideoGpuBackend::Vaapi => {
            if vaapi {
                "vaapi"
            } else {
                "none"
            }
        }
        VideoGpuBackend::Qsv => {
            if qsv {
                "qsv"
            } else {
                "none"
            }
        }
        VideoGpuBackend::Auto => {
            if nvidia {
                "cuda"
            } else if qsv {
                "qsv"
            } else if vaapi {
                "vaapi"
            } else {
                "none"
            }
        }
    }
}

fn nvidia_device_present() -> bool {
    Path::new("/dev/nvidia0").exists()
}

fn vaapi_device_present() -> bool {
    Path::new("/dev/dri/renderD128").exists() || Path::new("/dev/dri/renderD129").exists()
}

fn intel_qsv_likely() -> bool {
    Path::new("/dev/dri/renderD128").exists()
        && (Path::new("/sys/class/drm/card0/device/vendor").exists()
            || Path::new("/dev/dri/card0").exists())
}

fn ffmpeg_reports_hw_decode() -> bool {
    probe_ffmpeg_hw_via_configuration()
}

#[cfg(feature = "ffmpeg-decode")]
fn probe_ffmpeg_hw_via_configuration() -> bool {
    unsafe {
        use ffmpeg_next::ffi as sys;
        let cfg = sys::avcodec_configuration();
        if cfg.is_null() {
            return false;
        }
        let text = std::ffi::CStr::from_ptr(cfg)
            .to_string_lossy()
            .to_lowercase();
        text.contains("cuda")
            || text.contains("nvdec")
            || text.contains("cuvid")
            || text.contains("vaapi")
            || text.contains("qsv")
    }
}

#[cfg(not(feature = "ffmpeg-decode"))]
fn probe_ffmpeg_hw_via_configuration() -> bool {
    false
}

#[cfg(test)]
pub struct MockProbe {
    pub nvidia: bool,
    pub vaapi: bool,
    pub qsv: bool,
    pub ffmpeg_hw: bool,
}

#[cfg(test)]
impl SystemProbe for MockProbe {
    fn nvidia_device_present(&self) -> bool {
        self.nvidia
    }

    fn vaapi_device_present(&self) -> bool {
        self.vaapi
    }

    fn intel_qsv_likely(&self) -> bool {
        self.qsv
    }

    fn ffmpeg_reports_hw_decode(&self) -> bool {
        self.ffmpeg_hw
    }
}
