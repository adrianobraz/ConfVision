package feedback

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/feedback/enviar",
		Metodo: http.MethodPost,
		Funcao: enviar,
		Seguro: false,
	},
}
