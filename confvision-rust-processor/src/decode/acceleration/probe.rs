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

    /// Inicialização real CUDA via FFmpeg (separado de `/dev/nvidia0`).
    fn nvdec_cuda_init_ok(&self) -> bool;

    /// Decoder `h264_cuvid` registrado no libavcodec.
    fn ffmpeg_h264_cuvid_decoder_available(&self) -> bool;
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

    fn nvdec_cuda_init_ok(&self) -> bool {
        nvdec_cuda_init_ok_impl()
    }

    fn ffmpeg_h264_cuvid_decoder_available(&self) -> bool {
        ffmpeg_h264_cuvid_decoder_available_impl()
    }
}

pub fn run_probe(probe: &dyn SystemProbe, requested_backend: VideoGpuBackend) -> ProbeResult {
    let nvidia = probe.nvidia_device_present();
    let vaapi_dev = probe.vaapi_device_present();
    let qsv = probe.intel_qsv_likely();
    let ffmpeg_build_hw = probe.ffmpeg_reports_hw_decode();
    let cuvid = probe.ffmpeg_h264_cuvid_decoder_available();

    let gpu_detected = nvidia || vaapi_dev || qsv;

    let inferred = resolve_inferred_backend(requested_backend, nvidia, vaapi_dev, qsv);

    ProbeResult {
        gpu_detected,
        ffmpeg_hw_decode_available: ffmpeg_build_hw && cuvid,
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

#[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
fn nvdec_cuda_init_ok_impl() -> bool {
    if !nvidia_device_present() {
        return false;
    }
    crate::decode::backend::NvdecCapability::probe_cuda_device()
}

#[cfg(not(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec")))]
fn nvdec_cuda_init_ok_impl() -> bool {
    false
}

#[cfg(feature = "ffmpeg-decode")]
fn ffmpeg_h264_cuvid_decoder_available_impl() -> bool {
    #[cfg(feature = "ffmpeg-nvdec")]
    {
        crate::decode::backend::NvdecCapability::h264_cuvid_decoder_registered()
    }
    #[cfg(not(feature = "ffmpeg-nvdec"))]
    {
        ffmpeg_next::decoder::find_by_name("h264_cuvid").is_some()
    }
}

#[cfg(not(feature = "ffmpeg-decode"))]
fn ffmpeg_h264_cuvid_decoder_available_impl() -> bool {
    false
}

#[cfg(test)]
pub struct MockProbe {
    pub nvidia: bool,
    pub vaapi: bool,
    pub qsv: bool,
    pub ffmpeg_hw: bool,
    pub h264_cuvid: bool,
    pub nvdec_cuda_init_ok: bool,
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

    fn nvdec_cuda_init_ok(&self) -> bool {
        self.nvdec_cuda_init_ok
    }

    fn ffmpeg_h264_cuvid_decoder_available(&self) -> bool {
        self.h264_cuvid
    }
}

#[cfg(test)]
mod probe_tests {
    use super::*;
    use crate::decode::acceleration::policy::VideoGpuBackend;

    #[test]
    fn ffmpeg_hw_metric_requires_cuvid_decoder() {
        let probe = MockProbe {
            nvidia: true,
            vaapi: false,
            qsv: false,
            ffmpeg_hw: true,
            h264_cuvid: false,
            nvdec_cuda_init_ok: true,
        };
        let result = run_probe(&probe, VideoGpuBackend::Auto);
        assert!(result.gpu_detected);
        assert!(!result.ffmpeg_hw_decode_available);
    }
}
