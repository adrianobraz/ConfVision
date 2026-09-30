package contrato

// Modulos e limites padrão por plano FranqueadoPro (espelha fp-planos-dados.js)

func planoLiberadoNoPlano(minimo, plano string) bool {
	switch plano {
	case "pro_plus":
		return true
	case "pro":
		return minimo == "lite" || minimo == "pro"
	case "lite":
		return minimo == "lite"
	}
	return false
}

func modulosPorPlanoFP(plano string) map[string]bool {
	chaves := []struct {
		chave  string
		minimo string
	}{
		{"configuracao.dispositivo", "lite"},
		{"configuracao.usuarios-alarme", "lite"},
		{"configuracao.setores-alarme", "lite"},
		{"configuracao.grade", "pro"},
		{"configuracao.procedimento", "pro_plus"},
		{"configuracao.contactid", "pro_plus"},
		{"configuracao.email-evento", "pro"},
		{"atendimento.gerar-evento", "lite"},
		{"atendimento.configuracao", "pro_plus"},
		{"atendimento.inteligencia-artificial", "pro_plus"},
		{"minha-empresa.comercial.cliente", "lite"},
		{"minha-empresa.comercial.pacotes", "pro_plus"},
		{"minha-empresa.operacional", "pro"},
		{"minha-empresa.operacional.tecnico", "pro"},
		{"minha-empresa.operacional.viatura", "pro"},
		{"minha-empresa.gestao.dados", "lite"},
		{"minha-empresa.gestao.responsaveis", "lite"},
		{"minha-empresa.gestao.operador", "lite"},
		{"minha-empresa.gestao.permissoes", "pro_plus"},
		{"whitelabel", "pro"},
		{"relatorio.atendimento", "pro"},
		{"relatorio.ligacoes", "pro"},
		{"relatorio.eventos.lista", "lite"},
		{"relatorio.eventos.config-grupo", "pro"},
		{"relatorio.alarmes.armados", "lite"},
		{"relatorio.alarmes.desarmados", "lite"},
		{"relatorio.clientes.disp", "lite"},
		{"relatorio.clientes.lista", "lite"},
		{"relatorio.clientes.ativos", "lite"},
		{"relatorio.clientes.inativos", "lite"},
		{"dashboard.sem-comunicacao", "lite"},
		{"dashboard.eventos", "pro"},
		{"franqueadopro.saas", "pro_plus"},
		{"relatorio.central-disparos", "pro_plus"},
	}
	out := make(map[string]bool, len(chaves))
	for _, c := range chaves {
		out[c.chave] = planoLiberadoNoPlano(c.minimo, plano)
	}
	return out
}

func limitesPorPlanoFP(plano string) map[string]int {
	switch plano {
	case "lite":
		return map[string]int{
			"clientes_max": 50, "contas_max": 100,
			"usuarios_alarme_max": 5, "setores_alarme_max": 4,
		}
	case "pro":
		return map[string]int{
			"clientes_max": 500, "contas_max": 1000,
			"usuarios_alarme_max": 20, "setores_alarme_max": 15,
		}
	case "pro_plus":
		return map[string]int{
			"clientes_max": 9999, "contas_max": 9999,
			"usuarios_alarme_max": 999, "setores_alarme_max": 999,
		}
	default:
		return map[string]int{}
	}
}

func retencaoPorPlanoFP(plano string) int {
	switch plano {
	case "lite":
		return 30
	case "pro":
		return 90
	case "pro_plus":
		return 180
	default:
		return 30
	}
}

func limitesDeCotaQtd(qtd int) map[string]int {
	if qtd < 0 {
		qtd = 0
	}
	return map[string]int{
		"clientes_max":        qtd,
		"contas_max":          qtd * 2,
		"usuarios_alarme_max": qtd * 4,
		"setores_alarme_max":  qtd * 10,
		"cota_quantidade":     qtd,
	}
}

func mergeLimites(base, extra map[string]int) map[string]int {
	out := make(map[string]int)
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		if v > out[k] {
			out[k] = v
		}
	}
	return out
}

func mergeModulos(base, extra map[string]bool) map[string]bool {
	out := make(map[string]bool)
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		if v {
			out[k] = true
		}
	}
	return out
}
