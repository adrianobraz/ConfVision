package listarClientes

import (
	"webFranqueado/src/auxiliar"
	"net/http"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregar-listar-clientes",
		Metodo: http.MethodGet,
		Funcao: CarregarListarClientes,
		Aberto: false,
	},

	{
		URI:    "/RelatorioLigacoesListar",
		Metodo: http.MethodPost,
		Funcao: RelatorioLigacoesListar,
		Aberto: false,
	},
	{
		URI:    "/ListarClientesCarregarClientes",
		Metodo: http.MethodPost,
		Funcao: ListarClientesCarregarClientes,
		Aberto: false,
	},
}
