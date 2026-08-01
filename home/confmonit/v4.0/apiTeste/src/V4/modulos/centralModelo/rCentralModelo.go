package centralModeloV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/centralModelo/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/centralModelo/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: false,
	},
	{
		URI:    "/v4/centralModelo/alterarById",
		Metodo: http.MethodPost,
		Funcao: alterarById,
		Seguro: false,
	},
	{
		URI:    "/v4/centralModelo/deletaById",
		Metodo: http.MethodPost,
		Funcao: deletaById,
		Seguro: false,
	},
	{
		URI:    "/v4/centralModelo/deletaAllByIdFabricante",
		Metodo: http.MethodPost,
		Funcao: deletaAllByIdFabricante,
		Seguro: false,
	},
	{
		URI:    "/v4/centralModelo/listarByIdFabricante",
		Metodo: http.MethodPost,
		Funcao: listarByIdFabricante,
		Seguro: false,
	},
	{
		URI:    "/v4/centralModelo/getAtivoById",
		Metodo: http.MethodPost,
		Funcao: getAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/centralModelo/setAtivoById",
		Metodo: http.MethodPost,
		Funcao: setAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/centralModelo/invereteAtivoById",
		Metodo: http.MethodPost,
		Funcao: invereteAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/centralModelo/getEletrificadorById",
		Metodo: http.MethodPost,
		Funcao: getEletrificadorById,
		Seguro: false,
	},
	{
		URI:    "/v4/centralModelo/setEletrificadorById",
		Metodo: http.MethodPost,
		Funcao: setEletrificadorById,
		Seguro: false,
	},
	{
		URI:    "/v4/centralModelo/invereteEletrificadorById",
		Metodo: http.MethodPost,
		Funcao: invereteEletrificadorById,
		Seguro: false,
	},
}
