package notificacao

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/notificacao/criar",
		Metodo: http.MethodPost,
		Funcao: criar,
		Seguro: false,
	},
	{
		URI:    "/v4/notificacao/listar",
		Metodo: http.MethodPost,
		Funcao: listar,
		Seguro: false,
	},
	{
		URI:    "/v4/notificacao/desativar",
		Metodo: http.MethodPost,
		Funcao: desativar,
		Seguro: false,
	},
	{
		URI:    "/v4/notificacao/minhas",
		Metodo: http.MethodPost,
		Funcao: minhas,
		Seguro: false,
	},
	{
		URI:    "/v4/notificacao/contagem",
		Metodo: http.MethodPost,
		Funcao: contagem,
		Seguro: false,
	},
	{
		URI:    "/v4/notificacao/marcarVisto",
		Metodo: http.MethodPost,
		Funcao: marcarVisto,
		Seguro: false,
	},
	{
		URI:    "/v4/notificacao/marcarLido",
		Metodo: http.MethodPost,
		Funcao: marcarLido,
		Seguro: false,
	},
}
