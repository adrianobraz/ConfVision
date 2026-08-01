package processoV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/processo/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: true,
	},
	{
		URI:    "/v4/processo/getIdProcessoByIdDispositivo",
		Metodo: http.MethodPost,
		Funcao: getIdProcessoByIdDispositivo,
		Seguro: true,
	},
	{
		URI:    "/v4/processo/getIdAtendentelById",
		Metodo: http.MethodPost,
		Funcao: getIdAtendentelById,
		Seguro: true,
	},
	{
		URI:    "/v4/processo/setIdAtendenteById",
		Metodo: http.MethodPost,
		Funcao: setIdAtendenteById,
		Seguro: true,
	},
	{
		URI:    "/v4/processo/getNivelById",
		Metodo: http.MethodPost,
		Funcao: getNivelById,
		Seguro: true,
	},
	{
		URI:    "/v4/processo/setNivelById",
		Metodo: http.MethodPost,
		Funcao: setNivelById,
		Seguro: true,
	},
	{
		URI:    "/v4/processo/atualizaNivelById",
		Metodo: http.MethodPost,
		Funcao: atualizaNivelById,
		Seguro: true,
	},
	{
		URI:    "/v4/processo/getMsgAtendenteById",
		Metodo: http.MethodPost,
		Funcao: getMsgAtendenteById,
		Seguro: true,
	},
	{
		URI:    "/v4/processo/setMsgAtendenteById",
		Metodo: http.MethodPost,
		Funcao: setMsgAtendenteById,
		Seguro: true,
	},
	{
		URI:    "/v4/processo/ListarToOpen",
		Metodo: http.MethodPost,
		Funcao: ListarToOpen,
		Seguro: true,
	},
	{
		URI:    "/v4/processo/ListarToOpenByIdDispositivo",
		Metodo: http.MethodPost,
		Funcao: ListarToOpenByIdDispositivo,
		Seguro: true,
	},
	{
		URI:    "/v4/processo/ListarToOpenByIdCliente",
		Metodo: http.MethodPost,
		Funcao: ListarToOpenByIdCliente,
		Seguro: true,
	},
	{
		URI:    "/v4/processo/listarByFiltro",
		Metodo: http.MethodPost,
		Funcao: listarByFiltro,
		Seguro: true,
	},
	{
		URI:    "/v4/processo/listarEventosByFiltro",
		Metodo: http.MethodPost,
		Funcao: listarEventosByFiltro,
		Seguro: true,
	},
	{
		URI:    "/v4/processo/finalizarSistema",
		Metodo: http.MethodPost,
		Funcao: finalizarSistema,
		Seguro: true,
	},
}
