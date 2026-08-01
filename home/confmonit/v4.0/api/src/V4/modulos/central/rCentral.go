package centralV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/central/logar",
		Metodo: http.MethodPost,
		Funcao: logar,
		Seguro: false,
	},
	{
		URI:    "/v4/central/logarAdministrador",
		Metodo: http.MethodPost,
		Funcao: logarAdministrador,
		Seguro: false,
	},
	{
		URI:    "/v4/central/lista",
		Metodo: http.MethodPost,
		Funcao: lista,
		Seguro: false,
	},
	{
		URI:    "/v4/central/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/central/atualiza",
		Metodo: http.MethodPost,
		Funcao: atualiza,
		Seguro: false,
	},
}
