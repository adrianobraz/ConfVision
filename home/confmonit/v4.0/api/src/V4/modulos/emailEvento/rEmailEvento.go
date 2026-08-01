package emailEventoV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/emailEvento/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/emailEvento/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: false,
	},
	{
		URI:    "/v4/emailEvento/getDadosByIdDispositivo",
		Metodo: http.MethodPost,
		Funcao: getDadosByIdDispositivo,
		Seguro: false,
	},
	{
		URI:    "/v4/emailEvento/alteraById",
		Metodo: http.MethodPost,
		Funcao: alteraById,
		Seguro: false,
	},
	{
		URI:    "/v4/emailEvento/deletaById",
		Metodo: http.MethodPost,
		Funcao: deletaById,
		Seguro: false,
	},
	{
		URI:    "/v4/emailEvento/deletaByIdDispositivo",
		Metodo: http.MethodPost,
		Funcao: deletaByIdDispositivo,
		Seguro: false,
	},
}
