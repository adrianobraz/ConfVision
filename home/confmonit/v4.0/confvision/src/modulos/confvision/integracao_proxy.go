package confvision

import (
	"bytes"
	"confvision/src/auxiliar"
	"confvision/src/seguranca"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func ProxyIntegracaoGet(w http.ResponseWriter, r *http.Request) {
	idFra := idFranqueadoSessao(r)
	if idFra == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("id_franqueado obrigatorio"))
		return
	}
	path := fmt.Sprintf("/vis_integracao_franqueado?id_franqueado=%s", url.QueryEscape(idFra))
	proxyVisOrXano(w, r, http.MethodGet, path)
}

func ProxyIntegracaoSalvar(w http.ResponseWriter, r *http.Request) {
	idFra := idFranqueadoSessao(r)
	body, err := mergeFranqueadoBody(r, idFra)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	proxyVisOrXano(w, r, http.MethodPut, "/vis_integracao_franqueado")
}

func ProxyIntegracaoTestar(w http.ResponseWriter, r *http.Request) {
	idFra := idFranqueadoSessao(r)
	if idFra == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("id_franqueado obrigatorio"))
		return
	}
	body, _ := json.Marshal(map[string]any{"id_franqueado": idFra})
	r.Body = io.NopCloser(bytes.NewReader(body))
	proxyVisOrXano(w, r, http.MethodPost, "/vis_integracao_franqueado/test")
}

func ProxyIntegracaoLog(w http.ResponseWriter, r *http.Request) {
	idFra := idFranqueadoSessao(r)
	if idFra == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("id_franqueado obrigatorio"))
		return
	}
	limit := r.URL.Query().Get("limit")
	path := fmt.Sprintf("/vis_integracao_log?id_franqueado=%s&limit=%s", url.QueryEscape(idFra), url.QueryEscape(limit))
	proxyVisOrXano(w, r, http.MethodGet, path)
}

func idFranqueadoSessao(r *http.Request) string {
	cookie, _ := seguranca.LerCookies(r)
	return seguranca.IdFranqueadoDoCookie(cookie)
}

func mergeFranqueadoBody(r *http.Request, idFra string) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, err
		}
	}
	if idFra != "" {
		payload["id_franqueado"] = idFra
	}
	return json.Marshal(payload)
}
