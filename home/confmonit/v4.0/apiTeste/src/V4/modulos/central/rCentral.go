package centralV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/central/logar",
		Metodo: http.MethodPost,
		Funcao: logar,
		Seguro: false,
	},
}
