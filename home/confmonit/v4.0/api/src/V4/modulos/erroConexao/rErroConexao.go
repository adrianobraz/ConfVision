package erroConexao

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/erroConexao/listar",
		Metodo: http.MethodPost,
		Funcao: listar,
		Seguro: true,
	},
	{
		URI:    "/v4/erroConexao/limpar",
		Metodo: http.MethodPost,
		Funcao: limpar,
		Seguro: true,
	},
}
