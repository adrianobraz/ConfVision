package menuRelatorio

import (
	"franqueadopro/src/auxiliar"
	"net/http"
)

var RotasMenuRelatorio = []auxiliar.Rota{
	{
		URI:    "/carregar-menu-relatorio",
		Metodo: http.MethodGet,
		Funcao: CarregarMenuRelatorio,
		Aberto: false,
	},
	{
		URI:    "/RelatorioLigacoesCarregarClientes",
		Metodo: http.MethodPost,
		Funcao: RelatorioLigacoesCarregarClientes,
		Aberto: false,
	},
}
