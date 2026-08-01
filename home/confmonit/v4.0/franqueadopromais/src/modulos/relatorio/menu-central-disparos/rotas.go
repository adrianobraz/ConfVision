package menuCentralDisparos

import (
	"franqueadopro/src/auxiliar"
	"net/http"
)

var RotasMenuCentralDisparos = []auxiliar.Rota{
	{
		URI:    "/carregar-menu-central-disparos",
		Metodo: http.MethodGet,
		Funcao: CarregarMenuCentralDisparos,
		Aberto: false,
	},
}
