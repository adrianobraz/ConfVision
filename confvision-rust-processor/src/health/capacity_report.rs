use axum::extract::State;
use axum::Json;
use chrono::Utc;
use serde::Serialize;

use crate::camera::{CameraRuntimeState, CameraStatus};
use crate::capacity::{CapacitySnapshot, CapacityState};
use crate::load::{allow_new_camera, evaluate_load, LoadAdvisory, LoadPolicyConfig, LoadPolicyMode};

use super::runtime_phase62::build_phase62_view;
use super::{snapshot_camera_states, summarize_cameras, AppState, RuntimeIdentity};

#[derive(Serialize)]
pub struct CapacityReportBody {
    pub generated_at: String,
    pub identity: RuntimeIdentity,
    pub summary: ReportSummary,
    pub capacity: CapacitySnapshot,
    pub load: LoadReportSection,
    pub runtime: RuntimeReportSection,
    pub cameras: Vec<CameraReportRow>,
    pub recommended_actions: Vec<RecommendedAction>,
    pub docs: ReportDocLinks,
}

#[derive(Serialize)]
pub struct ReportSummary {
    pub cameras_total: usize,
    pub cameras_online: usize,
    pub cameras_offline: usize,
    pub cameras_reconnecting: usize,
    pub rtsp_404_count: usize,
    pub fps_total: f64,
}

#[derive(Serialize)]
pub struct LoadReportSection {
    pub load_admission_enabled: bool,
    pub load_policy_mode: &'static str,
    pub admission_active: bool,
    pub allow_new_camera: bool,
    pub advisory: LoadAdvisory,
    pub advisory_reason: String,
}

#[derive(Serialize)]
pub struct RuntimeReportSection {
    pub decode_backend_effective: String,
    pub load_advisory: LoadAdvisory,
    pub load_advisory_reason: String,
}

#[derive(Serialize)]
pub struct CameraReportRow {
    pub camera_id: i64,
    pub status: CameraStatus,
    pub fps: f64,
    pub issue: Option<&'static str>,
    pub last_error_excerpt: Option<String>,
    pub suggested_action: Option<&'static str>,
}

#[derive(Serialize, Clone)]
pub struct RecommendedAction {
    pub priority: &'static str,
    pub code: &'static str,
    pub title: String,
    pub detail: String,
}

#[derive(Serialize)]
pub struct ReportDocLinks {
    pub runbook_404_ativo: &'static str,
    pub load_admission: &'static str,
    pub sql_404_worker_python: &'static str,
}

pub async fn capacity_report_handler(State(st): State<AppState>) -> Json<CapacityReportBody> {
    let capacity = st.capacity.snapshot().await;
    let load_cfg = st.load_admission.config().clone();
    let allow_new = allow_new_camera(&capacity, &load_cfg);
    let load_eval = evaluate_load(&capacity, &load_cfg);
    let phase62 = build_phase62_view(
        &st.acceleration,
        &st.decode_policy,
        &capacity,
        &load_cfg,
    );
    let summary_cam = summarize_cameras(&st.camera_states).await;
    let cameras_raw = snapshot_camera_states(&st.camera_states).await;

    let camera_rows: Vec<CameraReportRow> = cameras_raw
        .iter()
        .map(|c| camera_row(c))
        .collect();
    let rtsp_404_count = camera_rows
        .iter()
        .filter(|r| r.issue == Some("rtsp_404"))
        .count();

    let summary = ReportSummary {
        cameras_total: summary_cam.total,
        cameras_online: summary_cam.online,
        cameras_offline: summary_cam.offline,
        cameras_reconnecting: summary_cam.reconnecting,
        rtsp_404_count,
        fps_total: summary_cam.fps_total,
    };

    let recommended_actions = build_recommended_actions(
        &capacity,
        &load_cfg,
        allow_new,
        &load_eval.advisory,
        rtsp_404_count,
        &camera_rows,
    );

    Json(CapacityReportBody {
        generated_at: Utc::now().to_rfc3339(),
        identity: st.identity.clone(),
        summary,
        capacity,
        load: LoadReportSection {
            load_admission_enabled: load_cfg.admission_enabled,
            load_policy_mode: policy_mode_label(load_cfg.mode),
            admission_active: load_cfg.admission_active(),
            allow_new_camera: allow_new,
            advisory: load_eval.advisory,
            advisory_reason: load_eval.reason,
        },
        runtime: RuntimeReportSection {
            decode_backend_effective: phase62.decode_backend_effective,
            load_advisory: phase62.load_advisory,
            load_advisory_reason: phase62.load_advisory_reason,
        },
        cameras: camera_rows,
        recommended_actions,
        docs: ReportDocLinks {
            runbook_404_ativo: "confvision-rust-processor/docs/RUNBOOK_CAMERAS_404_ATIVO.md",
            load_admission: "confvision-rust-processor/docs/LOAD_ADMISSION.md",
            sql_404_worker_python: "confvision-rust-processor/sql/cameras_404_worker_python.sql",
        },
    })
}

