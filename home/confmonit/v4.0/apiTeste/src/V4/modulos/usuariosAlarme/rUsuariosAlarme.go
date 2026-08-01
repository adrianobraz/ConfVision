package usuariosAlarmeV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/usuarioAlarme/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/usuarioAlarme/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: false,
	},
	{
		URI:    "/v4/usuarioAlarme/alteraById",
		Metodo: http.MethodPost,
		Funcao: alteraById,
		Seguro: false,
	},
	{
		URI:    "/v4/usuarioAlarme/deletaById",
		Metodo: http.MethodPost,
		Funcao: deletaById,
		Seguro: false,
	},
	{
		URI:    "/v4/usuarioAlarme/deletaAllByIdDispositivo",
		Metodo: http.MethodPost,
		Funcao: deletaAllByIdDispositivo,
		Seguro: false,
	},
	{
		URI:    "/v4/usuarioAlarme/listaByIdDispositivo",
		Metodo: http.MethodPost,
		Funcao: listaByIdDispositivo,
		Seguro: false,
	},
	{
		URI:    "/v4/usuarioAlarme/getAtivoByid",
		Metodo: http.MethodPost,
		Funcao: getAtivoByid,
		Seguro: false,
	},
	{
		URI:    "/v4/usuarioAlarme/setAtivoByid",
		Metodo: http.MethodPost,
		Funcao: setAtivoByid,
		Seguro: false,
	},
	{
		URI:    "/v4/usuarioAlarme/inverteAtivoByid",
		Metodo: http.MethodPost,
		Funcao: inverteAtivoByid,
		Seguro: false,
	},
}
