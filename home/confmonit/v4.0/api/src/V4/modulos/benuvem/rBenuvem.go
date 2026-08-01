package benuvemV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{

	{
		URI:    "/benuvem/gerarEvento",
		Metodo: http.MethodPost,
		Funcao: gerarEvento,
		Seguro: true,
	},
	{
		URI:    "/benuvem/finalizarEvento",
		Metodo: http.MethodPost,
		Funcao: finalizarEvento,
		Seguro: true,
	},
	{
		URI:    "/benuvem/getRtpsImagens",
		Metodo: http.MethodPost,
		Funcao: getRtpsImagens,
		Seguro: true,
	},
	{
		URI:    "/benuvem/getUrlImagem",
		Metodo: http.MethodPost,
		Funcao: getUrlImagem,
		Seguro: true,
	},
}
