package fabricanteV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/fabricante/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/fabricante/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: false,
	},
	{
		URI:    "/v4/fabricante/alterarById",
		Metodo: http.MethodPost,
		Funcao: alterarById,
		Seguro: false,
	},
	{
		URI:    "/v4/fabricante/deletaById",
		Metodo: http.MethodPost,
		Funcao: deletaById,
		Seguro: false,
	},
	{
		URI:    "/v4/fabricante/listar",
		Metodo: http.MethodPost,
		Funcao: listar,
		Seguro: false,
	},
	{
		URI:    "/v4/fabricante/getAtivoById",
		Metodo: http.MethodPost,
		Funcao: getAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/fabricante/setAtivoById",
		Metodo: http.MethodPost,
		Funcao: setAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/fabricante/inverteAtivoById",
		Metodo: http.MethodPost,
		Funcao: inverteAtivoById,
		Seguro: false,
	},
}
