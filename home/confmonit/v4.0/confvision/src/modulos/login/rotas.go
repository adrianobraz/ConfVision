package login

import (
	"confvision/src/auxiliar"
	"net/http"
)

var Rotas = []auxiliar.Rota{
	// Carrega a pagina de login
	{
		URI:    "/",
		Metodo: http.MethodGet,
		Funcao: CarregarLogin,
		Aberto: true,
	},
	// Carrega a pagina de login
	{
		URI:    "/login",
		Metodo: http.MethodGet,
		Funcao: CarregarLogin,
		Aberto: true,
	},
	// Efetua o login no sistema
	{
		URI:    "/LoginLogar",
		Metodo: http.MethodPost,
		Funcao: LoginLogar,
		Aberto: true,
	},
	{
		URI:    "/carregar-dados",
		Metodo: http.MethodPost,
		Funcao: CarregarDados,
		Aberto: true,
	},
	{
		URI:    "/logout",
		Metodo: http.MethodGet,
		Funcao: Logout,
		Aberto: false,
	},
	{
		URI:    "/getFranqDadosById",
		Metodo: http.MethodPost,
		Funcao: getFranqDadosById,
		Aberto: false,
	},
	{
		URI:    "/api/franqueado/usa-confvision",
		Metodo: http.MethodPost,
		Funcao: apiUsaConfVision,
		Aberto: false,
	},
}
