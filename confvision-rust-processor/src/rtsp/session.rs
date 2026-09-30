use std::sync::Arc;
use std::time::{Duration, Instant};

use futures::StreamExt;
use retina::client::{
    Credentials, Demuxed, PlayOptions, Session, SessionOptions, SetupOptions, Transport,
    UnassignedChannelDataPolicy,
};
use retina::codec::{CodecItem, FrameFormat, ParametersRef, VideoFrame};
use tracing::debug;

use crate::camera::{
    flush_pending_camera_state, record_frame_enqueued, record_frame_received,
    record_rtsp_au_throttled, CameraCancel, LiveCaptureContext,
};
use crate::decode::SessionDecodeContext;
use crate::error::{AppError, AppResult};
use crate::motion::MotionEnqueueGate;
use crate::pipeline::{FramePipeline, PipelineFrame, RtpTimestamp};
use crate::rtsp_hotpath::RtspHotpathStats;

/// Checagens de cancel/timeout a cada N AUs (retina já entrega ~15 FPS/câmera).
const SESSION_HEALTH_CHECK_EVERY: u32 = 8;
/// Cancelamento durante drenagem de RTCP/outros itens não-vídeo.
const SESSION_CANCEL_CHECK_EVERY_NON_VIDEO: u32 = 32;

fn video_setup_options() -> SetupOptions {
    SetupOptions::default()
        .transport(Transport::Tcp(
            retina::client::TcpTransportOptions::default(),
        ))
        // Default do depacketizer retina = ParameterSetInsertion::EachKeyFrame (copia SPS/PPS
        // em todo IDR). Decode usa extradata out-of-band — alinhar ao MP4 evita trabalho extra.
        .frame_format(FrameFormat::MP4)
}

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

fn handle_video_frame(
    frame: VideoFrame,
    seq: u64,
    session: &Demuxed,
    video_index: usize,
    decode_ctx: &SessionDecodeContext,
    enqueue_gate: &mut MotionEnqueueGate,
    live: &mut Option<&mut LiveCaptureContext<'_>>,
    pipeline: &FramePipeline,
    hotpath: &RtspHotpathStats,
) {
    let au_start = Instant::now();
    if frame.has_new_parameters() {
        sync_h264_extradata(session, video_index, decode_ctx);
    }
    let throttle_start = Instant::now();
    let is_keyframe = frame.is_random_access_point();
    let decision = enqueue_gate.decide(is_keyframe);
    let throttle_nanos = throttle_start.elapsed().as_nanos() as u64;

    let metrics_start = Instant::now();
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
    } else if let Some(ctx) = live.as_mut() {
        record_rtsp_au_throttled(ctx);
    }
    let metrics_nanos = metrics_start.elapsed().as_nanos() as u64;
    let post_au_nanos = au_start.elapsed().as_nanos() as u64;
    hotpath.record_video_au(post_au_nanos, throttle_nanos, metrics_nanos);
}

/// DESCRIBE → SETUP → PLAY → demux. Falhas (ex.: 404) retornam antes do loop de frames.
pub async fn connect_rtsp_demuxed(
    rtsp_url: &str,
    connect_timeout: Duration,
    decode_ctx: &SessionDecodeContext,
) -> AppResult<(Demuxed, usize)> {
    let url = rtsp_url
        .parse()
        .map_err(|e| AppError::Rtsp(format!("URL inválida: {e}")))?;

    let session_options = SessionOptions::default()
        .creds(None::<Credentials>)
        .unassigned_channel_data(UnassignedChannelDataPolicy::Ignore);

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
        .setup(video_index, video_setup_options())
        .await
        .map_err(|e| AppError::Rtsp(e.to_string()))?;

    let session = session
        .play(PlayOptions::default())
        .await
        .map_err(|e| AppError::Rtsp(e.to_string()))?
        .demuxed()
        .map_err(|e| AppError::Rtsp(e.to_string()))?;

    sync_h264_extradata(&session, video_index, decode_ctx);
    Ok((session, video_index))
}

// Lê frames de uma sessão RTSP já conectada.
#[derive(Debug, Clone, Copy)]
pub enum RtspDemuxOutcome {
    Finished(RtspLoopStats),
    /// Cena parada: pausa RTSP até próximo probe (gate motion).
    IdleSuspend(RtspLoopStats),
}

