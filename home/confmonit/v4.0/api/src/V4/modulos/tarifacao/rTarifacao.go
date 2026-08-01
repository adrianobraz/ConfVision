package tarifacaoV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/tarifacao/insereLancamento",
		Metodo: http.MethodPost,
		Funcao: insereLancamento,
		Seguro: false,
	},
	{
		URI:    "/v4/tarifacao/listaLancamentosPendentes",
		Metodo: http.MethodPost,
		Funcao: listaLancamentosPendentes,
		Seguro: false,
	},
	{
		URI:    "/v4/tarifacao/deleteLancamento",
		Metodo: http.MethodPost,
		Funcao: deleteLancamento,
		Seguro: false,
	},
}
