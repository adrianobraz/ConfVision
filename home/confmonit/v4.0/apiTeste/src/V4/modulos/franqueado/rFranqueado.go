package franqueadoV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

// ajustar para banco novo
var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/franqueado/logar",
		Metodo: http.MethodPost,
		Funcao: logar,
		Seguro: false,
	},
	{
		URI:    "/v4/franqueado/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/franqueado/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: false,
	},

	{
		URI:    "/v4/franqueado/getNomeById",
		Metodo: http.MethodPost,
		Funcao: getNomeById,
		Seguro: false,
	},
	{
		URI:    "/v4/franqueado/alterarById",
		Metodo: http.MethodPost,
		Funcao: alterarById,
		Seguro: false,
	},
	{
		URI:    "/v4/franqueado/deletaById",
		Metodo: http.MethodPost,
		Funcao: deletaById,
		Seguro: false,
	},
	{
		URI:    "/v4/franqueado/deletaAllByRepresentante",
		Metodo: http.MethodPost,
		Funcao: deletaAllByRepresentante,
		Seguro: false,
	},
	{
		URI:    "/v4/franqueado/listarByIdRepresentante",
		Metodo: http.MethodPost,
		Funcao: listarByIdRepresentante,
		Seguro: false,
	},
	{
		URI:    "/v4/franqueado/getAtivoById",
		Metodo: http.MethodPost,
		Funcao: getAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/franqueado/setAtivoById",
		Metodo: http.MethodPost,
		Funcao: setAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/franqueado/inverterAtivoById",
		Metodo: http.MethodPost,
		Funcao: inverterAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/franqueado/getEmailAtivoById",
		Metodo: http.MethodPost,
		Funcao: getEmailAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/franqueado/setEmailAtivoById",
		Metodo: http.MethodPost,
		Funcao: setEmailAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/franqueado/InverterEmailAtivoById",
		Metodo: http.MethodPost,
		Funcao: inverterEmailAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/franqueado/getSmsAtivoById",
		Metodo: http.MethodPost,
		Funcao: getSmsAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/franqueado/setSmsAtivoById",
		Metodo: http.MethodPost,
		Funcao: setSmsAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/franqueado/inverterSmsAtivoById",
		Metodo: http.MethodPost,
		Funcao: inverterSmsAtivoById,
		Seguro: false,
	},
	{
		URI:    "/v4/franqueado/cancelaById",
		Metodo: http.MethodPost,
		Funcao: cancelaById,
		Seguro: false,
	},
	{
		URI:    "/v4/franqueado/reverteCancelaById",
		Metodo: http.MethodPost,
		Funcao: reverteCancelaById,
		Seguro: false,
	},
}
