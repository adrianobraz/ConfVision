package contactidV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rota = []tiposV4.Rota{
	{
		URI:    "/v4/contactid/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/contactid/getDadosByCodigo",
		Metodo: http.MethodPost,
		Funcao: getDadosByCodigo,
		Seguro: false,
	},
	{
		URI:    "/v4/contactid/getDadosById",
		Metodo: http.MethodPost,
		Funcao: GetDadosById,
		Seguro: false,
	},
	{
		URI:    "/v4/contactid/alteraById",
		Metodo: http.MethodPost,
		Funcao: alteraById,
		Seguro: false,
	},
	{
		URI:    "/v4/contactid/deletaById",
		Metodo: http.MethodPost,
		Funcao: DeletaById,
		Seguro: false,
	},
	{
		URI:    "/v4/contactid/listaPadrao",
		Metodo: http.MethodPost,
		Funcao: listaPadrao,
		Seguro: false,
	},
	{
		URI:    "/v4/contactid/listaByGrupo",
		Metodo: http.MethodPost,
		Funcao: listaByGrupo,
		Seguro: false,
	},
	{
		URI:    "/v4/contactid/listaByIdVinculo",
		Metodo: http.MethodPost,
		Funcao: listaByIdVinculo,
		Seguro: false,
	},
	{
		URI:    "/v4/contactid/listaGrupos",
		Metodo: http.MethodPost,
		Funcao: listaGrupos,
		Seguro: false,
	},
	{
		URI:    "/v4/contactid/getNivelById",
		Metodo: http.MethodPost,
		Funcao: getNivelById,
		Seguro: false,
	},
}
