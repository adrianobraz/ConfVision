package terminalV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/terminal/bloqueiaProcessoById",
		Metodo: http.MethodPost,
		Funcao: bloqueiaProcessoById,
		Seguro: false,
	},
	{
		URI:    "/v4/terminal/liberaProcessoById",
		Metodo: http.MethodPost,
		Funcao: liberaProcessoById,
		Seguro: false,
	},
	{
		URI:    "/v4/terminal/listarProcessos",
		Metodo: http.MethodPost,
		Funcao: listarProcessos,
		Seguro: false,
	},
	{
		URI:    "/v4/terminal/listarProcessosByIdCliente",
		Metodo: http.MethodPost,
		Funcao: listarProcessosByIdCliente,
		Seguro: false,
	},
	{
		URI:    "/v4/terminal/listarProcessoToOpenByCliente",
		Metodo: http.MethodPost,
		Funcao: listarProcessoToOpenByCliente,
		Seguro: false,
	},
	{
		URI:    "/v4/terminal/getDadosProcessoById",
		Metodo: http.MethodPost,
		Funcao: getDadosProcessoById,
		Seguro: false,
	},
	{
		URI:    "/v4/terminal/listarEventosByProcesso",
		Metodo: http.MethodPost,
		Funcao: listarEventosByProcesso,
		Seguro: false,
	},
	{
		URI:    "/v4/terminal/finalizarProcesso",
		Metodo: http.MethodPost,
		Funcao: finalizarProcesso,
		Seguro: false,
	},
	// {
	// 	URI:    "/v4/terminal/",
	// 	Metodo: http.MethodPost,
	// 	Funcao: ,
	// 	Seguro: false,
	// },
	// {
	// 	URI:    "/v4/terminal/",
	// 	Metodo: http.MethodPost,
	// 	Funcao: ,
	// 	Seguro: false,
	// },
	// {
	// 	URI:    "/v4/terminal/",
	// 	Metodo: http.MethodPost,
	// 	Funcao: ,
	// 	Seguro: false,
	// },
}
