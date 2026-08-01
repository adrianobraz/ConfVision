package gradeV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/grade/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/grade/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: false,
	},
	{
		URI:    "/v4/grade/alteraById",
		Metodo: http.MethodPost,
		Funcao: alteraById,
		Seguro: false,
	},
	{
		URI:    "/v4/grade/deletaById",
		Metodo: http.MethodPost,
		Funcao: deletaById,
		Seguro: false,
	},
	{
		URI:    "/v4/grade/deletaAllByIdDispositivo",
		Metodo: http.MethodPost,
		Funcao: deletaAllByIdDispositivo,
		Seguro: false,
	},
	{
		URI:    "/v4/grade/listarByIdDispositivo",
		Metodo: http.MethodPost,
		Funcao: listarByIdDispositivo,
		Seguro: false,
	},
	{
		URI:    "/v4/grade/getAtivo",
		Metodo: http.MethodPost,
		Funcao: getAtivo,
		Seguro: false,
	},
	{
		URI:    "/v4/grade/setAtivo",
		Metodo: http.MethodPost,
		Funcao: setAtivo,
		Seguro: false,
	},
	{
		URI:    "/v4/grade/desabiltaAllAtivoByIdDispositivo",
		Metodo: http.MethodPost,
		Funcao: desabiltaAllAtivoByIdDispositivo,
		Seguro: false,
	},
	{
		URI:    "/v4/grade/inverteAtivo",
		Metodo: http.MethodPost,
		Funcao: inverteAtivo,
		Seguro: false,
	},
}
