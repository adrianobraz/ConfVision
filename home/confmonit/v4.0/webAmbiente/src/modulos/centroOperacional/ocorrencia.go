package centroOperacional

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"webAmbiente/src/resposta"
	"webAmbiente/src/xano"
)

func timelineListar(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		MapaAmbienteId int    `json:"mapa_ambiente_id"`
		IdProcesso     string `json:"idProcesso"`
		IdCliente      string `json:"idCliente"`
		Limite         int    `json:"limite"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	var req reqBody
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	if req.Limite <= 0 || req.Limite > 500 {
		req.Limite = 200
	}
	lista, err := xano.ListarCoTimeline(req.MapaAmbienteId, req.IdProcesso, req.IdCliente, req.Limite)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	if len(lista) == 0 {
		resposta.JsonVazio(w)
		return
	}
	resposta.JsonDados(w, http.StatusOK, lista)
}

func timelineSalvar(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Eventos []xano.CoTimelineEvento `json:"eventos"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	var req reqBody
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	if len(req.Eventos) == 0 {
		resposta.JsonDados(w, http.StatusOK, map[string]interface{}{"status": "OK", "inseridos": 0})
		return
	}
	resp, err := xano.SalvarCoTimeline(req.Eventos)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	resposta.JsonDados(w, http.StatusOK, resp)
}

func rastroListar(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		MapaAmbienteId int    `json:"mapa_ambiente_id"`
		IdProcesso     string `json:"idProcesso"`
		IdCliente      string `json:"idCliente"`
		Limite         int    `json:"limite"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	var req reqBody
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	if req.Limite <= 0 || req.Limite > 300 {
		req.Limite = 120
	}
	lista, err := xano.ListarCoRastro(req.MapaAmbienteId, req.IdProcesso, req.IdCliente, req.Limite)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "nao configurado") {
			resposta.JsonVazio(w)
			return
		}
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	if len(lista) == 0 {
		resposta.JsonVazio(w)
		return
	}
	resposta.JsonDados(w, http.StatusOK, lista)
}

func rastroSalvar(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Pontos []xano.CoRastroDisparo `json:"pontos"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	var req reqBody
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	if len(req.Pontos) == 0 {
		resposta.JsonDados(w, http.StatusOK, map[string]interface{}{"status": "OK", "inseridos": 0})
		return
	}
	resp, err := xano.SalvarCoRastro(req.Pontos)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	resposta.JsonDados(w, http.StatusOK, resp)
}
