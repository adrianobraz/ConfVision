package visdata

import (
	"context"
	"fmt"
	"strings"
)

var catalogoPlanos = []string{
	"online",
	"sensor_foto",
	"sensor_foto_video",
	"analitico_armado_evento",
	"analitico_armado_foto",
	"analitico_armado_foto_video",
	"analitico_24h_evento",
	"analitico_24h_foto",
	"analitico_24h_foto_video",
	"gravacao_7d",
	"gravacao_15d",
	"gravacao_30d",
	"gravacao_movimento_7d",
	"gravacao_movimento_15d",
	"gravacao_movimento_30d",
	"gravacao_timelapse_7d",
	"gravacao_timelapse_15d",
	"gravacao_timelapse_30d",
}

type CompraLicencaItem struct {
	Plano      string
	Quantidade int
}

// ResumoPortalWithFaturas inclui faturas abertas vindas do Xano.
func ResumoPortalWithFaturas(ctx context.Context, idFranqueado string, faturasAbertas []any) (map[string]any, error) {
	return resumoPortal(ctx, idFranqueado, faturasAbertas)
}

// ListPlanosCatalog retorna catalogo canonico de planos (Postgres-only, sem Xano).
func ListPlanosCatalog(_ context.Context, _ string) map[string]any {
	planos := make([]map[string]any, 0, len(catalogoPlanos))
	for _, slug := range catalogoPlanos {
		flags := PlanoFlagsFrom(slug)
		if flags.PlanoLabel == "Nenhum" {
			continue
		}
		planos = append(planos, map[string]any{
			"plano":              slug,
			"plano_label":        flags.PlanoLabel,
			"unidade":            flags.Unidade,
			"valor":              flags.Valor,
			"valor_com_desconto": flags.Valor,
			"desconto_pct":       0,
		})
	}
	return map[string]any{
		"planos":            planos,
		"desconto_pro_plus": false,
		"total":             len(planos),
	}
}

func ResumoPortal(ctx context.Context, idFranqueado string) (map[string]any, error) {
	return resumoPortal(ctx, idFranqueado, nil)
}

func resumoPortal(ctx context.Context, idFranqueado string, faturas []any) (map[string]any, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return nil, fmt.Errorf("id_franqueado obrigatorio")
	}

	out, err := ListLicencasByFranqueado(ctx, idFranqueado, "", "")
	if err != nil {
		return nil, err
	}

	liberado, motivo := acessoPorLicencasPostgres(ctx, idFranqueado)
	usa := "N"
	if liberado {
		usa = "S"
	}

	lista, _ := out["dados"].([]map[string]any)
	res := map[string]any{
		"liberado":        liberado,
		"motivo":          motivo,
		"usa_confvision":  usa,
		"licencas":        lista,
		"resumo":          out["resumo"],
		"faturas_abertas": []any{},
	}
	if faturas != nil {
		res["faturas_abertas"] = faturas
	}
	return res, nil
}

func acessoPorLicencasPostgres(ctx context.Context, idFranqueado string) (bool, string) {
	db, err := DB()
	if err != nil {
		return false, "sem_acesso"
	}
	var n int
	err = db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM vis_licenca
WHERE id_franqueado = $1
  AND status IN ('disponivel', 'em_uso')
  AND pago_em IS NOT NULL
  AND (valido_ate IS NULL OR valido_ate > NOW())`, idFranqueado).Scan(&n)
	if err == nil && n > 0 {
		return true, "licenca_ativa"
	}

	ok, err := FranqueadoTemRecursosOperacionais(ctx, idFranqueado)
	if err == nil && ok {
		return true, "plano_confvision_ativo"
	}
	return false, "sem_acesso"
}
