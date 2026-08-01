package clienteV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/cliente/logar",
		Metodo: http.MethodPost,
		Funcao: logar,
		Seguro: false,
	},
	{
		URI:    "/v4/cliente/webLogar",
		Metodo: http.MethodPost,
		Funcao: webLogar,
		Seguro: false,
	},
	{
		URI:    "/v4/cliente/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/getDadosByName",
		Metodo: http.MethodPost,
		Funcao: getDadosByName,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/getDadosByEmail",
		Metodo: http.MethodPost,
		Funcao: getDadosByEmail,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/alteraById",
		Metodo: http.MethodPost,
		Funcao: alteraById,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/deleteById",
		Metodo: http.MethodPost,
		Funcao: deleteById,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/preDeleteById",
		Metodo: http.MethodPost,
		Funcao: preDeleteById,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/restauraPreDeleteById",
		Metodo: http.MethodPost,
		Funcao: RestauraPreDeleteById,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/deleteAllByVinculo",
		Metodo: http.MethodPost,
		Funcao: deleteAllByVinculo,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/listarByIdFranqueado",
		Metodo: http.MethodPost,
		Funcao: listarByIdFranqueado,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/listarComDispByIdFranqueado",
		Metodo: http.MethodPost,
		Funcao: listarComDispByIdFranqueado,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/listarByIdFranqueadoAndNome",
		Metodo: http.MethodPost,
		Funcao: listarByIdFranqueadoAndNome,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/listarByIdFranqueado",
		Metodo: http.MethodPost,
		Funcao: listarByIdFranqueado,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/getAtivoById",
		Metodo: http.MethodPost,
		Funcao: getAtivoById,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/setAtivoById",
		Metodo: http.MethodPost,
		Funcao: setAtivoById,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/inverterAtivoById",
		Metodo: http.MethodPost,
		Funcao: inverterAtivoById,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/getEmailAtivoById",
		Metodo: http.MethodPost,
		Funcao: getEmailAtivoById,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/setEmailAtivoById",
		Metodo: http.MethodPost,
		Funcao: setEmailAtivoById,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/inverterEmailAtivoById",
		Metodo: http.MethodPost,
		Funcao: inverterEmailAtivoById,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/getSmsAtivoById",
		Metodo: http.MethodPost,
		Funcao: getSmsAtivoById,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/setSmsAtivoById",
		Metodo: http.MethodPost,
		Funcao: setSmsAtivoById,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/inverterSmsAtivoById",
		Metodo: http.MethodPost,
		Funcao: inverterSmsAtivoById,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/resetarSenha",
		Metodo: http.MethodPost,
		Funcao: resetarSenha,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/alterarSenhaById",
		Metodo: http.MethodPost,
		Funcao: alterarSenhaById,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/getEmail1LivreByEmail1",
		Metodo: http.MethodPost,
		Funcao: getEmail1LivreByEmail1,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/setEmail1ById",
		Metodo: http.MethodPost,
		Funcao: setEmail1ById,
		Seguro: true,
	},
	{
		URI:    "/v4/cliente/setDispPadraoBtnPanico",
		Metodo: http.MethodPost,
		Funcao: setDispPadraoBtnPanico,
		Seguro: true,
	},
}
