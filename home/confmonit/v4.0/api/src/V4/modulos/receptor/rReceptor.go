package receptorV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/receptor/recebeEvento",
		Metodo: http.MethodPost,
		Funcao: recebeEvento,
		Seguro: false,
	},
	// {
	// 	URI:    "/v4/",
	// 	Metodo: http.MethodPost,
	// 	Funcao: ,
	// 	Seguro: false,
	// },
}
