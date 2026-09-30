package pgcatalogo

import (
	"context"
	"encoding/json"
)

// SeedConfvisionLicencas — 18 licencas por camera/gravacao (produto confvision_licenca).
func SeedConfvisionLicencas(ctx context.Context, idCentral string) (int, error) {
	db, err := open()
	if err != nil {
		return 0, err
	}
	items := []struct {
		plano, nome, unidade string
		valor               float64
	}{
		{"online", "Camera online", "camera", 2.99},
		{"sensor_foto", "Sensor foto", "camera", 7.99},
		{"sensor_foto_video", "Sensor foto + video", "camera", 9.99},
		{"analitico_armado_evento", "Analitico armado — so evento", "camera", 11.99},
		{"analitico_armado_foto", "Analitico armado — foto", "camera", 13.99},
		{"analitico_armado_foto_video", "Analitico armado — foto + video", "camera", 14.99},
		{"analitico_24h_evento", "Analitico 24h — so evento", "camera", 16.99},
		{"analitico_24h_foto", "Analitico 24h — foto", "camera", 18.99},
		{"analitico_24h_foto_video", "Analitico 24h — foto + video", "camera", 19.99},
		{"gravacao_7d", "Gravacao continua 7 dias", "gravacao", 12.99},
		{"gravacao_15d", "Gravacao continua 15 dias", "gravacao", 17.99},
		{"gravacao_30d", "Gravacao continua 30 dias", "gravacao", 24.99},
		{"gravacao_movimento_7d", "Gravacao por movimento 7 dias", "gravacao", 9.99},
		{"gravacao_movimento_15d", "Gravacao por movimento 15 dias", "gravacao", 13.99},
		{"gravacao_movimento_30d", "Gravacao por movimento 30 dias", "gravacao", 19.99},
		{"gravacao_timelapse_7d", "Gravacao timelapse inteligente 7 dias", "gravacao", 6.99},
		{"gravacao_timelapse_15d", "Gravacao timelapse inteligente 15 dias", "gravacao", 9.99},
		{"gravacao_timelapse_30d", "Gravacao timelapse inteligente 30 dias", "gravacao", 13.99},
	}
	n := 0
	for _, it := range items {
		lim, _ := json.Marshal(map[string]string{"unidade": it.unidade})
		res, err := db.ExecContext(ctx, `
			INSERT INTO fp_catalogo_produto
			(id_central, produto, plano, nome_exibicao, valor_mensal, valor_piso_breakglass,
			 retencao_dias, limites_json, modulos_json, ativo)
			VALUES ($1,'confvision_licenca',$2,$3,$4,$4,30,$5,'{}','S')
			ON CONFLICT (id_central, id_representante, produto, plano) DO NOTHING
		`, idCentral, it.plano, it.nome, it.valor, lim)
		if err != nil {
			return n, err
		}
		if aff, _ := res.RowsAffected(); aff > 0 {
			n++
		}
	}
	return n, nil
}
