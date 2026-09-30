package visdata

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

// LicencaUsoValido indica se valido_ate ainda permite usar a licença (cadastro/troca).
func LicencaUsoValido(valido sql.NullTime, now time.Time) bool {
	if !valido.Valid {
		return true
	}
	return !valido.Time.Before(now)
}

// ExpireLicencasDisponiveisVencidas marca status expirada quando valido_ate passou e não há câmera vinculada.
func ExpireLicencasDisponiveisVencidas(ctx context.Context, idFranqueado string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	q := `
UPDATE vis_licenca SET status = 'expirada'
WHERE status = 'disponivel'
  AND valido_ate IS NOT NULL
  AND valido_ate < NOW()
  AND (vis_camera_id IS NULL OR vis_camera_id = 0)`
	args := []any{}
	if strings.TrimSpace(idFranqueado) != "" {
		q += ` AND id_franqueado = $1`
		args = append(args, idFranqueado)
	}
	_, err = db.ExecContext(ctx, q, args...)
	return err
}
