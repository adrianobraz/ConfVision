package visapi

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Rotas espelham paths do Xano confVision — workers e UI via proxyXano.
func RegistrarRotasWorkerAPI(r *mux.Router) {
	routes := []struct {
		path    string
		methods []string
	}{
		{"/vis_health", []string{http.MethodGet}},
		{"/vis_camera_query_ativas", []string{http.MethodGet}},
		{"/vis_camera_sync_ativas", []string{http.MethodGet}},
		{"/vis_camera_area_query_ativas", []string{http.MethodGet}},
		{"/vis_camera_query_gravacao_ativas", []string{http.MethodGet}},
		{"/vis_worker_ping", []string{http.MethodPost}},
		{"/vis_camera_stream_reactivate", []string{http.MethodPost}},
		{"/vis_camera/rtmp_auth/{vis_camera_id}", []string{http.MethodGet}},
		{"/vis_camera_rtmp_auth_sync", []string{http.MethodGet}},
		{"/vis_camera_by_franqueado", []string{http.MethodGet}},
		{"/vis_camera_by_cliente", []string{http.MethodGet}},
		{"/vis_camera_by_setor", []string{http.MethodGet}},
		{"/vis_camera", []string{http.MethodGet, http.MethodPost}},
		{"/vis_camera/{vis_camera_id}", []string{http.MethodGet, http.MethodPut, http.MethodDelete}},
		{"/vis_camera/snapshot/{vis_camera_id}", []string{http.MethodPut}},
		{"/vis_camera/bloquear/{vis_camera_id}", []string{http.MethodPost}},
		{"/vis_camera/analitico/pausar/{vis_camera_id}", []string{http.MethodPost}},
		{"/vis_camera/licenca/liberar/{vis_camera_id}", []string{http.MethodPost}},
		{"/vis_camera/gravacao/ativar/{vis_camera_id}", []string{http.MethodPost}},
		{"/vis_camera/gravacao/desativar/{vis_camera_id}", []string{http.MethodPost}},
		{"/vis_camera/gravacao/flush/{vis_camera_id}", []string{http.MethodPost}},
		{"/vis_camera/gravacao/flush/ack/{vis_camera_id}", []string{http.MethodPost}},
		{"/vis_camera_area_by_camera", []string{http.MethodGet}},
		{"/vis_camera_area", []string{http.MethodPost}},
		{"/vis_camera_area/{id}", []string{http.MethodPut, http.MethodDelete}},
		{"/vis_licenca_by_franqueado", []string{http.MethodGet}},
		{"/vis_licenca", []string{http.MethodPost}},
		{"/vis_evento", []string{http.MethodPost}},
		{"/vis_evento_disparo_sensor", []string{http.MethodPost}},
		{"/vis_evento_demo_moni", []string{http.MethodPost}},
		{"/vis_evento_finalizar", []string{http.MethodPost}},
		{"/vis_evento/{id}", []string{http.MethodPut}},
		{"/vis_evento/{id}/clips", []string{http.MethodGet}},
		{"/vis_evento_clip", []string{http.MethodPost}},
		{"/vis_evento_by_franqueado_page", []string{http.MethodGet}},
		{"/vis_evento_by_cliente_page", []string{http.MethodGet}},
		{"/vis_evento_query_sensor_pendentes", []string{http.MethodGet}},
		{"/vis_evento_by_contexto", []string{http.MethodGet}},
		{"/vis_evento_by_processo_setor", []string{http.MethodGet}},
		{"/vis_evento_ultimos25_setor", []string{http.MethodGet}},
		{"/vis_evento_ultimos25_dispositivo", []string{http.MethodGet}},
		{"/vis_gravacao_storage_by_franqueado", []string{http.MethodGet}},
		{"/vis_gravacao_storage_credenciais_by_franqueado", []string{http.MethodGet}},
		{"/vis_gravacao_storage", []string{http.MethodPost}},
		{"/vis_gravacao_storage/{vis_gravacao_storage_id}", []string{http.MethodPut}},
		{"/vis_gravacao_segmento", []string{http.MethodPost}},
		{"/vis_gravacao_segmento_by_camera", []string{http.MethodGet}},
		{"/vis_gravacao_segmento_by_franqueado", []string{http.MethodGet}},
		{"/vis_mediamtx_node", []string{http.MethodGet}},
		{"/cvg_worker_tick", []string{http.MethodPost}},
		{"/ops/arme/agenda", []string{http.MethodGet}},
		{"/ops/atendimento/resolve", []string{http.MethodGet}},
		{"/ops/atendimento/politica", []string{http.MethodGet, http.MethodPut}},
		{"/ops/atendimento/ia_bloqueio", []string{http.MethodGet, http.MethodPost}},
		{"/ops/atendimento/creditos", []string{http.MethodGet}},
		{"/ops/vis_licenca/criar_lote", []string{http.MethodPost}},
		{"/ops/vis_licenca/ativar_pagamento", []string{http.MethodPost}},
		{"/ops/vis_licenca/estornar_pagamento", []string{http.MethodPost}},
		{"/ops/vis_licenca/sync_lote", []string{http.MethodPost}},
		{"/ops/vis_licenca/sync_franqueado", []string{http.MethodPost}},
		{"/ops/vis_capacidade/config_listar", []string{http.MethodGet}},
		{"/ops/vis_capacidade/config_salvar", []string{http.MethodPost}},
		{"/ops/vis_capacidade/reservar_pendente", []string{http.MethodPost}},
		{"/ops/vis_capacidade/ativar_pagamento", []string{http.MethodPost}},
		{"/ops/vis_capacidade/estornar_pagamento", []string{http.MethodPost}},
		{"/ops/vis_capacidade/listar_pendentes_renovacao", []string{http.MethodPost}},
		{"/ops/vis_capacidade/contrato/{contrato_id}", []string{http.MethodGet}},
	}

	r.HandleFunc("/cvg_worker_tick", handleCvgWorkerTick).Methods(http.MethodPost)

	// Webhook inbound Moni (autenticacao propria via WEBHOOK_INBOUND_KEY)
	r.HandleFunc("/vis_integracao/moni/inbound", handleDispatch).Methods(http.MethodPost)

	// Healthcheck publico (monitoramento / curl sem credencial)
	r.HandleFunc("/vis_health", handleDispatch).Methods(http.MethodGet)

	for _, rt := range routes {
		if rt.path == "/vis_health" {
			continue
		}
		h := workerAuth(handleDispatch)
		r.HandleFunc(rt.path, h).Methods(rt.methods...)
	}
}
