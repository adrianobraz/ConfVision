package listaBoqueioV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/listaBloqueio/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/listaBloqueio/deleteById",
		Metodo: http.MethodPost,
		Funcao: deleteById,
		Seguro: false,
	},
	{
		URI:    "/v4/listaBloqueio/getBloquadoByIdAlvo",
		Metodo: http.MethodPost,
		Funcao: getBloquadoByIdAlvo,
		Seguro: false,
	},
	{
		URI:    "/v4/listaBloqueio/liberarProvisorio",
		Metodo: http.MethodPost,
		Funcao: liberarProvisorio,
		Seguro: false,
	},
	{
		URI:    "/v4/listaBloqueio/lista",
		Metodo: http.MethodPost,
		Funcao: lista,
		Seguro: false,
	},
}
