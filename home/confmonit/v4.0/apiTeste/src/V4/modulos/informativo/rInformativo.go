package informativo

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/informativo/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/informativo/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: false,
	},
	{
		URI:    "/v4/informativo/alteraById",
		Metodo: http.MethodPost,
		Funcao: alteraById,
		Seguro: false,
	},
	{
		URI:    "/v4/informativo/deletaById",
		Metodo: http.MethodPost,
		Funcao: deletaById,
		Seguro: false,
	},
	{
		URI:    "/v4/informativo/listaByIdVinculo",
		Metodo: http.MethodPost,
		Funcao: listaByIdVinculo,
		Seguro: false,
	},
}
