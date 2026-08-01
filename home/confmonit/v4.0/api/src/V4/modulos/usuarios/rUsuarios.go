package usuariosV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/usuario/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/getDadosByEmail1",
		Metodo: http.MethodPost,
		Funcao: getDadosByEmail1,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/getIdVinculoByEmail1",
		Metodo: http.MethodPost,
		Funcao: getIdVinculoByEmail1,
		Seguro: false,
	},

	{
		URI:    "/v4/usuario/getEmail1Livre",
		Metodo: http.MethodPost,
		Funcao: getEmail1Livre,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/getEmail2Livre",
		Metodo: http.MethodPost,
		Funcao: getEmail2Livre,
		Seguro: false,
	},

	{
		URI:    "/v4/usuario/alteraById",
		Metodo: http.MethodPost,
		Funcao: alteraById,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/deletaById",
		Metodo: http.MethodPost,
		Funcao: deletaById,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/deletaAllByVinculo",
		Metodo: http.MethodPost,
		Funcao: deletaAllByVinculo,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/listar",
		Metodo: http.MethodPost,
		Funcao: listar,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/listarByVinculo",
		Metodo: http.MethodPost,
		Funcao: listarByVinculo,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/listarByVinculoCentral",
		Metodo: http.MethodPost,
		Funcao: listarByVinculoCentral,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/listarByVinculoCentralPorReferencia",
		Metodo: http.MethodPost,
		Funcao: listarByVinculoCentralPorReferencia,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/listarByVinculoToMaster",
		Metodo: http.MethodPost,
		Funcao: listarByVinculoToMaster,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/listarByVinculoToNotMaster",
		Metodo: http.MethodPost,
		Funcao: listarByVinculoToNotMaster,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/listarByVinculoToAtivo",
		Metodo: http.MethodPost,
		Funcao: listarByVinculoToAtivo,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/listarByVinculoToNotAtivo",
		Metodo: http.MethodPost,
		Funcao: listarByVinculoToNotAtivo,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/getUsuarioAtivaById",
		Metodo: http.MethodPost,
		Funcao: getUsuarioAtivaById,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/setUsuarioAtivaById",
		Metodo: http.MethodPost,
		Funcao: setUsuarioAtivaById,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/inverteUsuarioAtivaById",
		Metodo: http.MethodPost,
		Funcao: inverteUsuarioAtivaById,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/getTerminalAtivaById",
		Metodo: http.MethodPost,
		Funcao: getTerminalAtivaById,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/setTerminalAtivaById",
		Metodo: http.MethodPost,
		Funcao: setTerminalAtivaById,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/inverteTerminalAtivaById",
		Metodo: http.MethodPost,
		Funcao: inverteTerminalAtivaById,
		Seguro: false,
	},

	{
		URI:    "/v4/usuario/getWebAtivaById",
		Metodo: http.MethodPost,
		Funcao: getWebAtivaById,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/setWebAtivaById",
		Metodo: http.MethodPost,
		Funcao: setWebAtivaById,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/inverteWebAtivaById",
		Metodo: http.MethodPost,
		Funcao: inverteWebAtivaById,
		Seguro: false,
	},
	// {
	// 	URI:    "/v4/usuario/getTipoById",
	// 	Metodo: http.MethodPost,
	// 	Funcao: getTipoById,
	// 	Seguro: false,
	// },
	// {
	// 	URI:    "/v4/usuario/setTipoById",
	// 	Metodo: http.MethodPost,
	// 	Funcao: setTipoById,
	// 	Seguro: false,
	// },
	// {
	// 	URI:    "/v4/usuario/masterAtiva",
	// 	Metodo: http.MethodPost,
	// 	Funcao: masterAtiva,
	// 	Seguro: false,
	// },
	{
		URI:    "/v4/usuario/validaEmail1",
		Metodo: http.MethodPost,
		Funcao: validaEmail1,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/getAtivarEnviarEmailById",
		Metodo: http.MethodPost,
		Funcao: getAtivarEnviarEmailById,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/setAtivarEnviarEmailById",
		Metodo: http.MethodPost,
		Funcao: setAtivarEnviarEmailById,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/inverteAtivarEnviarEmailById",
		Metodo: http.MethodPost,
		Funcao: inverteAtivarEnviarEmailById,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/resetarSenhaById",
		Metodo: http.MethodPost,
		Funcao: resetarSenhaById,
		Seguro: false,
	},
	{
		URI:    "/v4/usuario/alterarSenhaById",
		Metodo: http.MethodPost,
		Funcao: alterarSenhaById,
		Seguro: false,
	},
}
