package relatorioLigacoes

import (
	"webFranqueado/src/auxiliar"
	"net/http"
)

var RotasRelatorioLigacoes = []auxiliar.Rota{
	{
		URI:    "/carregar-relatorio-ligacoes",
		Metodo: http.MethodGet,
		Funcao: CarregarRelatorioLigacoes,
		Aberto: false,
	},
	{
		URI:    "/custo-atendimento-listar",
		Metodo: http.MethodPost,
		Funcao: CustoAtendimentoListar,
		Aberto: false,
	},
}
