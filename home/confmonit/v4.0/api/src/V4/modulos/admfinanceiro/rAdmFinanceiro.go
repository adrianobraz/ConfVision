package admfinanceiroV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/admfinanceiro/logar",
		Metodo: http.MethodPost,
		Funcao: logar,
		Seguro: false,
	},
	{
		URI:    "/v4/admfinanceiro/setFlag",
		Metodo: http.MethodPost,
		Funcao: setFlag,
		Seguro: false,
	},
}
