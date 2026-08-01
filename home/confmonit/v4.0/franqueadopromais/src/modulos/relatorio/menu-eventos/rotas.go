package menuEventos

import (
	"franqueadopro/src/auxiliar"
	"net/http"
)

var RotasMenuEventos = []auxiliar.Rota{
	{
		URI:    "/carregar-menu-eventos",
		Metodo: http.MethodGet,
		Funcao: CarregarMenuEventos,
		Aberto: false,
	},
}
