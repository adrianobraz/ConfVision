package listarDispositivoArmados

import (
	"webFranqueado/src/auxiliar"
	"net/http"
)

var RotasListarClienteArmado = []auxiliar.Rota{
	{
		URI:    "/carregar-listar-dispositivos-armados",
		Metodo: http.MethodGet,
		Funcao: CarregarListarClientesArmados,
		Aberto: false,
	},
	{
		URI:    "/dispositivoListarArmado",
		Metodo: http.MethodPost,
		Funcao: dispositivoListarArmado,
		Aberto: false,
	},
}
