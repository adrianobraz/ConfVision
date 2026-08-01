package licenca

import "strings"

// ModuloInfo descreve regra de plano por chave de menu.
type ModuloInfo struct {
	Chave  string
	Minimo string // lite | pro | pro_plus
}

var catalogoModulos = []ModuloInfo{
	{Chave: "configuracao.dispositivo", Minimo: "lite"},
	{Chave: "configuracao.usuarios-alarme", Minimo: "lite"},
	{Chave: "configuracao.setores-alarme", Minimo: "lite"},
	{Chave: "configuracao.grade", Minimo: "pro"},
	{Chave: "configuracao.procedimento", Minimo: "pro_plus"},
	{Chave: "configuracao.contactid", Minimo: "pro_plus"},
	{Chave: "configuracao.email-evento", Minimo: "pro"},
	{Chave: "atendimento.gerar-evento", Minimo: "lite"},
	{Chave: "atendimento.inteligencia-artificial", Minimo: "pro_plus"},
	{Chave: "minha-empresa.comercial.cliente", Minimo: "lite"},
	{Chave: "minha-empresa.operacional", Minimo: "pro"},
	{Chave: "minha-empresa.operacional.tecnico", Minimo: "pro"},
	{Chave: "minha-empresa.operacional.viatura", Minimo: "pro"},
	{Chave: "minha-empresa.gestao.dados", Minimo: "lite"},
	{Chave: "minha-empresa.gestao.responsaveis", Minimo: "lite"},
	{Chave: "minha-empresa.gestao.meu-plano", Minimo: "lite"},
	{Chave: "minha-empresa.gestao.montar-plano", Minimo: "lite"},
	{Chave: "minha-empresa.gestao.faturas", Minimo: "lite"},
	{Chave: "minha-empresa.gestao.permissoes", Minimo: "pro_plus"},
	{Chave: "whitelabel", Minimo: "pro"},
	{Chave: "relatorio.atendimento", Minimo: "pro"},
	{Chave: "relatorio.ligacoes", Minimo: "pro"},
	{Chave: "relatorio.eventos.lista", Minimo: "lite"},
	{Chave: "relatorio.eventos.config-grupo", Minimo: "pro"},
	{Chave: "relatorio.alarmes.armados", Minimo: "lite"},
	{Chave: "relatorio.alarmes.desarmados", Minimo: "lite"},
	{Chave: "relatorio.clientes.disp", Minimo: "lite"},
	{Chave: "relatorio.clientes.lista", Minimo: "lite"},
	{Chave: "relatorio.clientes.ativos", Minimo: "lite"},
	{Chave: "relatorio.clientes.inativos", Minimo: "lite"},
	{Chave: "relatorio.central-disparos", Minimo: "pro_plus"},
	{Chave: "dashboard.sem-comunicacao", Minimo: "lite"},
	{Chave: "dashboard.eventos", Minimo: "pro"},
	{Chave: "franqueadopro.saas", Minimo: "pro_plus"},
}

var modulosControlados []string

func init() {
	modulosControlados = make([]string, len(catalogoModulos))
	for i, m := range catalogoModulos {
		modulosControlados[i] = m.Chave
	}
}

func liberadoNoPlano(minimo, plano string) bool {
	switch plano {
	case "pro_plus":
		return true
	case "pro":
		return minimo == "lite" || minimo == "pro"
	case "lite":
		return minimo == "lite"
	default:
		return minimo == "lite"
	}
}

func modulosPadraoPorPlano(plano string) map[string]bool {
	out := map[string]bool{}
	for _, m := range catalogoModulos {
		out[m.Chave] = liberadoNoPlano(m.Minimo, plano)
	}
	return out
}

func limitesPadraoPorPlano(plano string) map[string]int {
	switch plano {
	case "lite":
		return map[string]int{
			"usuarios_alarme_max": 5,
			"setores_alarme_max":  4,
			"clientes_max":        50,
			"contas_max":          100,
		}
	case "pro":
		return map[string]int{
			"usuarios_alarme_max": 20,
			"setores_alarme_max":  15,
			"clientes_max":        500,
			"contas_max":          1000,
		}
	case "pro_plus":
		return map[string]int{
			"usuarios_alarme_max": 999,
			"setores_alarme_max":  999,
			"clientes_max":        9999,
			"contas_max":          9999,
		}
	default:
		return map[string]int{}
	}
}