pub async fn run_rtsp_demux_loop(
    camera_id: i64,
    mut session: Demuxed,
    video_index: usize,
    frame_timeout: Duration,
    cancel: CameraCancel,
    mut live: Option<&mut LiveCaptureContext<'_>>,
    pipeline: &FramePipeline,
    decode_ctx: &SessionDecodeContext,
    enqueue_gate: &mut MotionEnqueueGate,
    hotpath: &RtspHotpathStats,
) -> AppResult<RtspDemuxOutcome> {
    debug!(camera_id, "rtsp connected");
    let mut last_frame = Instant::now();
    let mut seq: u64 = 0;
    let mut health_tick: u32 = 0;

    loop {
        let loop_start = Instant::now();
        health_tick = health_tick.wrapping_add(1);
        if health_tick % SESSION_HEALTH_CHECK_EVERY == 0 {
            if cancel.is_cancelled() {
                hotpath.record_loop_iter(loop_start.elapsed().as_nanos() as u64);
                break;
            }
            if enqueue_gate.should_suspend_rtsp() {
                if let Some(ctx) = live.as_mut() {
                    flush_pending_camera_state(ctx);
                }
                hotpath.record_loop_iter(loop_start.elapsed().as_nanos() as u64);
                return Ok(RtspDemuxOutcome::IdleSuspend(RtspLoopStats {
                    frames_received: seq,
                    frames_dropped: 0,
                }));
            }
            if last_frame.elapsed() > frame_timeout {
                hotpath.record_loop_iter(loop_start.elapsed().as_nanos() as u64);
                return Err(AppError::Rtsp(format!(
                    "frame timeout após {}s",
                    frame_timeout.as_secs()
                )));
            }
        }

        tokio::select! {
            biased;
            () = cancel.wait_until_cancelled(), if !cancel.is_cancelled() => {
                if cancel.is_cancelled() {
                    hotpath.record_loop_iter(loop_start.elapsed().as_nanos() as u64);
                    break;
                }
            }
            item = async {
                let mut non_video_drained = 0u32;
                loop {
                    let session_next_start = Instant::now();
                    let item = session.next().await;
                    hotpath.record_session_next(session_next_start.elapsed().as_nanos() as u64);
                    match item {
                        Some(Ok(CodecItem::VideoFrame(_))) | Some(Err(_)) | None => {
                            return item;
                        }
                        Some(Ok(_)) => {
                            hotpath.record_session_next_non_video();
                            non_video_drained = non_video_drained.wrapping_add(1);
                            if non_video_drained % SESSION_CANCEL_CHECK_EVERY_NON_VIDEO == 0 {
                                if cancel.is_cancelled() {
                                    return None;
                                }
                                tokio::task::yield_now().await;
                            }
                        }
                    }
                }
            } => {
                match item {
                    Some(Ok(CodecItem::VideoFrame(frame))) => {
                        seq += 1;
                        handle_video_frame(
                            frame,
                            seq,
                            &session,
                            video_index,
                            decode_ctx,
                            enqueue_gate,
                            &mut live,
                            pipeline,
                            hotpath,
                        );
                        last_frame = Instant::now();
                        if seq == 1 || seq % 100 == 0 {
                            debug!(camera_id, seq, "frame received");
                        }
                    }
                    Some(Ok(_)) => {}
                    Some(Err(e)) => {
                        hotpath.record_loop_iter(loop_start.elapsed().as_nanos() as u64);
                        return Err(AppError::Rtsp(e.to_string()));
                    }
                    None => {
                        if cancel.is_cancelled() {
                            hotpath.record_loop_iter(loop_start.elapsed().as_nanos() as u64);
                            break;
                        }
                        hotpath.record_loop_iter(loop_start.elapsed().as_nanos() as u64);
                        return Err(AppError::Rtsp("stream RTSP encerrado".into()));
                    }
                }
            }
        }
        hotpath.record_loop_iter(loop_start.elapsed().as_nanos() as u64);
    }

    if let Some(ctx) = live.as_mut() {
        flush_pending_camera_state(ctx);
    }

    Ok(RtspDemuxOutcome::Finished(RtspLoopStats {
        frames_received: seq,
        frames_dropped: 0,
    }))
}

// Conecta ao RTSP e conta frames de vídeo até cancelamento ou erro.
pub async fn run_rtsp_frame_loop(
    camera_id: i64,
    rtsp_url: &str,
    connect_timeout: Duration,
    frame_timeout: Duration,
    cancel: CameraCancel,
    live: Option<&mut LiveCaptureContext<'_>>,
    pipeline: &FramePipeline,
    decode_ctx: &SessionDecodeContext,
    enqueue_gate: &mut MotionEnqueueGate,
    hotpath: &RtspHotpathStats,
) -> AppResult<RtspLoopStats> {
    let (session, video_index) =
        connect_rtsp_demuxed(rtsp_url, connect_timeout, decode_ctx).await?;
    match run_rtsp_demux_loop(
        camera_id,
        session,
        video_index,
        frame_timeout,
        cancel,
        live,
        pipeline,
        decode_ctx,
        enqueue_gate,
        hotpath,
    )
    .await?
    {
        RtspDemuxOutcome::Finished(s) | RtspDemuxOutcome::IdleSuspend(s) => Ok(s),
    }
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
