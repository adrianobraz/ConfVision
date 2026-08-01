package pacotev4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/pacote/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/pacote/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: false,
	},
	{
		URI:    "/v4/pacote/alteraById",
		Metodo: http.MethodPost,
		Funcao: alteraById,
		Seguro: false,
	},
	{
		URI:    "/v4/pacote/deletaById",
		Metodo: http.MethodPost,
		Funcao: deletaById,
		Seguro: false,
	},
	{
		URI:    "/v4/pacote/deletaAllByVinculo",
		Metodo: http.MethodPost,
		Funcao: deletaAllByVinculo,
		Seguro: false,
	},
	{
		URI:    "/v4/pacote/listaByVinculo",
		Metodo: http.MethodPost,
		Funcao: listaByVinculo,
		Seguro: false,
	},
	{
		URI:    "/v4/pacote/getAtivoById",
		Metodo: http.MethodPost,
		Funcao: getAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/pacote/setAtivoById",
		Metodo: http.MethodPost,
		Funcao: setAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/pacote/inverteAtivoById",
		Metodo: http.MethodPost,
		Funcao: inverteAtivoById,
		Seguro: false,
	},
}
