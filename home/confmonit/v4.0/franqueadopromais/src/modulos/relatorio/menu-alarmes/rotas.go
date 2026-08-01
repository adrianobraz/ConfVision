package menuAlarmes

import (
	"franqueadopro/src/auxiliar"
	"net/http"
)

var RotasMenuAlarmes = []auxiliar.Rota{
	{
		URI:    "/carregar-menu-alarmes",
		Metodo: http.MethodGet,
		Funcao: CarregarMenuAlarmes,
		Aberto: false,
	},
}
