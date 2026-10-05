package confvision

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"confvision/src/modulos/visdata"
	"confvision/src/xanopro"
)

// SyncLicencasFranqueadoFromXano espelha vis_licenca do Xano no Postgres operacional.
func SyncLicencasFranqueadoFromXano(ctx context.Context, idFranqueado string) (int, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return 0, fmt.Errorf("id_franqueado obrigatorio")
	}
	licencas, err := FetchXanoLicencas(idFranqueado)
	if err != nil {
		return 0, err
	}
	return visdata.SyncLicencasFromXanoRecords(ctx, licencas)
}

// FetchXanoLicencas lista licencas ConfVision no Xano (fonte financeira).
func FetchXanoLicencas(idFranqueado string) ([]map[string]any, error) {
	raw, err := xanopro.Post("/fp_confvision_resumo_franqueado", map[string]any{
		"id_franqueado": idFranqueado,
	})
	if err != nil {
		return nil, err
	}
	var resp map[string]any
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	return asMapSlice(resp["licencas"]), nil
}

// FetchXanoResumoFranqueado retorna resumo completo do Xano (faturas + acesso).
func FetchXanoResumoFranqueado(idFranqueado string) (map[string]any, error) {
	raw, err := xanopro.Post("/fp_confvision_resumo_franqueado", map[string]any{
		"id_franqueado": idFranqueado,
	})
	if err != nil {
		return nil, err
	}
	var resp map[string]any
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func asMapSlice(v any) []map[string]any {
	switch t := v.(type) {
	case []any:
		out := make([]map[string]any, 0, len(t))
		for _, el := range t {
			if m, ok := el.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	case []map[string]any:
		return t
	default:
		return nil
	}
}
