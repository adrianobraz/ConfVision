package relatorioAtendimento

import (
	"franqueadopro/src/auxiliar"
	"net/http"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregar-relatorio-atendimento",
		Metodo: http.MethodGet,
		Funcao: CarregarRelatorioAtendimento,
		Aberto: false,
	},	

	{
		URI:    "/relatorioAtendimentoListar",
		Metodo: http.MethodPost,
		Funcao: relatorioAtendimentoListar,
		Aberto: false,
	},
}
