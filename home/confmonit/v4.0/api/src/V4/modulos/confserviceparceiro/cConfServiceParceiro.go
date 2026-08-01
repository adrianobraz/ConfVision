package confserviceparceiroV4

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"api/src/V4/respApp"
)

func parceirosGlobal(w http.ResponseWriter, r *http.Request) {
	p, err := readAuthPayload(r)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	if _, err := resolveEscopo(p); err != nil {
		respApp.Erro(w, http.StatusForbidden, err)
		return
	}
	raw, status, err := csRequest(http.MethodGet, "/internal/parceiros/global", nil)
	if err != nil {
		respApp.Erro(w, http.StatusBadGateway, err)
		return
	}
	if status >= 400 {
		respApp.Erro(w, status, errors.New(csErroMsg(raw)))
		return
	}
	proxyDados(w, raw)
}

func parceirosRepListar(w http.ResponseWriter, r *http.Request) {
	p, err := readAuthPayload(r)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	esc, err := escopoCentralFromPayload(p)
	if err != nil {
		respApp.Erro(w, http.StatusForbidden, err)
		return
	}
	idRep := strings.TrimSpace(p.IDRepresentante)
	if idRep == "" {
		respApp.Erro(w, http.StatusBadRequest, errors.New("idRepresentante obrigatorio"))
		return
	}
	if err := assertRepDaCentral(esc.IDCentralCatalogo, idRep); err != nil {
		respApp.Erro(w, http.StatusForbidden, err)
		return
	}
	path := "/internal/parceiros/rep?idCentral=" + url.QueryEscape(esc.IDCentralCatalogo) +
		"&idRepresentante=" + url.QueryEscape(idRep)
	raw, status, err := csRequest(http.MethodGet, path, nil)
	if err != nil {
		respApp.Erro(w, http.StatusBadGateway, err)
		return
	}
	if status >= 400 {
		respApp.Erro(w, status, errors.New(csErroMsg(raw)))
		return
	}
	proxyDados(w, raw)
}

func parceirosRepSalvar(w http.ResponseWriter, r *http.Request) {
	p, err := readAuthPayload(r)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	esc, err := escopoCentralFromPayload(p)
	if err != nil {
		respApp.Erro(w, http.StatusForbidden, err)
		return
	}
	idRep := strings.TrimSpace(p.IDRepresentante)
	if idRep == "" {
		respApp.Erro(w, http.StatusBadRequest, errors.New("idRepresentante obrigatorio"))
		return
	}
	if err := assertRepDaCentral(esc.IDCentralCatalogo, idRep); err != nil {
		respApp.Erro(w, http.StatusForbidden, err)
		return
	}
	out, _ := json.Marshal(map[string]any{
		"idCentral":       esc.IDCentralCatalogo,
		"idRepresentante": idRep,
		"idsParceiro":     p.IDsParceiro,
	})
	raw, status, err := csRequest(http.MethodPost, "/internal/parceiros/rep/salvar", out)
	if err != nil {
		respApp.Erro(w, http.StatusBadGateway, err)
		return
	}
	if status >= 400 {
		respApp.Erro(w, status, errors.New(csErroMsg(raw)))
		return
	}
	proxyDados(w, raw)
}

func parceirosFranqueadoListar(w http.ResponseWriter, r *http.Request) {
	p, err := readAuthPayload(r)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	esc, err := escopoRepresentanteFromPayload(p)
	if err != nil {
		respApp.Erro(w, http.StatusForbidden, err)
		return
	}
	idFranq := strings.TrimSpace(p.IDFranqueado)
	if idFranq == "" {
		respApp.Erro(w, http.StatusBadRequest, errors.New("idFranqueado obrigatorio"))
		return
	}
	if err := assertFranqueadoDoRep(esc.IDVinculo, idFranq); err != nil {
		respApp.Erro(w, http.StatusForbidden, err)
		return
	}
	path := "/internal/parceiros/franqueado?idRepresentante=" + url.QueryEscape(esc.IDVinculo) +
		"&idFranqueado=" + url.QueryEscape(idFranq)
	raw, status, err := csRequest(http.MethodGet, path, nil)
	if err != nil {
		respApp.Erro(w, http.StatusBadGateway, err)
		return
	}
	if status >= 400 {
		respApp.Erro(w, status, errors.New(csErroMsg(raw)))
		return
	}
	proxyDados(w, raw)
}

func parceirosFranqueadoSalvar(w http.ResponseWriter, r *http.Request) {
	p, err := readAuthPayload(r)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	esc, err := escopoRepresentanteFromPayload(p)
	if err != nil {
		respApp.Erro(w, http.StatusForbidden, err)
		return
	}
	idFranq := strings.TrimSpace(p.IDFranqueado)
	if idFranq == "" {
		respApp.Erro(w, http.StatusBadRequest, errors.New("idFranqueado obrigatorio"))
		return
	}
	if err := assertFranqueadoDoRep(esc.IDVinculo, idFranq); err != nil {
		respApp.Erro(w, http.StatusForbidden, err)
		return
	}
	out, _ := json.Marshal(map[string]any{
		"idRepresentante": esc.IDVinculo,
		"idFranqueado":    idFranq,
		"idsParceiro":     p.IDsParceiro,
	})
	raw, status, err := csRequest(http.MethodPost, "/internal/parceiros/franqueado/salvar", out)
	if err != nil {
		respApp.Erro(w, http.StatusBadGateway, err)
		return
	}
	if status >= 400 {
		respApp.Erro(w, status, errors.New(csErroMsg(raw)))
		return
	}
	proxyDados(w, raw)
}

func parceirosRepDisponiveis(w http.ResponseWriter, r *http.Request) {
	p, err := readAuthPayload(r)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	esc, err := escopoRepresentanteFromPayload(p)
	if err != nil {
		respApp.Erro(w, http.StatusForbidden, err)
		return
	}
	path := "/internal/parceiros/rep?idCentral=" + url.QueryEscape(esc.IDCentralCatalogo) +
		"&idRepresentante=" + url.QueryEscape(esc.IDVinculo)
	raw, status, err := csRequest(http.MethodGet, path, nil)
	if err != nil {
		respApp.Erro(w, http.StatusBadGateway, err)
		return
	}
	if status >= 400 {
		respApp.Erro(w, status, errors.New(csErroMsg(raw)))
		return
	}
	proxyDados(w, raw)
}

func proxyDados(w http.ResponseWriter, raw []byte) {
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		respApp.Erro(w, http.StatusBadGateway, err)
		return
	}
	if v, ok := parsed["dados"]; ok {
		respApp.Dados(w, http.StatusOK, v)
		return
	}
	respApp.Dados(w, http.StatusOK, parsed)
}
