package menuClientes

import (
	"franqueadopro/src/auxiliar"
	"net/http"
)

var RotasMenuClientes = []auxiliar.Rota{
	{
		URI:    "/carregar-menu-clientes",
		Metodo: http.MethodGet,
		Funcao: CarregarMenuClientes,
		Aberto: false,
	},
}
