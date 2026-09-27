use serde::{Deserialize, Serialize};
use serde_json::{Map, Value};
use tracing::debug;

use super::ConfVisionClient;
use crate::error::{AppError, AppResult};

#[derive(Debug, Serialize)]
struct CreateEventoBody {
    vis_camera_id: i64,
    #[serde(skip_serializing_if = "Option::is_none")]
    id_franqueado: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    id_cliente: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    id_dispositivo: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    conta: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    particao: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    canal: Option<Value>,
    tipo_deteccao: String,
    confianca: f64,
    processado: bool,
    ignorado: bool,
    status: String,
    clip_count: i32,
}

#[derive(Debug, Deserialize)]
pub struct EventoResponse {
    pub id: Option<i64>,
    #[serde(default)]
    pub dados: Option<Map<String, Value>>,
}

impl ConfVisionClient {
    pub async fn create_evento(
        &self,
        camera: &Value,
        confianca: f64,
        status: &str,
    ) -> AppResult<EventoResponse> {
        let camera_id = camera
            .get("id")
            .and_then(|v| v.as_i64())
            .ok_or_else(|| AppError::Api("camera sem id".into()))?;
        let body = CreateEventoBody {
            vis_camera_id: camera_id,
            id_franqueado: camera.get("id_franqueado").cloned(),
            id_cliente: camera.get("id_cliente").cloned(),
            id_dispositivo: camera.get("id_dispositivo").cloned(),
            conta: camera.get("conta").cloned(),
            particao: camera.get("particao").cloned(),
            canal: camera.get("canal").cloned(),
            tipo_deteccao: "humano".into(),
            confianca,
            processado: false,
            ignorado: false,
            status: status.into(),
            clip_count: 0,
        };
        let url = format!("{}/vis_evento", self.base);
        debug!(camera_id, "create_evento");
        let resp = self
            .http
            .post(&url)
            .headers(self.headers.clone())
            .json(&body)
            .send()
            .await
            .map_err(|e| AppError::Api(e.to_string()))?;
        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            return Err(AppError::Api(format!("vis_evento HTTP {status}: {text}")));
        }
        let raw: Value = resp
            .json()
            .await
            .map_err(|e| AppError::Api(e.to_string()))?;
        parse_evento_response(&raw)
    }

    pub async fn finalizar_evento(
        &self,
        evento_id: i64,
        snapshot_url: Option<&str>,
        video_url: Option<&str>,
        clip_duracao_seg: u32,
    ) -> AppResult<()> {
        let url = format!("{}/vis_evento_finalizar", self.base);
        let body = serde_json::json!({
            "vis_evento_id": evento_id,
            "snapshot_url": snapshot_url,
            "video_url": video_url,
            "status": "pronto",
            "clip_count": if video_url.is_some() { 1 } else { 0 },
            "processado": false,
            "clip_seq": 1,
            "clip_duracao_seg": clip_duracao_seg,
            "clip_snapshot_url": snapshot_url,
        });
        let resp = self
            .http
            .post(&url)
            .headers(self.headers.clone())
            .json(&body)
            .send()
            .await
            .map_err(|e| AppError::Api(e.to_string()))?;
        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            return Err(AppError::Api(format!(
                "vis_evento_finalizar HTTP {status}: {text}"
            )));
        }
        Ok(())
    }
}

pub fn evento_id_from_response(resp: &EventoResponse) -> Option<i64> {
    if let Some(id) = resp.id {
        return Some(id);
    }
    resp.dados
        .as_ref()
        .and_then(|d| d.get("id"))
        .and_then(|v| v.as_i64())
}

fn parse_evento_response(raw: &Value) -> AppResult<EventoResponse> {
    if let Ok(r) = serde_json::from_value::<EventoResponse>(raw.clone()) {
        if r.id.is_some() || r.dados.is_some() {
            return Ok(r);
        }
    }
    if let Some(id) = raw.get("id").and_then(|v| v.as_i64()) {
        return Ok(EventoResponse {
            id: Some(id),
            dados: None,
        });
    }
    Err(AppError::Api(format!("resposta vis_evento inválida: {raw}")))
}
