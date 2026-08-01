package custoAtendimentoV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/custoAtendimento/listarByFiltro",
		Metodo: http.MethodPost,
		Funcao: listarByFiltro,
		Seguro: true,
	},
	// Aliases da API legada usados pelo FranqueadoPro / webFranqueado
	{
		URI:    "/CustoAtendimentoListar",
		Metodo: http.MethodPost,
		Funcao: listarByFiltro,
		Seguro: true,
	},
	{
		URI:    "/custo-atendimento-listar",
		Metodo: http.MethodPost,
		Funcao: listarByFiltro,
		Seguro: true,
	},
}
