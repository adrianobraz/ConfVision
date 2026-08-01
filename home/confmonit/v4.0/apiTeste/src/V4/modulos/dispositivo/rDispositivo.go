package dispositivoV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/dispositivo/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/getDadosByIdFisico1",
		Metodo: http.MethodPost,
		Funcao: getDadosByIdFisico1,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/getDadosByIdFisico2",
		Metodo: http.MethodPost,
		Funcao: getDadosByIdFisico2,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/getMsgAtendenteById",
		Metodo: http.MethodPost,
		Funcao: getMsgAtendenteById,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/setMsgAtendenteById",
		Metodo: http.MethodPost,
		Funcao: setMsgAtendenteById,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/getArmadoById",
		Metodo: http.MethodPost,
		Funcao: getArmadoById,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/workerGetArmadoById",
		Metodo: http.MethodPost,
		Funcao: workerGetArmadoById,
		// Leitura S/N para o worker ConfVision (sem JWT de sessão).
		Seguro: false,
	},
	{
		URI:    "/v4/dispositivo/setArmadoById",
		Metodo: http.MethodPost,
		Funcao: setArmadoById,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/inverteArmadoById",
		Metodo: http.MethodPost,
		Funcao: inverteArmadoById,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/getAtivoById",
		Metodo: http.MethodPost,
		Funcao: getAtivoById,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/setAtivoById",
		Metodo: http.MethodPost,
		Funcao: setAtivoById,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/inverteAtivoById",
		Metodo: http.MethodPost,
		Funcao: inverteAtivoById,
		Seguro: true,
	},

	{
		URI:    "/v4/dispositivo/gravarUltimoEventoById",
		Metodo: http.MethodPost,
		Funcao: gravarUltimoEventoById,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/alteraById",
		Metodo: http.MethodPost,
		Funcao: alteraById,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/deletaById",
		Metodo: http.MethodPost,
		Funcao: deletaById,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/getNomeClienteById",
		Metodo: http.MethodPost,
		Funcao: getNomeClienteById,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/getIdClienteById",
		Metodo: http.MethodPost,
		Funcao: getIdClienteById,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/getIdFranqueadoById",
		Metodo: http.MethodPost,
		Funcao: getIdFranqueadoById,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/getIdRepresentanteById",
		Metodo: http.MethodPost,
		Funcao: getIdRepresentanteById,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/listarSemComunicacaoByIdFranqueado",
		Metodo: http.MethodPost,
		Funcao: listarSemComunicacaoByIdFranqueado,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/listarByIdCliente",
		Metodo: http.MethodPost,
		Funcao: listarByIdCliente,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/listarByIdFranqueado",
		Metodo: http.MethodPost,
		Funcao: listarByIdFranqueado,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/gerarContaByIdFranqueado",
		Metodo: http.MethodPost,
		Funcao: gerarContaByIdFranqueado,
		Seguro: true,
	},
	{
		URI:    "/v4/dispositivo/vericaContaByIdFranquado",
		Metodo: http.MethodPost,
		Funcao: vericaContaByIdFranquado,
		Seguro: true,
	},
}
