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
		Seguro: false,
	},
	{
		URI:    "/v4/processo/getIdProcessoByIdDispositivo",
		Metodo: http.MethodPost,
		Funcao: getIdProcessoByIdDispositivo,
		Seguro: false,
	},
	{
		URI:    "/v4/processo/getIdAtendentelById",
		Metodo: http.MethodPost,
		Funcao: getIdAtendentelById,
		Seguro: false,
	},
	{
		URI:    "/v4/processo/setIdAtendenteById",
		Metodo: http.MethodPost,
		Funcao: setIdAtendenteById,
		Seguro: false,
	},
	{
		URI:    "/v4/processo/getNivelById",
		Metodo: http.MethodPost,
		Funcao: getNivelById,
		Seguro: false,
	},
	{
		URI:    "/v4/processo/setNivelById",
		Metodo: http.MethodPost,
		Funcao: setNivelById,
		Seguro: false,
	},
	{
		URI:    "/v4/processo/atualizaNivelById",
		Metodo: http.MethodPost,
		Funcao: atualizaNivelById,
		Seguro: false,
	},
	{
		URI:    "/v4/processo/getMsgAtendenteById",
		Metodo: http.MethodPost,
		Funcao: getMsgAtendenteById,
		Seguro: false,
	},
	{
		URI:    "/v4/processo/setMsgAtendenteById",
		Metodo: http.MethodPost,
		Funcao: setMsgAtendenteById,
		Seguro: false,
	},
	{
		URI:    "/v4/processo/ListarToOpen",
		Metodo: http.MethodPost,
		Funcao: ListarToOpen,
		Seguro: false,
	},
	{
		URI:    "/v4/processo/ListarToOpenByIdDispositivo",
		Metodo: http.MethodPost,
		Funcao: ListarToOpenByIdDispositivo,
		Seguro: false,
	},
	{
		URI:    "/v4/processo/ListarToOpenByIdCliente",
		Metodo: http.MethodPost,
		Funcao: ListarToOpenByIdCliente,
		Seguro: false,
	},
	{
		URI:    "/v4/processo/listarByFiltro",
		Metodo: http.MethodPost,
		Funcao: listarByFiltro,
		Seguro: false,
	},
	{
		URI:    "/v4/processo/listarEventosByFiltro",
		Metodo: http.MethodPost,
		Funcao: listarEventosByFiltro,
		Seguro: false,
	},
}
