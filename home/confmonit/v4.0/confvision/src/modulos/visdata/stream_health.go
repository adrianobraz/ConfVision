package visdata

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

type CameraStreamHealthInput struct {
	CameraID            int64   `json:"camera_id"`
	Event               string  `json:"event"`
	FailuresConsecutive *int    `json:"failures_consecutive"`
	HourlyAttempts      *int    `json:"hourly_attempts"`
	LastError           *string `json:"last_error"`
	ErrorClass          *string `json:"error_class"`
	PauseReason         *string `json:"pause_reason"`
}

func ApplyCameraStreamHealthBatch(ctx context.Context, reports []CameraStreamHealthInput) error {
	if len(reports) == 0 {
		return nil
	}
	db, err := DB()
	if err != nil {
		return err
	}
	for _, r := range reports {
		if r.CameraID <= 0 {
			continue
		}
		ev := strings.TrimSpace(strings.ToLower(r.Event))
		switch ev {
		case "stream_ok":
			_, err = db.ExecContext(ctx, `
UPDATE vis_camera SET
  ultimo_stream_ok_em = NOW(),
  stream_falhas_consecutivas = 0,
  stream_tentativas_horarias = 0,
  stream_motivo_pausa = NULL,
  stream_ultimo_erro = NULL,
  stream_erro_classe = NULL,
  analitico_pausado = CASE
    WHEN stream_motivo_pausa LIKE 'sistema_stream%' THEN FALSE
    ELSE analitico_pausado
  END
WHERE id = $1`, r.CameraID)
		case "stream_failure":
			f := 0
			h := 0
			if r.FailuresConsecutive != nil {
				f = *r.FailuresConsecutive
			}
			if r.HourlyAttempts != nil {
				h = *r.HourlyAttempts
			}
			lastErr := sql.NullString{}
			if r.LastError != nil {
				lastErr = sql.NullString{String: truncateStreamText(*r.LastError, 500), Valid: true}
			}
			errClass := sql.NullString{}
			if r.ErrorClass != nil {
				errClass = sql.NullString{String: truncateStreamText(*r.ErrorClass, 64), Valid: true}
			}
			_, err = db.ExecContext(ctx, `
UPDATE vis_camera SET
  stream_falhas_consecutivas = $2,
  stream_tentativas_horarias = $3,
  stream_ultimo_erro = COALESCE($4, stream_ultimo_erro),
  stream_erro_classe = COALESCE($5, stream_erro_classe)
WHERE id = $1`, r.CameraID, f, h, lastErr, errClass)
		case "stream_incident":
			lastErr := sql.NullString{}
			if r.LastError != nil {
				lastErr = sql.NullString{String: truncateStreamText(*r.LastError, 500), Valid: true}
			}
			errClass := sql.NullString{}
			if r.ErrorClass != nil {
				errClass = sql.NullString{String: truncateStreamText(*r.ErrorClass, 64), Valid: true}
			}
			_, err = db.ExecContext(ctx, `
UPDATE vis_camera SET
  stream_ultimo_erro = COALESCE($2, stream_ultimo_erro),
  stream_erro_classe = COALESCE($3, stream_erro_classe)
WHERE id = $1`, r.CameraID, lastErr, errClass)
		case "pause_analytic":
			reason := "sistema_stream"
			if r.PauseReason != nil && strings.TrimSpace(*r.PauseReason) != "" {
				reason = truncateStreamText(*r.PauseReason, 120)
			}
			lastErr := sql.NullString{}
			if r.LastError != nil {
				lastErr = sql.NullString{String: truncateStreamText(*r.LastError, 500), Valid: true}
			}
			errClass := sql.NullString{}
			if r.ErrorClass != nil {
				errClass = sql.NullString{String: truncateStreamText(*r.ErrorClass, 64), Valid: true}
			}
			_, err = db.ExecContext(ctx, `
UPDATE vis_camera SET
  analitico_pausado = TRUE,
  stream_motivo_pausa = $2,
  stream_ultimo_erro = COALESCE($3, stream_ultimo_erro),
  stream_erro_classe = COALESCE($4, stream_erro_classe)
WHERE id = $1`, r.CameraID, reason, lastErr, errClass)
			if err != nil {
				return err
			}
			if err := notifyCameraStreamPaused(ctx, db, r.CameraID, reason, lastErr, errClass); err != nil {
				return err
			}
			continue
		default:
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func notifyCameraStreamPaused(
	ctx context.Context,
	db *sql.DB,
	cameraID int64,
	reason string,
	lastErr sql.NullString,
	errClass sql.NullString,
) error {
	if !strings.HasPrefix(reason, "sistema_stream") {
		return nil
	}

	var idFra, nome sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT id_franqueado, nome FROM vis_camera WHERE id = $1`, cameraID).
		Scan(&idFra, &nome); err != nil {
		return err
	}

	motivoCodigo := reason
	if motivoCodigo == "" {
		motivoCodigo = "sistema_stream"
	}
	titulo := fmt.Sprintf("Analítico pausado — câmera %s (#%d)", strings.TrimSpace(nome.String), cameraID)
	dica := "Verifique a publicação RTMP no MediaMTX (path cam/{hash}) e clique em Reativar stream após corrigir."
	if reason == "sistema_stream_rtsp_404" {
		titulo = fmt.Sprintf("Stream RTSP 404 — analítico desligado (câmera %s)", strings.TrimSpace(nome.String))
		dica = "O path RTSP não existe no MediaMTX. Corrija o encode/publicação RTMP e use Reativar stream."
	}
	if reason == "sistema_stream_h264_nal" {
		titulo = fmt.Sprintf("Vídeo H.264 inválido — analítico desligado (câmera %s)", strings.TrimSpace(nome.String))
		dica = "Stream corrompido (NAL). Ajuste codec/bitrate/RTMP na câmera ou DVR e clique em Reativar stream."
	}
	if reason == "sistema_stream_video_track_not_set_up" {
		titulo = fmt.Sprintf("Vídeo RTMP inconsistente — analítico pausado (câmera %s)", strings.TrimSpace(nome.String))
		dica = "O sistema detectou pacote de vídeo sem trilha configurada (video track not set up). " +
			"Corrija encoder/H.264+AAC no DVR ou câmera e use Reativar stream no painel."
	}

	detail := map[string]any{
		"stream_motivo_pausa": reason,
		"analitico_pausado":   true,
	}
	if lastErr.Valid {
		detail["stream_ultimo_erro"] = lastErr.String
	}
	if errClass.Valid {
		detail["stream_erro_classe"] = errClass.String
	}
	dj, _ := json.Marshal(detail)
	dedupe := fmt.Sprintf("vis_camera:%d:pause:%s", cameraID, motivoCodigo)

	var exists int
	_ = db.QueryRowContext(ctx, `
SELECT 1 FROM vis_stream_relatorio
WHERE referencia_dedupe = $1 AND created_at > NOW() - INTERVAL '30 minutes'
LIMIT 1`, dedupe).Scan(&exists)
	if exists != 1 {
		_, err := db.ExecContext(ctx, `
INSERT INTO vis_stream_relatorio (
  id_franqueado, vis_camera_id, fonte, motivo_codigo, titulo, dica, severidade, detalhe_json, referencia_dedupe
) VALUES ($1,$2,'vis_camera',$3,$4,$5,'error',$6,$7)`,
			nullStrVal(idFra), cameraID, motivoCodigo, titulo, dica, dj, dedupe)
		if err != nil {
			return err
		}
	}

	_ = db.QueryRowContext(ctx, `
SELECT 1 FROM vis_evento
WHERE vis_camera_id = $1 AND tipo_deteccao = 'sistema_stream'
  AND ignorado IS NOT TRUE AND created_at > NOW() - INTERVAL '24 hours'
LIMIT 1`, cameraID).Scan(&exists)
	if exists == 1 {
		return nil
	}

	_, err := db.ExecContext(ctx, `
INSERT INTO vis_evento (
  vis_camera_id, id_franqueado, tipo_deteccao, confianca, processado, ignorado, status, clip_count, id_evento, id_processo
) VALUES ($1,$2,'sistema_stream',1,TRUE,FALSE,'pronto',0,$3,$4)`,
		cameraID, nullStrVal(idFra), fmt.Sprintf("sistema_stream_pause:%d", cameraID), truncateStreamText(dica, 200))
	return err
}

func ReactivateCameraStream(ctx context.Context, cameraID int) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	res, err := db.ExecContext(ctx, `
UPDATE vis_camera SET
  analitico_pausado = FALSE,
  stream_motivo_pausa = NULL,
  stream_falhas_consecutivas = 0,
  stream_tentativas_horarias = 0,
  stream_policy_generation = stream_policy_generation + 1
WHERE id = $1
  AND (stream_motivo_pausa LIKE 'sistema_stream%' OR analitico_pausado IS TRUE)`, cameraID)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	return map[string]any{"camera_id": cameraID, "updated": n}, nil
}

func bumpStreamPolicyGeneration(ctx context.Context, tx *sql.Tx, cameraID int, prevAtivo, newAtivo bool) error {
	if prevAtivo || !newAtivo {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
UPDATE vis_camera SET
  stream_policy_generation = stream_policy_generation + 1,
  stream_falhas_consecutivas = 0,
  stream_tentativas_horarias = 0,
  stream_motivo_pausa = NULL
WHERE id = $1`, cameraID)
	return err
}

func truncateStreamText(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func nullStrVal(v sql.NullString) any {
	if v.Valid && strings.TrimSpace(v.String) != "" {
		return v.String
	}
	return nil
}
