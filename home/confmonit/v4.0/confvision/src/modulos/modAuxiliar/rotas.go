package modAuxiliar

import (
	"confvision/src/auxiliar"
	"net/http"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregarClientes",
		Metodo: http.MethodPost,
		Funcao: carregarClientes,
		Aberto: false,
	},
	{
		URI:    "/carregarDispositivos",
		Metodo: http.MethodPost,
		Funcao: carregarDispositivos,
		Aberto: false,
	},
	{
		URI:    "/auxiliar-carregar-operadores",
		Metodo: http.MethodPost,
		Funcao: AxiliarCarregarOperadores,
		Aberto: false,
	},
	{
		URI:    "/auxiliar-carregar-dispositivos",
		Metodo: http.MethodPost,
		Funcao: AxiliarCarregarDosítivos,
		Aberto: false,
	},
	{
		URI:    "/GerenciarDispositivoFabricantesListar",
		Metodo: http.MethodPost,
		Funcao: GerenciarDispositivoFabricantesListar,
		Aberto: false,
	},
}