fn policy_mode_label(mode: LoadPolicyMode) -> &'static str {
    match mode {
        LoadPolicyMode::Advisory => "advisory",
        LoadPolicyMode::Admission => "admission",
        LoadPolicyMode::Disabled => "disabled",
    }
}

fn camera_row(c: &CameraRuntimeState) -> CameraReportRow {
    let issue = classify_camera_issue(c);
    let suggested_action = issue.map(suggested_action_for_issue);
    CameraReportRow {
        camera_id: c.camera_id,
        status: c.status,
        fps: c.fps,
        issue,
        last_error_excerpt: c.last_error.as_ref().map(|e| excerpt(e, 160)),
        suggested_action,
    }
}

fn excerpt(s: &str, max: usize) -> String {
    if s.len() <= max {
        s.to_string()
    } else {
        format!("{}…", &s[..max])
    }
}

pub fn classify_camera_issue(c: &CameraRuntimeState) -> Option<&'static str> {
    let err = c
        .last_error
        .as_deref()
        .unwrap_or("")
        .to_ascii_lowercase();
    if err.contains("404") || err.contains("not found") || err.contains("describe failed") {
        return Some("rtsp_404");
    }
    match c.status {
        CameraStatus::Offline | CameraStatus::Error if !err.is_empty() => Some("rtsp_or_runtime_error"),
        CameraStatus::Offline => Some("offline_no_stream"),
        CameraStatus::Reconnecting => Some("reconnecting"),
        _ => None,
    }
}

fn suggested_action_for_issue(issue: &str) -> &'static str {
    match issue {
        "rtsp_404" => "move_to_python_worker_or_fix_publish_path",
        "offline_no_stream" => "check_rtmp_publish_or_mark_inactive",
        "rtsp_or_runtime_error" => "inspect_last_error_then_runbook",
        "reconnecting" => "wait_or_inspect_rtsp",
        _ => "inspect_camera",
    }
}

