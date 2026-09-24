//! Fixture H.264 mínimo para testes de decode (feature `ffmpeg-decode`).
//!
//! **Origem:** avcC Baseline **16×16** (quadro preto), típico de:
//!
//! ```text
//! ffmpeg -f lavfi -i color=c=black:s=16x16:r=1 -frames:v 1 \
//!   -c:v libx264 -profile:v baseline -pix_fmt yuv420p \
//!   -x264-params repeat-headers=1:keyint=1:min-keyint=1 -f h264 pipe:1
//! ```
//!
//! `EXTRADATA` = registro **avcC** (SPS/PPS). Access units reais do RTSP usam NALs
//! length-prefixed (4 bytes BE), como produz o `retina`.

/// avcC Baseline 16×16 (~39 bytes).
pub const EXTRADATA: &[u8] = &[
    0x01, 0x42, 0x00, 0x0a, 0xff, 0xe1, 0x00, 0x0b, 0x67, 0x42, 0x00, 0x0a, 0xda, 0x14, 0x0a,
    0x66, 0x11, 0x00, 0x00, 0x03, 0x00, 0x01, 0x00, 0x00, 0x03, 0x00, 0x64, 0x1e, 0x2c, 0x5b,
    0x20, 0x01, 0x00, 0x04, 0x68, 0xee, 0x3c, 0xb0,
];
