package viaturaV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/viatura/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/viatura/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: false,
	},
	{
		URI:    "/v4/viatura/alteraById",
		Metodo: http.MethodPost,
		Funcao: alteraById,
		Seguro: false,
	},
	{
		URI:    "/v4/viatura/deleteById",
		Metodo: http.MethodPost,
		Funcao: deleteById,
		Seguro: false,
	},
	{
		URI:    "/v4/viatura/deleteAllByIdFranqueado",
		Metodo: http.MethodPost,
		Funcao: deleteAllByIdFranqueado,
		Seguro: false,
	},
	{
		URI:    "/v4/viatura/listaByIdFranqueado",
		Metodo: http.MethodPost,
		Funcao: listaByIdFranqueado,
		Seguro: false,
	},
	{
		URI:    "/v4/viatura/setEmailById",
		Metodo: http.MethodPost,
		Funcao: setEmailById,
		Seguro: false,
	},
	{
		URI:    "/v4/viatura/verificaEmailLivreByEmail",
		Metodo: http.MethodPost,
		Funcao: verificaEmailLivreByEmail,
		Seguro: false,
	},
	{
		URI:    "/v4/viatura/getAtivoById",
		Metodo: http.MethodPost,
		Funcao: getAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/viatura/setAtivoById",
		Metodo: http.MethodPost,
		Funcao: setAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/viatura/inverteAtivoById",
		Metodo: http.MethodPost,
		Funcao: inverteAtivoById,
		Seguro: false,
	},
}
