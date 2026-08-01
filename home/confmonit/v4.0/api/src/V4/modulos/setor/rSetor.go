package setorV4

import (
	tiposV4 "api/src/V4/tipos"
	"net/http"
)

var Rotas = []tiposV4.Rota{
	{
		URI:    "/v4/setor/insere",
		Metodo: http.MethodPost,
		Funcao: insere,
		Seguro: false,
	},
	{
		URI:    "/v4/setor/getDadosById",
		Metodo: http.MethodPost,
		Funcao: getDadosById,
		Seguro: false,
	},
	{
		URI:    "/v4/setor/alteraById",
		Metodo: http.MethodPost,
		Funcao: alteraById,
		Seguro: false,
	},
	{
		URI:    "/v4/setor/deletaById",
		Metodo: http.MethodPost,
		Funcao: deletaById,
		Seguro: false,
	},
	{
		URI:    "/v4/setor/deletaAllByDispositivo",
		Metodo: http.MethodPost,
		Funcao: deletaAllByDispositivo,
		Seguro: false,
	},
	{
		URI:    "/v4/setor/getCameraAtivaById",
		Metodo: http.MethodPost,
		Funcao: getCameraAtivaById,
		Seguro: false,
	},
	{
		URI:    "/v4/setor/setCameraAtivaById",
		Metodo: http.MethodPost,
		Funcao: setCameraAtivaById,
		Seguro: false,
	},
	{
		URI:    "/v4/setor/inverteCameraAtivaById",
		Metodo: http.MethodPost,
		Funcao: inverteCameraAtivaById,
		Seguro: false,
	},
	{
		URI:    "/v4/setor/getSetorAtivaById",
		Metodo: http.MethodPost,
		Funcao: getSetorAtivaById,
		Seguro: false,
	},
	{
		URI:    "/v4/setor/setSetorAtivaById",
		Metodo: http.MethodPost,
		Funcao: setSetorAtivaById,
		Seguro: false,
	},
	{
		URI:    "/v4/setor/inverteSetorAtivaById",
		Metodo: http.MethodPost,
		Funcao: inverteSetorAtivaById,
		Seguro: false,
	},
	{
		URI:    "/v4/setor/listaByIdDispositivo",
		Metodo: http.MethodPost,
		Funcao: listaByIdDispositivo,
		Seguro: false,
	},
}
