package representanteV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/representante/logar",
		Metodo: http.MethodPost,
		Funcao: logar,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/getDadosFullById",
		Metodo: http.MethodPost,
		Funcao: getDadosFullById,
		Seguro: false,
	},

	{
		URI:    "/v4/representante/getEnviarEmailById",
		Metodo: http.MethodPost,
		Funcao: getEnviarEmailById,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/getIdPacoteById",
		Metodo: http.MethodPost,
		Funcao: getIdPacoteById,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/alteraById",
		Metodo: http.MethodPost,
		Funcao: alteraById,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/cancelaById",
		Metodo: http.MethodPost,
		Funcao: cancelaById,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/reverteCancelaById",
		Metodo: http.MethodPost,
		Funcao: reverteCancelaById,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/deletaById",
		Metodo: http.MethodPost,
		Funcao: deletaById,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/getAtivoById",
		Metodo: http.MethodPost,
		Funcao: getAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/setAtivoById",
		Metodo: http.MethodPost,
		Funcao: setAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/inverteAtivoById",
		Metodo: http.MethodPost,
		Funcao: inverteAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/getEmailAtivoById",
		Metodo: http.MethodPost,
		Funcao: getEmailAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/setEmailAtivoById",
		Metodo: http.MethodPost,
		Funcao: setEmailAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/inverteEmailAtivoById",
		Metodo: http.MethodPost,
		Funcao: inverteEmailAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/getSmsAtivoById",
		Metodo: http.MethodPost,
		Funcao: getSmsAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/setSmsAtivoById",
		Metodo: http.MethodPost,
		Funcao: setSmsAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/inverteSmsAtivoById",
		Metodo: http.MethodPost,
		Funcao: inverteSmsAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/setUsaAdmConfmonitById",
		Metodo: http.MethodPost,
		Funcao: setUsaAdmConfmonitById,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/listar",
		Metodo: http.MethodPost,
		Funcao: listar,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/listarCentral",
		Metodo: http.MethodPost,
		Funcao: listarCentral,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/listarCentralPorReferencia",
		Metodo: http.MethodPost,
		Funcao: listarCentralPorReferencia,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/listarFull",
		Metodo: http.MethodPost,
		Funcao: listarFull,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/listarToAtivos",
		Metodo: http.MethodPost,
		Funcao: listarToAtivos,
		Seguro: false,
	},
	{
		URI:    "/v4/representante/listarToDesativado",
		Metodo: http.MethodPost,
		Funcao: listarToDesativado,
		Seguro: false,
	},
}
