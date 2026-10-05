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
	{
		URI:    "/api/administrator/infra-pool/status",
		Metodo: http.MethodGet,
		Funcao: ApiInfraPoolStatus,
		Aberto: false,
	},
	{
		URI:    "/api/administrator/infra-pool/mediamtx-nodes",
		Metodo: http.MethodGet,
		Funcao: ApiInfraMediamtxNodes,
		Aberto: false,
	},
	{
		URI:    "/api/administrator/infra-pool/mediamtx",
		Metodo: http.MethodPost,
		Funcao: ApiInfraPoolMediamtxSave,
		Aberto: false,
	},
	{
		URI:    "/api/administrator/infra-pool/worker",
		Metodo: http.MethodPost,
		Funcao: ApiInfraPoolWorkerSave,
		Aberto: false,
	},
	{
		URI:    "/api/administrator/infra-pool/franqueados",
		Metodo: http.MethodGet,
		Funcao: ApiInfraFranqueadosBusca,
		Aberto: false,
	},
}
