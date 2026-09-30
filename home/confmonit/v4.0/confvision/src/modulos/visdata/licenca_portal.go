package visdata

import (
	"context"
	"fmt"
	"strings"
	"time"
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

// ResumoPortal licencas e acesso do franqueado a partir do Postgres central.
func ResumoPortal(ctx context.Context, idFranqueado string) (map[string]any, error) {
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
	return map[string]any{
		"liberado":        liberado,
		"motivo":          motivo,
		"usa_confvision":  usa,
		"licencas":        lista,
		"resumo":          out["resumo"],
		"faturas_abertas": []any{},
	}, nil
}

func ResumoPortalWithFaturas(ctx context.Context, idFranqueado string, faturas []any) (map[string]any, error) {
	out, err := ResumoPortal(ctx, idFranqueado)
	if err != nil {
		return nil, err
	}
	if faturas != nil {
		out["faturas_abertas"] = faturas
	}
	return out, nil
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

// ComprarLicencas cria licencas disponiveis direto no Postgres (sem Xano/fatura).
func ComprarLicencas(ctx context.Context, idFranqueado string, itens []CompraLicencaItem, observacao string) (map[string]any, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return nil, fmt.Errorf("id_franqueado obrigatorio")
	}
	if len(itens) == 0 {
		return nil, fmt.Errorf("itens obrigatorio")
	}

	now := time.Now().UTC()
	valido := now.Add(30 * 24 * time.Hour)
	obsBase := strings.TrimSpace(observacao)
	if obsBase == "" {
		obsBase = "Compra portal ConfVision"
	}

	var criadas []map[string]any
	for _, item := range itens {
		plano := strings.TrimSpace(item.Plano)
		flags := PlanoFlagsFrom(plano)
		if flags.PlanoLabel == "Nenhum" {
			return nil, fmt.Errorf("plano invalido: %s", plano)
		}
		qtd := item.Quantidade
		if qtd < 1 {
			qtd = 1
		}
		for i := 0; i < qtd; i++ {
			lic, err := insertLicencaDisponivel(ctx, idFranqueado, plano, flags, now, valido, obsBase)
			if err != nil {
				return nil, err
			}
			criadas = append(criadas, lic)
		}
	}

	return map[string]any{
		"licencas": criadas,
		"total":    len(criadas),
		"status":   "Licencas disponiveis no Postgres",
	}, nil
}

func insertLicencaDisponivel(ctx context.Context, idFranqueado, plano string, flags PlanoFlags, pago, valido time.Time, obs string) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var id int
	var created time.Time
	err = db.QueryRowContext(ctx, `
INSERT INTO vis_licenca (
    id_franqueado, plano, unidade, valor, status, observacao, pago_em, valido_ate
) VALUES ($1,$2,$3,$4,'disponivel',$5,$6,$7)
RETURNING id, created_at`,
		idFranqueado, plano, flags.Unidade, flags.Valor, obs, pago, valido,
	).Scan(&id, &created)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id":            id,
		"created_at":    created.UTC().Format(time.RFC3339),
		"id_franqueado": idFranqueado,
		"plano":         plano,
		"unidade":       flags.Unidade,
		"valor":         flags.Valor,
		"status":        "disponivel",
		"pago_em":       pago.Format(time.RFC3339),
		"valido_ate":    valido.Format(time.RFC3339),
	}, nil
}
