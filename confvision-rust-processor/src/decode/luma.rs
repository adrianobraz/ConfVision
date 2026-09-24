use std::sync::Arc;

use super::error::DecodeError;
use super::types::{DECODED_LUMA_HEIGHT, DECODED_LUMA_WIDTH};

/// Reduz o plano Y (com stride) para grade fixa 160×120 (média por célula).
pub fn downscale_y_plane(
    data: &[u8],
    src_width: u32,
    src_height: u32,
    stride: usize,
) -> Result<Arc<[u8]>, DecodeError> {
    if src_width == 0 || src_height == 0 {
        return Err(DecodeError::Ffmpeg("dimensão de vídeo inválida".into()));
    }
    if stride < src_width as usize {
        return Err(DecodeError::Ffmpeg("stride Y menor que largura".into()));
    }

    let dst_w = DECODED_LUMA_WIDTH as usize;
    let dst_h = DECODED_LUMA_HEIGHT as usize;
    let mut out = vec![0u8; dst_w * dst_h];
    let sw = src_width as usize;
    let sh = src_height as usize;

    for dy in 0..dst_h {
        let sy0 = dy * sh / dst_h;
        let sy1 = ((dy + 1) * sh / dst_h).max(sy0 + 1);
        for dx in 0..dst_w {
            let sx0 = dx * sw / dst_w;
            let sx1 = ((dx + 1) * sw / dst_w).max(sx0 + 1);
            let mut sum = 0u64;
            let mut count = 0u64;
            for sy in sy0..sy1 {
                let row_off = sy * stride;
                for sx in sx0..sx1 {
                    if sx < sw && row_off + sx < data.len() {
                        sum += data[row_off + sx] as u64;
                        count += 1;
                    }
                }
            }
            out[dy * dst_w + dx] = if count > 0 { (sum / count) as u8 } else { 0 };
        }
    }

    Ok(Arc::from(out.into_boxed_slice()))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn downscale_uniform_plane() {
        let w = 640usize;
        let h = 480usize;
        let stride = w;
        let data = vec![42u8; stride * h];
        let arc = downscale_y_plane(&data, w as u32, h as u32, stride).unwrap();
        assert_eq!(arc.len(), 160 * 120);
        assert!(arc.iter().all(|&b| b == 42));
    }
}
