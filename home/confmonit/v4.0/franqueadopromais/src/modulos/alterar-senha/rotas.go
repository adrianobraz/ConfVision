package alterarSenha

import (
	"franqueadopro/src/auxiliar"
	"net/http"
)

var RotasAlterarSenha = []auxiliar.Rota{
	{
		URI:    "/carregar-alterar-senha",
		Metodo: http.MethodGet,
		Funcao: CarregarAlterarSenha,
		Aberto: false,
	},
	{
		URI:    "/alterar-senha",
		Metodo: http.MethodPost,
		Funcao: AlterarSenha,
		Aberto: false,
	},
}
