package faturaV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/fatura/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/fatura/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: false,
	},
	{
		URI:    "/v4/fatura/getStatusById",
		Metodo: http.MethodPost,
		Funcao: getStatusById,
		Seguro: false,
	},
	{
		URI:    "/v4/fatura/setStatusById",
		Metodo: http.MethodPost,
		Funcao: setStatusById,
		Seguro: false,
	},
	{
		URI:    "/v4/fatura/getValorById",
		Metodo: http.MethodPost,
		Funcao: getValorById,
		Seguro: false,
	},
	{
		URI:    "/v4/fatura/setValorById",
		Metodo: http.MethodPost,
		Funcao: setValorById,
		Seguro: false,
	},
	{
		URI:    "/v4/fatura/addValorById",
		Metodo: http.MethodPost,
		Funcao: addValorById,
		Seguro: false,
	},
	{
		URI:    "/v4/fatura/subValorById",
		Metodo: http.MethodPost,
		Funcao: subValorById,
		Seguro: false,
	},
	{
		URI:    "/v4/fatura/deleteById",
		Metodo: http.MethodPost,
		Funcao: deleteById,
		Seguro: false,
	},
	{
		URI:    "/v4/fatura/deleteAllByIdOrigem",
		Metodo: http.MethodPost,
		Funcao: deleteAllByIdOrigem,
		Seguro: false,
	},
	{
		URI:    "/v4/fatura/deleteAllByIdDestino",
		Metodo: http.MethodPost,
		Funcao: deleteAllByIdDestino,
		Seguro: false,
	},
	{
		URI:    "/v4/fatura/listaByIdOrigem",
		Metodo: http.MethodPost,
		Funcao: listaByIdOrigem,
		Seguro: false,
	},
	{
		URI:    "/v4/fatura/listaPagoByIdOrigem",
		Metodo: http.MethodPost,
		Funcao: listaPagoByIdOrigem,
		Seguro: false,
	},
	{
		URI:    "/v4/fatura/listaByIdDestino",
		Metodo: http.MethodPost,
		Funcao: listaByIdDestino,
		Seguro: false,
	},
	{
		URI:    "/v4/fatura/listaPagoByIdDestino",
		Metodo: http.MethodPost,
		Funcao: listaPagoByIdDestino,
		Seguro: false,
	},
	{
		URI:    "/v4/faturaIntem/itemInsere",
		Metodo: http.MethodPost,
		Funcao: itemInsere,
		Seguro: false,
	},
	{
		URI:    "/v4/faturaIntem/itemGetDadosById",
		Metodo: http.MethodPost,
		Funcao: itemGetDadosById,
		Seguro: false,
	},
	{
		URI:    "/v4/faturaIntem/itemAlteraById",
		Metodo: http.MethodPost,
		Funcao: itemAlteraById,
		Seguro: false,
	},
	{
		URI:    "/v4/faturaIntem/itemDeleteById",
		Metodo: http.MethodPost,
		Funcao: itemDeleteById,
		Seguro: false,
	},
	{
		URI:    "/v4/faturaIntem/itemDeleteAllByIdFatura",
		Metodo: http.MethodPost,
		Funcao: itemDeleteAllByIdFatura,
		Seguro: false,
	},
	{
		URI:    "/v4/faturaIntem/itemListaByIdFatura",
		Metodo: http.MethodPost,
		Funcao: itemListaByIdFatura,
		Seguro: false,
	},
}
