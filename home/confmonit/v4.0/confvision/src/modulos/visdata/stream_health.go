package visdata

import (
	"context"
	"database/sql"
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
  stream_ultimo_erro_em = NULL,
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
			lastErr, errClass := streamErrorFields(r)
			_, err = db.ExecContext(ctx, `
UPDATE vis_camera SET
  stream_falhas_consecutivas = $2,
  stream_tentativas_horarias = $3,
  stream_ultimo_erro = COALESCE($4, stream_ultimo_erro),
  stream_erro_classe = COALESCE($5, stream_erro_classe),
  stream_ultimo_erro_em = CASE WHEN $4 IS NOT NULL OR $5 IS NOT NULL THEN NOW() ELSE stream_ultimo_erro_em END
WHERE id = $1`, r.CameraID, f, h, lastErr, errClass)
		case "stream_incident":
			lastErr, errClass := streamErrorFields(r)
			if lastErr == nil && errClass == nil {
				continue
			}
			_, err = db.ExecContext(ctx, `
UPDATE vis_camera SET
  stream_ultimo_erro = COALESCE($2, stream_ultimo_erro),
  stream_erro_classe = COALESCE($3, stream_erro_classe),
  stream_ultimo_erro_em = NOW()
WHERE id = $1`, r.CameraID, lastErr, errClass)
		case "pause_analytic":
			reason := "sistema_stream"
			if r.PauseReason != nil && strings.TrimSpace(*r.PauseReason) != "" {
				reason = truncateStreamText(*r.PauseReason, 120)
			}
			lastErr, errClass := streamErrorFields(r)
			_, err = db.ExecContext(ctx, `
UPDATE vis_camera SET
  analitico_pausado = TRUE,
  stream_motivo_pausa = $2,
  stream_ultimo_erro = COALESCE($3, stream_ultimo_erro),
  stream_erro_classe = COALESCE($4, stream_erro_classe),
  stream_ultimo_erro_em = CASE WHEN $3 IS NOT NULL OR $4 IS NOT NULL THEN NOW() ELSE stream_ultimo_erro_em END
WHERE id = $1`, r.CameraID, reason, lastErr, errClass)
		default:
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func streamErrorFields(r CameraStreamHealthInput) (lastErr, errClass any) {
	if r.LastError != nil && strings.TrimSpace(*r.LastError) != "" {
		lastErr = truncateStreamText(*r.LastError, 500)
	}
	if r.ErrorClass != nil && strings.TrimSpace(*r.ErrorClass) != "" {
		errClass = truncateStreamText(*r.ErrorClass, 64)
	}
	return lastErr, errClass
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
