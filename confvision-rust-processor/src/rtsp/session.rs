use std::sync::Arc;
use std::time::{Duration, Instant};

use futures::StreamExt;
use retina::client::{
    Credentials, Demuxed, PlayOptions, Session, SessionOptions, SetupOptions, Transport,
};
use retina::codec::{CodecItem, ParametersRef, VideoFrame};
use tracing::debug;

use crate::camera::{
    flush_pending_camera_state, record_frame_enqueued, record_frame_received,
    record_rtsp_au_throttled, CameraCancel, LiveCaptureContext,
};
use crate::decode::SessionDecodeContext;
use crate::error::{AppError, AppResult};
use crate::motion::MotionEnqueueGate;
use crate::pipeline::{FramePipeline, PipelineFrame, RtpTimestamp};

fn sync_h264_extradata(session: &Demuxed, video_index: usize, decode_ctx: &SessionDecodeContext) {
    let Some(ParametersRef::Video(params)) = session.streams()[video_index].parameters() else {
        return;
    };
    decode_ctx.update_extradata(params.extra_data());
}

fn pipeline_frame_from_video(
    seq: u64,
    frame: VideoFrame,
    is_keyframe: bool,
    decoder_reset: bool,
) -> PipelineFrame {
    let ts = frame.timestamp();
    let rtp_timestamp = RtpTimestamp {
        timestamp: ts.timestamp(),
        clock_rate_hz: ts.clock_rate().get(),
        stream_start: ts.start(),
    };
    // Move do buffer interno do retina — sem cópia extra dos bytes do AU.
    let payload: Arc<[u8]> = Arc::from(frame.into_data());
    PipelineFrame::with_decoder_reset(seq, payload, is_keyframe, rtp_timestamp, decoder_reset)
}

// Conecta ao RTSP e conta frames de vídeo até cancelamento ou erro.
pub async fn run_rtsp_frame_loop(
    camera_id: i64,
    rtsp_url: &str,
    connect_timeout: Duration,
    frame_timeout: Duration,
    cancel: CameraCancel,
    mut live: Option<&mut LiveCaptureContext<'_>>,
    pipeline: &FramePipeline,
    decode_ctx: &SessionDecodeContext,
    enqueue_gate: &mut MotionEnqueueGate,
) -> AppResult<RtspLoopStats> {
    let url = rtsp_url
        .parse()
        .map_err(|e| AppError::Rtsp(format!("URL inválida: {e}")))?;

    let session_options = SessionOptions::default().creds(None::<Credentials>);

    let mut session =
        tokio::time::timeout(connect_timeout, Session::describe(url, session_options))
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
    sync_h264_extradata(&session, video_index, decode_ctx);

    let mut last_frame = Instant::now();
    let mut seq: u64 = 0;
    loop {
        if cancel.is_cancelled() {
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
            Ok(Some(Ok(CodecItem::VideoFrame(frame)))) => {
                if frame.has_new_parameters() {
                    sync_h264_extradata(&session, video_index, decode_ctx);
                }
                seq += 1;
                let is_keyframe = frame.is_random_access_point();
                let decision = enqueue_gate.decide(is_keyframe);
                if decision.enqueue {
                    if let Some(ctx) = live.as_mut() {
                        record_frame_enqueued(ctx);
                    }
                    pipeline.try_enqueue(pipeline_frame_from_video(
                        seq,
                        frame,
                        is_keyframe,
                        decision.decoder_reset,
                    ));
                } else {
                    if let Some(ctx) = live.as_mut() {
                        record_rtsp_au_throttled(ctx);
                    }
                }
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

    if let Some(ctx) = live.as_mut() {
        flush_pending_camera_state(ctx);
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

fn simulated_pipeline_frame(seq: u64) -> PipelineFrame {
    let payload = Arc::from(format!("sim-h264-{seq}").into_bytes());
    PipelineFrame::new(
        seq,
        payload,
        seq == 1 || seq % 30 == 0,
        RtpTimestamp {
            timestamp: seq as i64 * 3_000,
            clock_rate_hz: 90_000,
            stream_start: 0,
        },
    )
}

// Simula recebimento de frames (testes / RTSP_SIMULATE=1).
pub async fn simulate_frame_loop(
    camera_id: i64,
    fps: f64,
    cancel: CameraCancel,
    mut live: Option<&mut LiveCaptureContext<'_>>,
    pipeline: &FramePipeline,
) -> RtspLoopStats {
    let interval = Duration::from_secs_f64(1.0 / fps.max(0.1));
    let mut seq = 0u64;
    while !cancel.is_cancelled() {
        tokio::time::sleep(interval).await;
        if cancel.is_cancelled() {
            break;
        }
        seq += 1;
        if let Some(ctx) = live.as_mut() {
            record_frame_received(ctx);
        }
        pipeline.try_enqueue(simulated_pipeline_frame(seq));
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
    use std::sync::Arc;

    use crate::camera::{CameraCancel, CameraRuntimeState, SharedCameraState};
    use crate::metrics::ProcessorMetrics;

    #[tokio::test]
    async fn simulate_respects_cancel() {
        let metrics = Arc::new(ProcessorMetrics::new("test"));
        let state: SharedCameraState = Arc::new(tokio::sync::RwLock::new(CameraRuntimeState::new(
            1,
            "x".into(),
        )));
        let pipeline = FramePipeline::new(2, metrics, state);
        let (gtx, grx) = tokio::sync::watch::channel(false);
        let (ltx, lrx) = tokio::sync::watch::channel(false);
        let cancel = CameraCancel::new(grx, lrx);
        let handle =
            tokio::spawn(
                async move { simulate_frame_loop(1, 50.0, cancel, None, &pipeline).await },
            );
        tokio::time::sleep(Duration::from_millis(80)).await;
        ltx.send(true).unwrap();
        let _ = gtx;
        let stats = handle.await.unwrap();
        assert!(stats.frames_received > 0);
    }
}
