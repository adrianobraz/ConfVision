use std::path::{Path, PathBuf};

use image::{codecs::jpeg::JpegEncoder, ExtendedColorType, ImageBuffer, Luma};
use tracing::warn;

use crate::buffer;
use crate::decode::{DECODED_LUMA_HEIGHT, DECODED_LUMA_WIDTH};
use crate::error::{AppError, AppResult};

pub fn write_luma_jpeg(
    luma: &[u8],
    width: u32,
    height: u32,
    quality: u8,
    dest: &Path,
) -> AppResult<PathBuf> {
    if luma.len() != (width as usize) * (height as usize) {
        return Err(AppError::Config("luma buffer size mismatch".into()));
    }
    if let Some(parent) = dest.parent() {
        std::fs::create_dir_all(parent).map_err(|e| AppError::Other(e.into()))?;
    }
    let img: ImageBuffer<Luma<u8>, Vec<u8>> =
        ImageBuffer::from_raw(width, height, luma.to_vec())
            .ok_or_else(|| AppError::Config("failed to build luma image".into()))?;
    let file = std::fs::File::create(dest).map_err(|e| AppError::Other(e.into()))?;
    let mut enc = JpegEncoder::new_with_quality(file, quality);
    enc.encode(img.as_raw(), width, height, ExtendedColorType::L8)
        .map_err(|e| AppError::Other(e.into()))?;
    Ok(dest.to_path_buf())
}

pub fn luma_to_jpeg_bytes(luma: &[u8], width: u32, height: u32, quality: u8) -> AppResult<Vec<u8>> {
    let plane_len = (width as usize) * (height as usize);
    let mut raw = buffer::acquire(plane_len);
    if raw.len() < plane_len {
        raw.resize(plane_len, 0);
    }
    raw[..plane_len].copy_from_slice(luma);
    let img: ImageBuffer<Luma<u8>, Vec<u8>> =
        ImageBuffer::from_raw(width, height, raw)
            .ok_or_else(|| AppError::Config("failed to build luma image".into()))?;
    let mut buf = buffer::acquire(plane_len / 2);
    let mut enc = JpegEncoder::new_with_quality(&mut buf, quality);
    enc.encode(img.as_raw(), width, height, ExtendedColorType::L8)
        .map_err(|e| AppError::Other(e.into()))?;
    buffer::release(img.into_raw(), plane_len);
    Ok(buf)
}

pub fn write_detection_snapshot(
    camera_id: i64,
    luma: &[u8],
    capture_dir: &Path,
    quality: u8,
) -> AppResult<PathBuf> {
    let detect_dir = capture_dir.join("detect");
    let name = format!(
        "cam{camera_id}_{}.jpg",
        chrono::Utc::now().timestamp_millis()
    );
    let path = detect_dir.join(name);
    write_luma_jpeg(
        luma,
        DECODED_LUMA_WIDTH,
        DECODED_LUMA_HEIGHT,
        quality,
        &path,
    )
    .map_err(|e| {
        warn!(camera_id, error = %e, "snapshot write failed");
        e
    })
}
