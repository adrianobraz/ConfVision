package confvision

import (
	"context"

	"confvision/src/config"
	"confvision/src/modulos/visdata"
)

func configVisEnabled() bool {
	return config.VisPostgresEnabled
}

func listCamerasByFranqueadoCtx(ctx context.Context, idFranqueado string) ([]map[string]any, error) {
	return visdata.ListCamerasByFranqueado(ctx, idFranqueado)
}
