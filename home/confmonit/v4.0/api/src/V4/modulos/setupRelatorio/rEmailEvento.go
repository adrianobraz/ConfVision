package setupRelatorioV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/setupRelatorio/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/setupRelatorio/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: false,
	},
	{
		URI:    "/v4/setupRelatorio/getDadosByIdCliente",
		Metodo: http.MethodPost,
		Funcao: getDadosByIdCliente,
		Seguro: false,
	},
	{
		URI:    "/v4/setupRelatorio/alteraById",
		Metodo: http.MethodPost,
		Funcao: alteraById,
		Seguro: false,
	},
	{
		URI:    "/v4/setupRelatorio/deletaById",
		Metodo: http.MethodPost,
		Funcao: deletaById,
		Seguro: false,
	},
	{
		URI:    "/v4/setupRelatorio/deletaByIdCliente",
		Metodo: http.MethodPost,
		Funcao: deletaByIdCliente,
		Seguro: false,
	},
}
