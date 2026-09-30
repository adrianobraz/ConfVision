package pgcentraldominio

// Apps suportados na marca da central (admConfmonit + provisionador).
var AppsSuportados = []string{
	"franqueadopro",
	"webcliente",
	"webterminal",
	"terminalmovel",
	"webambiente",
	"dialyze",
	"confvision",
}

// HubProduto mapeia app de domínio -> id do card em FP_PRODUTOS_ECOSISTEMA.
func HubProduto(app string) string {
	switch app {
	case "webterminal":
		return "webterminal"
	case "terminalmovel":
		return "terminalmovel"
	case "webambiente":
		return "webambiente"
	case "dialyze":
		return "dialyze"
	case "confvision":
		return "confvision"
	default:
		return ""
	}
}

func AppValido(app string) bool {
	for _, a := range AppsSuportados {
		if a == app {
			return true
		}
	}
	return false
}
