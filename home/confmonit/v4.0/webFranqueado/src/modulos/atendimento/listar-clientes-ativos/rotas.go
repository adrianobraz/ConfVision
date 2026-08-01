package listarClientesAtivos

import (
	"webFranqueado/src/auxiliar"
	"net/http"
)

var RotasListarClientesAtivos = []auxiliar.Rota{
	{
		URI:    "/carregar-listar-clientes-ativos",
		Metodo: http.MethodGet,
		Funcao: CarregarListarClientesAtivos,
		Aberto: false,
	},
	{
		URI:    "/ListarClientesAtivos",
		Metodo: http.MethodPost,
		Funcao: ListarClientesAtivos,
		Aberto: false,
	},
}
