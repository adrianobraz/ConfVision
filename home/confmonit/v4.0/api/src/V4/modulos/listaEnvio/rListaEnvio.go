package listaEnvioV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/listaEnvio/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/listaEnvio/deleteByIdAlvo",
		Metodo: http.MethodPost,
		Funcao: deleteByIdAlvo,
		Seguro: false,
	},
	{
		URI:    "/v4/listaEnvio/getEmailAtivoByIdAlvo",
		Metodo: http.MethodPost,
		Funcao: getEmailAtivoByIdAlvo,
		Seguro: false,
	},
	{
		URI:    "/v4/listaEnvio/setEmailAtivoByIdAlvo",
		Metodo: http.MethodPost,
		Funcao: setEmailAtivoByIdAlvo,
		Seguro: false,
	},
	{
		URI:    "/v4/listaEnvio/inverteEmailAtivoByIdAlvo",
		Metodo: http.MethodPost,
		Funcao: inverteEmailAtivoByIdAlvo,
		Seguro: false,
	},
	{
		URI:    "/v4/listaEnvio/getSmsAtivoByIdAlvo",
		Metodo: http.MethodPost,
		Funcao: getSmsAtivoByIdAlvo,
		Seguro: false,
	},
	{
		URI:    "/v4/listaEnvio/setSmsAtivoByIdAlvo",
		Metodo: http.MethodPost,
		Funcao: setSmsAtivoByIdAlvo,
		Seguro: false,
	},
	{
		URI:    "/v4/listaEnvio/inverteSmsAtivoByIdAlvo",
		Metodo: http.MethodPost,
		Funcao: inverteSmsAtivoByIdAlvo,
		Seguro: false,
	},
}
