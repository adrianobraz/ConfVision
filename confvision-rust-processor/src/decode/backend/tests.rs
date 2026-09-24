#[cfg(test)]
mod selection {
    use crate::decode::acceleration::probe::MockProbe;
    use crate::decode::acceleration::{
        AccelerationPolicy, AccelerationRuntime, PlannedDecodeBackend, VideoAccelerationMode,
        VideoGpuBackend,
    };
    use crate::decode::backend::DecodeBackendKind;

    fn mock(nvidia: bool, ffmpeg_hw: bool, cuvid: bool, cuda_init: bool) -> MockProbe {
        MockProbe {
            nvidia,
            vaapi: false,
            qsv: false,
            ffmpeg_hw,
            h264_cuvid: cuvid,
            nvdec_cuda_init_ok: cuda_init,
        }
    }

    #[test]
    fn policy_auto_without_gpu_selects_cpu_backend() {
        let rt = AccelerationRuntime::bootstrap_with_probe(
            AccelerationPolicy::from_parts(VideoAccelerationMode::Auto, VideoGpuBackend::Auto),
            &mock(false, false, false, false),
        )
        .unwrap();
        assert_eq!(rt.planned_decode_backend(), PlannedDecodeBackend::Cpu);
        assert_eq!(DecodeBackendKind::Cpu.metric_label(), "cpu");
    }

    #[test]
    fn policy_cpu_selects_cpu_backend() {
        let rt = AccelerationRuntime::bootstrap_with_probe(
            AccelerationPolicy::from_parts(VideoAccelerationMode::Cpu, VideoGpuBackend::Cuda),
            &mock(true, true, true, true),
        )
        .unwrap();
        assert_eq!(rt.planned_decode_backend(), PlannedDecodeBackend::Cpu);
    }

    #[test]
    fn policy_gpu_without_stack_is_config_error() {
        let err = AccelerationRuntime::bootstrap_with_probe(
            AccelerationPolicy::from_parts(VideoAccelerationMode::Gpu, VideoGpuBackend::Auto),
            &mock(true, true, true, false),
        );
        assert!(err.is_err());
    }

    #[cfg(all(feature = "ffmpeg-decode", feature = "ffmpeg-nvdec"))]
    #[test]
    fn policy_auto_full_stack_selects_nvdec_kind() {
        let rt = AccelerationRuntime::bootstrap_with_probe(
            AccelerationPolicy::from_parts(VideoAccelerationMode::Auto, VideoGpuBackend::Auto),
            &mock(true, true, true, true),
        )
        .unwrap();
        assert_eq!(rt.planned_decode_backend(), PlannedDecodeBackend::Nvdec);
        assert_eq!(DecodeBackendKind::Nvdec.metric_label(), "nvdec");
    }
}