func normalizarChaveAddon(chave string) string {
	chave = strings.TrimSpace(chave)
	if chave == "franqueadopro" {
		return "franqueadopro.saas"
	}
	return chave
}

func normalizarChaveLegacyModulo(mod map[string]bool) {
	if mod == nil {
		return
	}
	if v, ok := mod["franqueadopro"]; ok {
		if v {
			mod["franqueadopro.saas"] = true
		}
		delete(mod, "franqueadopro")
	}
}

func sliceTextoJSON(raw []interface{}) []string {
	out := []string{}
	for _, item := range raw {
		switch t := item.(type) {
		case string:
			if strings.TrimSpace(t) != "" {
				out = append(out, t)
			}
		}
	}
	return out
}

func mesclarAddonsPagos(mod map[string]bool, addons []string, pendentes []string) {
	if mod == nil {
		return
	}
	pend := map[string]bool{}
	for _, ch := range pendentes {
		pend[normalizarChaveAddon(ch)] = true
	}
	for _, ch := range addons {
		ch = normalizarChaveAddon(ch)
		if ch == "" || pend[ch] {
			continue
		}
		mod[ch] = true
	}
}

func mesclarAddonsEfetivos(mod map[string]bool, addons []string) {
	if mod == nil {
		return
	}
	for _, ch := range addons {
		ch = normalizarChaveAddon(ch)
		if ch == "" {
			continue
		}
		mod[ch] = true
	}
}

func flattenModulosJSON(raw map[string]interface{}) map[string]bool {
	out := map[string]bool{}
	if raw == nil {
		return out
	}
	for k, v := range raw {
		switch t := v.(type) {
		case bool:
			out[k] = t
		case map[string]interface{}:
			for sk, sv := range t {
				key := k + "." + sk
				if b, ok := sv.(bool); ok {
					out[key] = b
				}
			}
		case float64:
			out[k] = t != 0
		}
	}
	return out
}

func mergeModulos(plano string, api map[string]bool) map[string]bool {
	base := modulosPadraoPorPlano(plano)
	if len(api) == 0 {
		return base
	}
	out := map[string]bool{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range api {
		out[k] = v
	}
	return out
}

func moduloLiberadoNoPlano(plano, chave string) bool {
	for _, m := range catalogoModulos {
		if m.Chave == chave {
			return liberadoNoPlano(m.Minimo, plano)
		}
	}
	return true
}

func mergeLimites(plano string, api map[string]int) map[string]int {
	base := limitesPadraoPorPlano(plano)
	if len(api) == 0 {
		return base
	}
	out := map[string]int{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range api {
		out[k] = v
	}
	return out
}

func parseLimitesJSON(raw map[string]interface{}) map[string]int {
	out := map[string]int{}
	if raw == nil {
		return out
	}
	for k, v := range raw {
		switch t := v.(type) {
		case float64:
			out[k] = int(t)
		case int:
			out[k] = t
		}
	}
	return out
}

func moduloLiberadoNoMapa(mod map[string]bool, chave string) bool {
	if mod == nil {
		return true
	}
	if v, ok := mod[chave]; ok {
		return v
	}
	if chave == "franqueadopro.saas" {
		if v, ok := mod["franqueadopro"]; ok && v {
			return true
		}
	}
	parts := splitChave(chave)
	for i := len(parts) - 1; i >= 1; i-- {
		parent := joinChave(parts[:i])
		if v, ok := mod[parent]; ok && !v {
			return false
		}
	}
	for _, m := range catalogoModulos {
		if m.Chave == chave {
			return false
		}
	}
	return true
}

func splitChave(chave string) []string {
	out := []string{}
	cur := ""
	for _, c := range chave {
		if c == '.' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			continue
		}
		cur += string(c)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

func joinChave(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += "." + parts[i]
	}
	return out
}
