package confserviceparceiroV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/confservice/parceiros/global",
		Metodo: http.MethodPost,
		Funcao: parceirosGlobal,
		Seguro: false,
	},
	{
		URI:    "/v4/confservice/parceiros-rep/listar",
		Metodo: http.MethodPost,
		Funcao: parceirosRepListar,
		Seguro: false,
	},
	{
		URI:    "/v4/confservice/parceiros-rep/salvar",
		Metodo: http.MethodPost,
		Funcao: parceirosRepSalvar,
		Seguro: false,
	},
	{
		URI:    "/v4/confservice/parceiros-rep/disponiveis",
		Metodo: http.MethodPost,
		Funcao: parceirosRepDisponiveis,
		Seguro: false,
	},
	{
		URI:    "/v4/confservice/parceiros-franqueado/listar",
		Metodo: http.MethodPost,
		Funcao: parceirosFranqueadoListar,
		Seguro: false,
	},
	{
		URI:    "/v4/confservice/parceiros-franqueado/salvar",
		Metodo: http.MethodPost,
		Funcao: parceirosFranqueadoSalvar,
		Seguro: false,
	},
}
