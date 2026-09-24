//! Backend FFmpeg — uso interno da crate `decode` apenas.

use std::sync::Once;

use ffmpeg_next as ffmpeg;
use ffmpeg_next::util::error::{Error as FfmpegError, EAGAIN};

use super::error::DecodeError;
use super::types::{DecodedFrame, PixelFormat};

static FFMPEG_INIT: Once = Once::new();

fn ensure_ffmpeg_init() {
    FFMPEG_INIT.call_once(|| {
        let _ = ffmpeg::init();
    });
}

fn set_extradata(context: &mut ffmpeg::codec::Context, data: &[u8]) -> Result<(), DecodeError> {
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

pub(crate) struct FfmpegH264Decoder {
    decoder: ffmpeg::decoder::Video,
    extradata_generation: u64,
}

impl FfmpegH264Decoder {
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
        })
    }

    pub fn needs_reinit(&self, generation: u64) -> bool {
        self.extradata_generation != generation
    }

    pub fn decode_access_unit(
        &mut self,
        data: &[u8],
        _is_keyframe: bool,
    ) -> Result<DecodedFrame, DecodeError> {
        if data.is_empty() {
            return Err(DecodeError::Ffmpeg("payload vazio".into()));
        }

        let packet = ffmpeg::Packet::copy(data);
        self.decoder.send_packet(&packet).map_err(map_ffmpeg_err)?;

        let mut frame = ffmpeg::util::frame::Video::empty();
        self.decoder
            .receive_frame(&mut frame)
            .map_err(map_ffmpeg_err)?;

        let format = match frame.format() {
            ffmpeg::format::Pixel::YUV420P | ffmpeg::format::Pixel::YUVJ420P => {
                PixelFormat::Yuv420p
            }
            ffmpeg::format::Pixel::NV12 => PixelFormat::Nv12,
            other => {
                return Err(DecodeError::UnsupportedPixelFormat(format!("{other:?}")));
            }
        };

        Ok(DecodedFrame {
            width: frame.width(),
            height: frame.height(),
            format,
        })
    }
}

fn map_ffmpeg_err(err: FfmpegError) -> DecodeError {
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
