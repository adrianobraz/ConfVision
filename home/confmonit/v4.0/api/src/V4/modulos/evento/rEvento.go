package eventoV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/evento/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/evento/listarByIdProcesso",
		Metodo: http.MethodPost,
		Funcao: listaByIdProcesso,
		Seguro: false,
	},
	{
		URI:    "/v4/evento/listarAgrupadoByIdProcesso",
		Metodo: http.MethodPost,
		Funcao: listaAgrupadoByIdProcesso,
		Seguro: false,
	},
	{
		URI:    "/v4/evento/listarByDispStartEnd",
		Metodo: http.MethodPost,
		Funcao: listarByDispStartEnd,
		Seguro: false,
	},
	{
		URI:    "/v4/evento/listarByDispStartEndGrupo",
		Metodo: http.MethodPost,
		Funcao: listarByDispStartEndGrupo,
		Seguro: false,
	},
	{
		URI:    "/v4/evento/listarByDispProcOpen",
		Metodo: http.MethodPost,
		Funcao: listarByDispProcOpen,
		Seguro: false,
	},
	{
		URI:    "/v4/evento/contarByIdFranqueadoPeriodo",
		Metodo: http.MethodPost,
		Funcao: contarByIdFranqueadoPeriodo,
		Seguro: true,
	},
	{
		URI:    "/v4/evento/contarByIdFranqueadoGrupo",
		Metodo: http.MethodPost,
		Funcao: contarByIdFranqueadoGrupo,
		Seguro: true,
	},
	{
		URI:    "/v4/evento/listarByIdFranqueadoStartEndGrupo",
		Metodo: http.MethodPost,
		Funcao: listarByIdFranqueadoStartEndGrupo,
		Seguro: true,
	},
}
