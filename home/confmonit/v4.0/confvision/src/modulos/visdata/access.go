package visdata

import "context"

// FranqueadoTemRecursosOperacionais indica se o franqueado ja possui cameras ou licencas no Postgres central.
func FranqueadoTemRecursosOperacionais(ctx context.Context, idFranqueado string) (bool, error) {
	if idFranqueado == "" {
		return false, nil
	}
	db, err := DB()
	if err != nil {
		return false, err
	}
	var ok bool
	err = db.QueryRowContext(ctx, `
SELECT EXISTS(SELECT 1 FROM vis_camera WHERE id_franqueado = $1)
    OR EXISTS(SELECT 1 FROM vis_licenca WHERE id_franqueado = $1)`, idFranqueado).Scan(&ok)
	return ok, err
}
