package tecnicoV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/tecnico/logar",
		Metodo: http.MethodPost,
		Funcao: logar,
		Seguro: false,
	},
	{
		URI:    "/v4/tecnico/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/tecnico/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: false,
	},
	{
		URI:    "/v4/tecnico/alteraById",
		Metodo: http.MethodPost,
		Funcao: alteraById,
		Seguro: false,
	},
	{
		URI:    "/v4/tecnico/deletaById",
		Metodo: http.MethodPost,
		Funcao: deletaById,
		Seguro: false,
	},
	{
		URI:    "/v4/tecnico/deletaAllByIdVinculo",
		Metodo: http.MethodPost,
		Funcao: deletaAllByIdVinculo,
		Seguro: false,
	},
	{
		URI:    "/v4/tecnico/listaByIdVinculo",
		Metodo: http.MethodPost,
		Funcao: listaByIdVinculo,
		Seguro: false,
	},
	{
		URI:    "/v4/tecnico/resetSenhaById",
		Metodo: http.MethodPost,
		Funcao: resetSenhaById,
		Seguro: false,
	},
	{
		URI:    "/v4/tecnico/alteraSenhaById",
		Metodo: http.MethodPost,
		Funcao: alteraSenhaById,
		Seguro: false,
	},
	{
		URI:    "/v4/tecnico/getAtivoById",
		Metodo: http.MethodPost,
		Funcao: getAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/tecnico/setAtivoById",
		Metodo: http.MethodPost,
		Funcao: setAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/tecnico/inverteAtivoById",
		Metodo: http.MethodPost,
		Funcao: inverteAtivoById,
		Seguro: false,
	},
}
