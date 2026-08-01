package ticketV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/ticket/insere",
		Metodo: http.MethodPost,
		Funcao: Insere,
		Seguro: true,
	},
	{
		URI:    "/v4/ticket/getDadosById",
		Metodo: http.MethodPost,
		Funcao: GetDadosById,
		Seguro: true,
	},
	{
		URI:    "/v4/ticket/alteraById",
		Metodo: http.MethodPost,
		Funcao: AlteraById,
		Seguro: true,
	},
	{
		URI:    "/v4/ticket/deletaById",
		Metodo: http.MethodPost,
		Funcao: DeletaById,
		Seguro: true,
	},
	{
		URI:    "/v4/ticket/deletaAllByMster",
		Metodo: http.MethodPost,
		Funcao: DeletaAllByMster,
		Seguro: true,
	},
	{
		URI:    "/v4/ticket/deletaAllBySlave",
		Metodo: http.MethodPost,
		Funcao: DeletaAllBySlave,
		Seguro: true,
	},
	{
		URI:    "/v4/ticket/listarByIdMaster",
		Metodo: http.MethodPost,
		Funcao: ListarByIdMaster,
		Seguro: true,
	},
	{
		URI:    "/v4/ticket/listarByIdSlave",
		Metodo: http.MethodPost,
		Funcao: ListarByIdSlave,
		Seguro: true,
	},
}
