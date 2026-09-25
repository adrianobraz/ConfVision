//! Decode H.264 via libavcodec (CPU).

use std::sync::Arc;
use std::sync::Once;

use ffmpeg_next as ffmpeg;
use ffmpeg_next::util::error::{Error as FfmpegError, EAGAIN};

use super::super::error::DecodeError;
use super::super::luma::downscale_y_plane_into;
use super::super::types::{DecodedFrame, PixelFormat};
use super::super::types::{DECODED_LUMA_HEIGHT, DECODED_LUMA_WIDTH};
use super::VideoDecodeBackend;

static FFMPEG_INIT: Once = Once::new();

pub(crate) fn ensure_ffmpeg_init() {
    FFMPEG_INIT.call_once(|| {
        let _ = ffmpeg::init();
    });
}

pub(crate) fn set_extradata(
    context: &mut ffmpeg::codec::Context,
    data: &[u8],
) -> Result<(), DecodeError> {
    unsafe {
        use ffmpeg_next::ffi as sys;

        let ctx = context.as_mut_ptr();
        if !(*ctx).extradata.is_null() {
            sys::av_freep(std::ptr::addr_of_mut!((*ctx).extradata).cast::<std::ffi::c_void>());
        }
        let padding = sys::AV_INPUT_BUFFER_PADDING_SIZE as usize;
        let ed = sys::av_malloc(data.len() + padding) as *mut u8;
        if ed.is_null() {
            return Err(DecodeError::Ffmpeg("av_malloc extradata falhou".into()));
        }
        std::ptr::copy_nonoverlapping(data.as_ptr(), ed, data.len());
        std::ptr::write_bytes(ed.add(data.len()), 0, padding);
        (*ctx).extradata = ed;
        (*ctx).extradata_size = data.len() as i32;
    }
    Ok(())
}

pub struct CpuFfmpegDecoder {
    decoder: ffmpeg::decoder::Video,
    extradata_generation: u64,
    receive_frame: ffmpeg::util::frame::Video,
    luma_scratch: Vec<u8>,
}

impl CpuFfmpegDecoder {
    pub fn try_new(extradata: &[u8], generation: u64) -> Result<Self, DecodeError> {
        ensure_ffmpeg_init();
        if extradata.is_empty() {
            return Err(DecodeError::NoExtradata);
        }

        let mut context = ffmpeg::codec::Context::new();
        unsafe {
            use ffmpeg_next::ffi as sys;
            (*context.as_mut_ptr()).codec_id = sys::AVCodecID::AV_CODEC_ID_H264;
            (*context.as_mut_ptr()).codec_type = sys::AVMediaType::AVMEDIA_TYPE_VIDEO;
        }
        set_extradata(&mut context, extradata)?;

        let decoder = context
            .decoder()
            .video()
            .map_err(|e| DecodeError::Ffmpeg(e.to_string()))?;

        Ok(Self {
            decoder,
            extradata_generation: generation,
            receive_frame: ffmpeg::util::frame::Video::empty(),
            luma_scratch: Vec::with_capacity(
                DECODED_LUMA_WIDTH as usize * DECODED_LUMA_HEIGHT as usize,
            ),
        })
    }
}

impl VideoDecodeBackend for CpuFfmpegDecoder {
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

        let packet = ffmpeg::Packet::copy(data);
        self.decoder.send_packet(&packet).map_err(map_ffmpeg_err)?;

        self.decoder
            .receive_frame(&mut self.receive_frame)
            .map_err(map_ffmpeg_err)?;

        decoded_from_video_frame(&self.receive_frame, &mut self.luma_scratch)
    }
}

pub(crate) fn decoded_from_video_frame(
    frame: &ffmpeg::util::frame::Video,
    luma_scratch: &mut Vec<u8>,
) -> Result<DecodedFrame, DecodeError> {
    let format = match frame.format() {
        ffmpeg::format::Pixel::YUV420P | ffmpeg::format::Pixel::YUVJ420P => PixelFormat::Yuv420p,
        ffmpeg::format::Pixel::NV12 => PixelFormat::Nv12,
        other => {
            return Err(DecodeError::UnsupportedPixelFormat(format!("{other:?}")));
        }
    };

    let src_w = frame.width();
    let src_h = frame.height();
    let y_stride = frame.stride(0);
    let y_data = frame.data(0);
    downscale_y_plane_into(y_data, src_w, src_h, y_stride, luma_scratch)?;
    let luma = Arc::from(std::mem::take(luma_scratch).into_boxed_slice());
    *luma_scratch = Vec::with_capacity(luma.len());

    Ok(DecodedFrame {
        width: src_w,
        height: src_h,
        format,
        luma,
        luma_width: super::super::types::DECODED_LUMA_WIDTH,
        luma_height: super::super::types::DECODED_LUMA_HEIGHT,
    })
}

pub fn map_ffmpeg_err(err: FfmpegError) -> DecodeError {
    if err == (FfmpegError::Other { errno: EAGAIN }) {
        DecodeError::Ffmpeg("Resource temporarily unavailable".into())
    } else {
        DecodeError::Ffmpeg(err.to_string())
    }
}

pub fn is_eagain(err: &DecodeError) -> bool {
    matches!(
        err,
        DecodeError::Ffmpeg(msg)
            if msg.contains("Resource temporarily unavailable") || msg.contains("EAGAIN")
    )
}
