package listarClientesInativos

import (
	"franqueadopro/src/auxiliar"
	"net/http"
)

var RotasListarClientesAtivos = []auxiliar.Rota{
	{
		URI:    "/carregar-listar-clientes-inativos",
		Metodo: http.MethodGet,
		Funcao: CarregarGerenciarClientesInativos,
		Aberto: false,
	},
	{
		URI:    "/ListarClientesInativos",
		Metodo: http.MethodPost,
		Funcao: ListarClientesInativos,
		Aberto: false,
	},
}
