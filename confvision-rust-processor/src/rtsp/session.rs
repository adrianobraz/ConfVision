use std::time::{Duration, Instant};

use futures::StreamExt;
use retina::client::{Credentials, PlayOptions, Session, SessionOptions, SetupOptions, Transport};
use retina::codec::CodecItem;
use tracing::debug;

use crate::camera::{LiveCaptureContext, record_frame_received};
use crate::error::{AppError, AppResult};
use crate::pipeline::FramePipeline;

/// Conecta ao RTSP e conta frames de vídeo até cancelamento ou erro.
pub async fn run_rtsp_frame_loop(
    camera_id: i64,
    rtsp_url: &str,
    connect_timeout: Duration,
    frame_timeout: Duration,
    mut cancel: tokio::sync::watch::Receiver<bool>,
    mut live: Option<&mut LiveCaptureContext<'_>>,
    pipeline: &FramePipeline,
) -> AppResult<RtspLoopStats> {
    let url = rtsp_url
        .parse()
        .map_err(|e| AppError::Rtsp(format!("URL inválida: {e}")))?;

    let session_options = SessionOptions::default().creds(None::<Credentials>);

    let mut session = tokio::time::timeout(connect_timeout, Session::describe(url, session_options))
        .await
        .map_err(|_| AppError::Rtsp("timeout ao conectar RTSP".into()))?
        .map_err(|e| AppError::Rtsp(e.to_string()))?;

    let video_index = session
        .streams()
        .iter()
        .position(|s| s.media() == "video")
        .ok_or_else(|| AppError::Rtsp("nenhum stream de video no RTSP".into()))?;

    session
        .setup(
            video_index,
            SetupOptions::default().transport(Transport::Tcp(
                retina::client::TcpTransportOptions::default(),
            )),
        )
        .await
        .map_err(|e| AppError::Rtsp(e.to_string()))?;

    let mut session = session
        .play(PlayOptions::default())
        .await
        .map_err(|e| AppError::Rtsp(e.to_string()))?
        .demuxed()
        .map_err(|e| AppError::Rtsp(e.to_string()))?;

    debug!(camera_id, "rtsp connected");

    let mut last_frame = Instant::now();
    let mut seq: u64 = 0;

    loop {
        if *cancel.borrow() {
            break;
        }

        if last_frame.elapsed() > frame_timeout {
            return Err(AppError::Rtsp(format!(
                "frame timeout após {}s",
                frame_timeout.as_secs()
            )));
        }

        let item = tokio::time::timeout(Duration::from_secs(2), session.next()).await;

        match item {
            Ok(Some(Ok(CodecItem::VideoFrame(_frame)))) => {
                seq += 1;
                if let Some(ctx) = live.as_mut() {
                    record_frame_received(ctx).await;
                }
                pipeline.try_enqueue(seq);
                last_frame = Instant::now();
                if seq == 1 || seq % 100 == 0 {
                    debug!(camera_id, seq, "frame received");
                }
            }
            Ok(Some(Ok(_))) => {}
            Ok(Some(Err(e))) => {
                return Err(AppError::Rtsp(e.to_string()));
            }
            Ok(None) => {
                return Err(AppError::Rtsp("stream RTSP encerrado".into()));
            }
            Err(_) => continue,
        }
    }

    Ok(RtspLoopStats {
        frames_received: seq,
        frames_dropped: 0,
    })
}

#[derive(Debug, Clone, Copy)]
pub struct RtspLoopStats {
    pub frames_received: u64,
    pub frames_dropped: u64,
}

/// Simula recebimento de frames (testes / RTSP_SIMULATE=1).
pub async fn simulate_frame_loop(
    camera_id: i64,
    fps: f64,
    mut cancel: tokio::sync::watch::Receiver<bool>,
    mut live: Option<&mut LiveCaptureContext<'_>>,
    pipeline: &FramePipeline,
) -> RtspLoopStats {
    let interval = Duration::from_secs_f64(1.0 / fps.max(0.1));
    let mut seq = 0u64;
    while !*cancel.borrow() {
        tokio::time::sleep(interval).await;
        if *cancel.borrow() {
            break;
        }
        seq += 1;
        if let Some(ctx) = live.as_mut() {
            record_frame_received(ctx).await;
        }
        pipeline.try_enqueue(seq);
        if seq % 30 == 0 {
            debug!(camera_id, seq, "simulated frame");
        }
    }
    RtspLoopStats {
        frames_received: seq,
        frames_dropped: 0,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashMap;
    use std::sync::Arc;

    use crate::metrics::ProcessorMetrics;

    #[tokio::test]
    async fn simulate_respects_cancel() {
        let metrics = Arc::new(ProcessorMetrics::new("test"));
        let states = Arc::new(tokio::sync::RwLock::new(HashMap::new()));
        let pipeline = FramePipeline::new(2, metrics, states, 1);
        let (tx, rx) = tokio::sync::watch::channel(false);
        let handle = tokio::spawn(async move {
            simulate_frame_loop(1, 50.0, rx, None, &pipeline).await
        });
        tokio::time::sleep(Duration::from_millis(80)).await;
        tx.send(true).unwrap();
        let stats = handle.await.unwrap();
        assert!(stats.frames_received > 0);
    }
}