pub fn build_recommended_actions(
    capacity: &CapacitySnapshot,
    load_cfg: &LoadPolicyConfig,
    allow_new: bool,
    advisory: &LoadAdvisory,
    rtsp_404_count: usize,
    camera_rows: &[CameraReportRow],
) -> Vec<RecommendedAction> {
    let mut out: Vec<RecommendedAction> = Vec::new();

    if rtsp_404_count > 0 {
        let ids: Vec<String> = camera_rows
            .iter()
            .filter(|r| r.issue == Some("rtsp_404"))
            .map(|r| r.camera_id.to_string())
            .collect();
        out.push(RecommendedAction {
            priority: "high",
            code: "rtsp_404_active_cameras",
            title: "Câmeras ativas com RTSP 404 no MediaMTX".into(),
            detail: format!(
                "{rtsp_404_count} câmera(s) com path RTSP inexistente (ids: {}). \
                 Mover worker_id para o Python ou corrigir publish (ver runbook; rtmp_guard normaliza barra final). \
                 SQL: sql/cameras_404_worker_python.sql",
                ids.join(", ")
            ),
        });
    }

    let no_headroom = capacity.estimated_available_cameras.unwrap_or(0) <= 0
        || matches!(capacity.state, CapacityState::Critical)
        || matches!(advisory, LoadAdvisory::Saturated | LoadAdvisory::RejectAdmission);

    if no_headroom {
        out.push(RecommendedAction {
            priority: "high",
            code: "split_second_processor",
            title: "Sem headroom — dividir câmeras em 2º processor".into(),
            detail: format!(
                "capacity_state={:?}, estimated_available_cameras={:?}, limiting_resource={:?}. \
                 Crie um 2º serviço EasyPanel (novo PROCESSOR_ID/WORKER_ID), ajuste worker_id no Postgres \
                 e mantenha MAX_CAMERAS como teto de segurança por instância. Não escale antes de liberar carga (404/offline).",
                capacity.state,
                capacity.estimated_available_cameras,
                capacity.limiting_resource,
            ),
        });
    }

    if matches!(capacity.state, CapacityState::Critical) && !load_cfg.admission_enabled {
        out.push(RecommendedAction {
            priority: "medium",
            code: "enable_load_admission",
            title: "Ativar LOAD_ADMISSION_ENABLED=1".into(),
            detail: "CPU/recursos em critical: com admission ON, novas câmeras no sync são rejeitadas; sessões existentes continuam. Ver docs/LOAD_ADMISSION.md.".into(),
        });
    }

    if load_cfg.admission_enabled && !allow_new {
        out.push(RecommendedAction {
            priority: "medium",
            code: "admission_blocking_new_cameras",
            title: "Admission bloqueando novas câmeras".into(),
            detail: "Novas câmeras não entram até capacity sair de critical. Reduza carga (404→Python/inativo) ou adicione 2º processor.".into(),
        });
    }

    if out.is_empty() {
        out.push(RecommendedAction {
            priority: "info",
            code: "ok",
            title: "Nenhuma ação urgente".into(),
            detail: "Monitorar /metrics e repetir este relatório após mudanças de worker_id ou perfil motion.".into(),
        });
    }

    out
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::capacity::{
        CapacitySnapshot, CapacityState, LimitingResource, NetworkSnapshot, ResourceMetric,
        ResourceScope, StorageSnapshot,
    };
    use crate::config::CapacityMode;

    fn empty_capacity(state: CapacityState) -> CapacitySnapshot {
        let m = ResourceMetric {
            percent: Some(90.0),
            target_percent: 80.0,
            headroom_percent: None,
            scope: ResourceScope::Process,
        };
        CapacitySnapshot {
            mode: CapacityMode::Dynamic,
            state,
            worker_id: "w".into(),
            processor_id: "p".into(),
            cpu: m.clone(),
            memory: m.clone(),
            gpu: m.clone(),
            vram: m,
            network: NetworkSnapshot {
                rx_bytes_total: None,
                tx_bytes_total: None,
                rx_bps: None,
                tx_bps: None,
                scope: ResourceScope::Unavailable,
            },
            storage: StorageSnapshot {
                disk_total_bytes: None,
                disk_used_bytes: None,
                disk_free_bytes: None,
                disk_used_percent: None,
                scope: ResourceScope::Unavailable,
            },
            current_cameras: 6,
            current_online_cameras: 6,
            current_fps: 1.0,
            average_fps_per_camera: None,
            current_frames_received: 0,
            current_frames_processed: 0,
            current_frames_dropped: 0,
            drop_rate_percent: None,
            queue_depth: 0,
            processing_latency_ms: 0,
            estimated_capacity_cameras: Some(4),
            estimated_capacity_fps: None,
            estimated_available_cameras: Some(0),
            estimated_available_fps: None,
            capacity_used_percent: Some(95.0),
            limiting_resource: LimitingResource::Cpu,
            max_cameras_safety_limit: Some(10),
            cameras: vec![],
            observation_ready: true,
            samples_in_window: 10,
        }
    }

    #[test]
    fn split_processor_when_no_headroom() {
        let cap = empty_capacity(CapacityState::Critical);
        let cfg = LoadPolicyConfig {
            mode: LoadPolicyMode::Advisory,
            admission_enabled: false,
        };
        let actions = build_recommended_actions(
            &cap,
            &cfg,
            true,
            &LoadAdvisory::Saturated,
            0,
            &[],
        );
        assert!(actions.iter().any(|a| a.code == "split_second_processor"));
    }

    #[test]
    fn classifies_rtsp_404() {
        let mut c = CameraRuntimeState::new(2, "rtsp://redacted".into());
        c.last_error = Some("DESCRIBE failed: 404 Not Found".into());
        assert_eq!(classify_camera_issue(&c), Some("rtsp_404"));
    }
}
