use std::path::PathBuf;

use serde_json::Value;
use tracing::{info, warn};

use crate::api::{evento_id_from_response, ConfVisionClient};
use crate::camera::resolve_rtsp_url_from_value;
use crate::config::Config;
use crate::error::AppResult;
use crate::events::EventJob;
use crate::media::{evento_clip_key, evento_snapshot_key, MediaKind, MediaStorage};

use super::ffmpeg::{capture_clip_rtsp, capture_snapshot_rtsp, clip_timeout};
use super::locks::CaptureGuard;

fn plan_grava_foto(camera: &Value) -> bool {
    if let Some(v) = camera.get("evento_grava_foto").and_then(|x| x.as_bool()) {
        return v;
    }
    camera
        .get("captura_analitico")
        .and_then(|x| x.as_bool())
        .unwrap_or(false)
        || camera
            .get("captura_sensor")
            .and_then(|x| x.as_bool())
            .unwrap_or(false)
}

fn plan_grava_video(camera: &Value) -> bool {
    if let Some(v) = camera.get("evento_grava_video").and_then(|x| x.as_bool()) {
        return v;
    }
    plan_grava_foto(camera)
}

pub async fn processar_deteccao(
    cfg: &Config,
    client: &ConfVisionClient,
    job: &EventJob,
    _guard: CaptureGuard,
) {
    let camera_val = Value::Object(job.camera.clone());
    let camera = &camera_val;
    let camera_id = camera.get("id").and_then(|v| v.as_i64()).unwrap_or(0);
    let conf = job.confianca;
    let grava_foto = plan_grava_foto(camera);
    let grava_video = plan_grava_video(camera);

    if !grava_foto && !grava_video {
        match client.create_evento(camera, conf, "capturando").await {
            Ok(resp) => {
                if let Some(eid) = evento_id_from_response(&resp) {
                    let _ = client
                        .finalizar_evento(eid, None, None, cfg.clip_duracao_seg)
                        .await;
                    info!(camera_id, evento_id = eid, "evento somente registro");
                }
            }
            Err(e) => warn!(camera_id, error = %e, "create_evento falhou"),
        }
        discard_snapshot(job.snapshot_path.as_deref());
        return;
    }

    let rtsp = match resolve_rtsp_url_from_value(camera, cfg) {
        Ok(u) => u,
        Err(e) => {
            warn!(camera_id, error = %e, "rtsp para captura indisponível");
            discard_snapshot(job.snapshot_path.as_deref());
            return;
        }
    };

    let evento = match client.create_evento(camera, conf, "capturando").await {
        Ok(r) => r,
        Err(e) => {
            warn!(camera_id, error = %e, "create_evento");
            discard_snapshot(job.snapshot_path.as_deref());
            return;
        }
    };
    let evento_id = match evento_id_from_response(&evento) {
        Some(id) => id,
        None => {
            warn!(camera_id, "sem evento_id na resposta");
            discard_snapshot(job.snapshot_path.as_deref());
            return;
        }
    };

    info!(
        camera_id,
        evento_id,
        detected_at = job.detected_at,
        component = "capture",
        operation = "create_evento",
        "evento vis_evento criado"
    );

    let work_dir = cfg.capture_dir.join(evento_id.to_string());
    let snapshot_path = work_dir.join("snapshot.jpg");
    let clip_path = work_dir.join("clip_001.mp4");

    if let Err(e) = std::fs::create_dir_all(&work_dir) {
        warn!(camera_id, error = %e, "mkdir work_dir");
        discard_snapshot(job.snapshot_path.as_deref());
        return;
    }

    if grava_foto {
        if let Some(src) = job.snapshot_path.as_ref() {
            if let Err(e) = install_snapshot(src, &snapshot_path) {
                warn!(camera_id, error = %e, "install detection snapshot");
            }
        } else if grava_video {
            let _ = capture_snapshot_rtsp(&rtsp, &snapshot_path).await;
        } else {
            let _ = capture_snapshot_rtsp(&rtsp, &snapshot_path).await;
        }
    }

    let mut video_url: Option<String> = None;
    if grava_video {
        match tokio::time::timeout(
            clip_timeout(cfg.clip_duracao_seg),
            capture_clip_rtsp(&rtsp, &clip_path, cfg.clip_duracao_seg),
        )
        .await
        {
            Ok(Ok(())) => {}
            Ok(Err(e)) => warn!(camera_id, error = %e, "clip capture"),
            Err(_) => warn!(camera_id, "clip timeout"),
        }
    }

    let id_franqueado = camera
        .get("id_franqueado")
        .and_then(|v| v.as_str().map(String::from))
        .or_else(|| {
            camera
                .get("id_franqueado")
                .and_then(|v| v.as_i64())
                .map(|n| n.to_string())
        });

    let storage = MediaStorage::from_config(cfg);
    let mut snapshot_url: Option<String> = None;
    if grava_foto {
        let key = evento_snapshot_key(id_franqueado.as_deref(), evento_id);
        snapshot_url = storage
            .put_or_log(
                MediaKind::EventSnapshot,
                camera_id,
                &snapshot_path,
                &key,
                "image/jpeg",
            )
            .await;
    }

    if grava_video {
        let key = evento_clip_key(id_franqueado.as_deref(), evento_id, 1);
        if let Some(url) = storage
            .put_or_log(
                MediaKind::EventClip,
                camera_id,
                &clip_path,
                &key,
                "video/mp4",
            )
            .await
        {
            video_url = Some(url);
        }
    }

    if let Err(e) = client
        .finalizar_evento(
            evento_id,
            snapshot_url.as_deref(),
            video_url.as_deref(),
            cfg.clip_duracao_seg,
        )
        .await
    {
        warn!(camera_id, evento_id, error = %e, "finalizar_evento");
    } else {
        info!(
            camera_id,
            evento_id,
            snapshot = snapshot_url.is_some(),
            video = video_url.is_some(),
            component = "capture",
            operation = "finalizar",
            "captura concluída"
        );
    }

    let _ = std::fs::remove_dir_all(&work_dir);
    discard_snapshot(job.snapshot_path.as_deref());
}

fn install_snapshot(src: &str, dest: &PathBuf) -> AppResult<()> {
    std::fs::copy(src, dest).map_err(|e| crate::error::AppError::Other(e.into()))?;
    Ok(())
}

fn discard_snapshot(path: Option<&str>) {
    if let Some(p) = path {
        let _ = std::fs::remove_file(p);
    }
}
