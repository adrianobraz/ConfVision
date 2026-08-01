package listarDispositivoDesarmado

import (
	"net/http"
	"confvision/src/auxiliar"
)

var RotasListarClientesDesarmados = []auxiliar.Rota{
	{
		URI:    "/carregar-listar-dispositivos-desarmados",
		Metodo: http.MethodGet,
		Funcao: CarregarListarClientesDesarmados,
		Aberto: false,
	},
	{
		URI:    "/DispositivoListarDesarmado",
		Metodo: http.MethodPost,
		Funcao: DispositivoListarDesarmados,
		Aberto: false,
	},
}
