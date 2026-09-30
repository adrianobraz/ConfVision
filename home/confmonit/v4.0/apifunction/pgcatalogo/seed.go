package pgcatalogo

import (
	"context"
	"encoding/json"
	"strings"
)

func SeedProdutos(ctx context.Context, idCentral string) (int, error) {
	db, err := open()
	if err != nil {
		return 0, err
	}
	n := 0
	produtos := []struct {
		prod, plano, nome string
		valor             float64
		retencao          int
		limites, modulos  map[string]interface{}
	}{
		{"franqueadopro", "lite", "FranqueadoPro Lite", 99, 30, limitesLite(), modulosLite()},
		{"franqueadopro", "pro", "FranqueadoPro Pro", 199, 90, limitesPro(), modulosPro()},
		{"franqueadopro", "pro_plus", "FranqueadoPro Pro+", 299, 180, limitesProPlus(), modulosProPlus()},
		{"webterminal", "lite", "WebTerminal Lite", 49, 30, nil, nil},
		{"webterminal", "pro", "WebTerminal Pro", 79, 90, nil, nil},
		{"webterminal", "pro_plus", "WebTerminal Pro+", 129, 180, nil, nil},
		{"terminalmovel", "padrao", "Terminal Movel", 59, 30, nil, nil},
		{"confvision", "padrao", "ConfVision - Plano (software)", 79, 30, nil, nil},
		{"webambiente", "pro", "webAmbiente Pro", 69, 90, nil, nil},
		{"webambiente", "pro_plus", "webAmbiente Pro+", 99, 180, nil, nil},
		{"dialyze", "padrao", "Dialyze", 99, 30, nil, nil},
	}
	for _, p := range produtos {
		lim := []byte("{}")
		mod := []byte("{}")
		if p.limites != nil {
			lim, _ = json.Marshal(p.limites)
		}
		if p.modulos != nil {
			mod, _ = json.Marshal(p.modulos)
		}
		res, err := db.ExecContext(ctx, `
			INSERT INTO fp_catalogo_produto
			(id_central, produto, plano, nome_exibicao, valor_mensal, valor_piso_breakglass,
			 retencao_dias, limites_json, modulos_json, ativo)
			VALUES ($1,$2,$3,$4,$5,$5,$6,$7,$8,'S')
			ON CONFLICT (id_central, id_representante, produto, plano) DO NOTHING
		`, idCentral, p.prod, p.plano, p.nome, p.valor, p.retencao, lim, mod)
		if err != nil {
			return n, err
		}
		if aff, _ := res.RowsAffected(); aff > 0 {
			n++
		}
	}
	return n, nil
}

func SeedPacotes(ctx context.Context, idCentral string) (int, error) {
	db, err := open()
	if err != nil {
		return 0, err
	}
	pacotes := []struct {
		nome string
		qtd  int
		val  float64
	}{
		{"Cota 50", 50, 50},
		{"Cota 200", 200, 100},
		{"Cota 800", 800, 200},
	}
	n := 0
	for _, p := range pacotes {
		var exists int
		if err := db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM fp_pacote_cota WHERE id_central = $1 AND quantidade = $2
		`, idCentral, p.qtd).Scan(&exists); err != nil {
			return n, err
		}
		if exists > 0 {
			continue
		}
		if _, err := db.ExecContext(ctx, `
			INSERT INTO fp_pacote_cota (id_central, nome, quantidade, valor, valor_piso_breakglass, ativo)
			VALUES ($1,$2,$3,$4,$4,'S')
		`, idCentral, p.nome, p.qtd, p.val); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func limitesLite() map[string]interface{} {
	return map[string]interface{}{
		"usuarios_alarme_max": 5, "setores_alarme_max": 4,
		"clientes_max": 50, "contas_max": 100,
	}
}

func limitesPro() map[string]interface{} {
	return map[string]interface{}{
		"usuarios_alarme_max": 20, "setores_alarme_max": 15,
		"clientes_max": 500, "contas_max": 1000,
	}
}

func limitesProPlus() map[string]interface{} {
	return map[string]interface{}{
		"usuarios_alarme_max": 0, "setores_alarme_max": 0,
		"clientes_max": 0, "contas_max": 0,
	}
}

func modulosLite() map[string]interface{} {
	return map[string]interface{}{"whitelabel": false, "franqueadopro.saas": false}
}

func modulosPro() map[string]interface{} {
	return map[string]interface{}{"whitelabel": true, "franqueadopro.saas": false}
}

func modulosProPlus() map[string]interface{} {
	return map[string]interface{}{"whitelabel": true, "franqueadopro.saas": true}
}

func normalizeAtivo(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "N" {
		return "N"
	}
	return "S"
}
