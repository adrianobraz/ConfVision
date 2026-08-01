package menuAtendimento

import (
	"franqueadopro/src/auxiliar"
	"net/http"
)

var RotasMenuAtendimento = []auxiliar.Rota{
	{
		URI:    "/carregar-menu-atendimento",
		Metodo: http.MethodGet,
		Funcao: CarregarMenuAtendimento,
		Aberto: false,
	},
}
