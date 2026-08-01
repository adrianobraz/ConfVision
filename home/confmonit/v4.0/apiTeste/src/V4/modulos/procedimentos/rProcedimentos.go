package procedimentosV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/procedimento/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/procedimento/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: false,
	},
	{
		URI:    "/v4/procedimento/alteraById",
		Metodo: http.MethodPost,
		Funcao: alteraById,
		Seguro: false,
	},
	{
		URI:    "/v4/procedimento/deletaById",
		Metodo: http.MethodPost,
		Funcao: deletaById,
		Seguro: false,
	},
	{
		URI:    "/v4/procedimento/deletaAllByIdFranqueado",
		Metodo: http.MethodPost,
		Funcao: deletaAllByIdFranqueado,
		Seguro: false,
	},
	{
		URI:    "/v4/procedimento/deletaAllByIdCliente",
		Metodo: http.MethodPost,
		Funcao: deletaAllByIdCliente,
		Seguro: false,
	},
	{
		URI:    "/v4/procedimento/listaByAllIdFranqueado",
		Metodo: http.MethodPost,
		Funcao: listaByAllIdFranqueado,
		Seguro: false,
	},
	{
		URI:    "/v4/procedimento/getAtivoByid",
		Metodo: http.MethodPost,
		Funcao: getAtivoByid,
		Seguro: false,
	},
	{
		URI:    "/v4/procedimento/setAtivoByid",
		Metodo: http.MethodPost,
		Funcao: setAtivoByid,
		Seguro: false,
	},
	{
		URI:    "/v4/procedimento/inverteAtivoByid",
		Metodo: http.MethodPost,
		Funcao: inverteAtivoByid,
		Seguro: false,
	},
}
