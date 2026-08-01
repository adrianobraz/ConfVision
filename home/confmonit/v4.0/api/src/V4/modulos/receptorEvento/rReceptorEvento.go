package receptorEvento

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/receptorEvento/carregarDados",
		Metodo: http.MethodPost,
		Funcao: carregarDados,
		Seguro: false,
	},
}
