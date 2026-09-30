//! Decode H.264 via FFmpeg NVDEC (h264_cuvid + CUDA hwdevice).

use std::time::Instant;

use ffmpeg_next as ffmpeg;
use ffmpeg_next::util::error::{Error as FfmpegError, EAGAIN};

use super::super::error::DecodeError;
use super::super::types::DecodedFrame;
use super::cpu::{decoded_from_video_frame, ensure_ffmpeg_init, map_ffmpeg_err, set_extradata};
use super::VideoDecodeBackend;

pub struct NvdecCapability;

impl NvdecCapability {
    /// Valida criação real de dispositivo CUDA (não basta `/dev/nvidia0`).
    pub fn probe_cuda_device() -> bool {
        ensure_ffmpeg_init();
        unsafe {
            use ffmpeg_next::ffi as sys;

            let mut hw_device_ctx = std::ptr::null_mut();
            let ret = sys::av_hwdevice_ctx_create(
                &mut hw_device_ctx,
                sys::AVHWDeviceType::AV_HWDEVICE_TYPE_CUDA,
                std::ptr::null(),
                std::ptr::null_mut(),
                0,
            );
            if ret >= 0 && !hw_device_ctx.is_null() {
                sys::av_buffer_unref(&mut hw_device_ctx);
                return true;
            }
            false
        }
    }

    pub fn h264_cuvid_decoder_registered() -> bool {
        ensure_ffmpeg_init();
        ffmpeg::decoder::find_by_name("h264_cuvid").is_some()
    }
}

pub struct NvdecH264Decoder {
    decoder: ffmpeg::decoder::Video,
    hw_device_ctx: *mut ffmpeg_next::ffi::AVBufferRef,
    extradata_generation: u64,
    pub(crate) last_transfer_to_cpu_ms: u64,
}

impl Drop for NvdecH264Decoder {
    fn drop(&mut self) {
        unsafe {
            use ffmpeg_next::ffi as sys;
            if !self.hw_device_ctx.is_null() {
                sys::av_buffer_unref(&mut self.hw_device_ctx);
            }
        }
    }
}

/// Libera `AVBufferRef` de dispositivo CUDA se `try_new` falhar antes de entregar o decoder.
struct HwDeviceCtxGuard(*mut ffmpeg_next::ffi::AVBufferRef);

impl HwDeviceCtxGuard {
    fn disarm(&mut self) -> *mut ffmpeg_next::ffi::AVBufferRef {
        let ptr = self.0;
        self.0 = std::ptr::null_mut();
        ptr
    }
}

impl Drop for HwDeviceCtxGuard {
    fn drop(&mut self) {
        unsafe {
            use ffmpeg_next::ffi as sys;
            if !self.0.is_null() {
                sys::av_buffer_unref(&mut self.0);
            }
        }
    }
}

impl NvdecH264Decoder {
    pub fn try_new(extradata: &[u8], generation: u64) -> Result<Self, DecodeError> {
        ensure_ffmpeg_init();
        if extradata.is_empty() {
            return Err(DecodeError::NoExtradata);
        }
        if !NvdecCapability::h264_cuvid_decoder_registered() {
            return Err(DecodeError::Ffmpeg(
                "decoder h264_cuvid não disponível no FFmpeg".into(),
            ));
        }

        unsafe {
            use ffmpeg_next::ffi as sys;

            let mut hw_device_ctx = std::ptr::null_mut();
            let ret = sys::av_hwdevice_ctx_create(
                &mut hw_device_ctx,
                sys::AVHWDeviceType::AV_HWDEVICE_TYPE_CUDA,
                std::ptr::null(),
                std::ptr::null_mut(),
                0,
            );
            if ret < 0 || hw_device_ctx.is_null() {
                return Err(DecodeError::Ffmpeg(format!(
                    "av_hwdevice_ctx_create(CUDA) falhou: {ret}"
                )));
            }

            let mut device_guard = HwDeviceCtxGuard(hw_device_ctx);

            let codec = match ffmpeg::decoder::find_by_name("h264_cuvid") {
                Some(c) => c,
                None => {
                    return Err(DecodeError::Ffmpeg("h264_cuvid não encontrado".into()));
                }
            };

            let mut context = ffmpeg::codec::Context::new();
            (*context.as_mut_ptr()).hw_device_ctx = sys::av_buffer_ref(device_guard.0);
            if (*context.as_mut_ptr()).hw_device_ctx.is_null() {
                return Err(DecodeError::Ffmpeg(
                    "av_buffer_ref hw_device_ctx falhou".into(),
                ));
            }

            if let Err(e) = set_extradata(&mut context, extradata) {
                return Err(e);
            }

            let opened = match context.decoder().open_as(codec) {
                Ok(o) => o,
                Err(e) => return Err(DecodeError::Ffmpeg(e.to_string())),
            };

            let decoder = match opened.video() {
                Ok(d) => d,
                Err(e) => return Err(DecodeError::Ffmpeg(e.to_string())),
            };

            Ok(Self {
                decoder,
                hw_device_ctx: device_guard.disarm(),
                extradata_generation: generation,
                last_transfer_to_cpu_ms: 0,
            })
        }
    }
}

impl VideoDecodeBackend for NvdecH264Decoder {
    fn needs_reinit(&self, extradata_generation: u64) -> bool {
        self.extradata_generation != extradata_generation
    }

    fn decode_access_unit(
        &mut self,
        data: &[u8],
        _is_keyframe: bool,
    ) -> Result<DecodedFrame, DecodeError> {
        if data.is_empty() {
            return Err(DecodeError::Ffmpeg("payload vazio".into()));
        }

        self.last_transfer_to_cpu_ms = 0;

        let packet = ffmpeg::Packet::copy(data);
        self.decoder.send_packet(&packet).map_err(map_ffmpeg_err)?;

        let mut frame = ffmpeg::util::frame::Video::empty();
        self.decoder
            .receive_frame(&mut frame)
            .map_err(map_ffmpeg_err)?;

        if frame_is_hardware(&frame) {
            let mut sw_frame = ffmpeg::util::frame::Video::empty();
            let started = Instant::now();
            unsafe {
                use ffmpeg_next::ffi as sys;
                let ret = sys::av_hwframe_transfer_data(sw_frame.as_mut_ptr(), frame.as_ptr(), 0);
                if ret < 0 {
                    return Err(DecodeError::Ffmpeg(format!(
                        "av_hwframe_transfer_data falhou: {ret}"
                    )));
                }
            }
            self.last_transfer_to_cpu_ms = started.elapsed().as_millis() as u64;
            let mut luma_scratch = Vec::new();
            decoded_from_video_frame(&sw_frame, &mut luma_scratch)
        } else {
            let mut luma_scratch = Vec::new();
            decoded_from_video_frame(&frame, &mut luma_scratch)
        }
    }
}

fn frame_is_hardware(frame: &ffmpeg::util::frame::Video) -> bool {
    unsafe {
        use ffmpeg_next::ffi as sys;
        let fmt = (*frame.as_ptr()).format;
        sys::av_pix_fmt_desc_get(fmt)
            .map(|desc| (*desc).flags & sys::AV_PIX_FMT_FLAG_HWACCEL as i64 != 0)
            .unwrap_or(false)
    }
}

pub fn is_eagain(err: &DecodeError) -> bool {
    super::cpu::is_eagain(err)
}

#[allow(dead_code)]
pub fn map_nvdec_ffmpeg_err(err: FfmpegError) -> DecodeError {
    map_ffmpeg_err(err)
}
