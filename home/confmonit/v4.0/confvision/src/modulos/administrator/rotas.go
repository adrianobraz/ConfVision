package administrator

import (
	"confvision/src/auxiliar"
	"net/http"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/administrator",
		Metodo: http.MethodGet,
		Funcao: CarregarLogin,
		Aberto: true,
	},
	{
		URI:    "/administrator/login",
		Metodo: http.MethodPost,
		Funcao: Login,
		Aberto: true,
	},
	{
		URI:    "/administrator/rtmp-falhas",
		Metodo: http.MethodGet,
		Funcao: CarregarRtmpFalhas,
		Aberto: false,
	},
	{
		URI:    "/api/administrator/cameras",
		Metodo: http.MethodGet,
		Funcao: ApiListarCameras,
		Aberto: false,
	},
}
