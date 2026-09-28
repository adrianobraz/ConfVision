package confvision

import (
	"confvision/src/auxiliar"
	"net/http"
	"net/url"
)

func CarregarRelatorioOperacional(w http.ResponseWriter, r *http.Request) {
	carregarPagina(w, r, "relatorio-operacional.html", paginaBase("Relatório operacional", "/carregar-menu-confvision"))
}

func proxyRelatorioOperacional(w http.ResponseWriter, r *http.Request, suffix string) {
	idFra, err := assertFranqueadoQuery(r, r.URL.Query().Get("id_franqueado"))
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusForbidden, err)
		return
	}
	q := url.Values{}
	q.Set("id_franqueado", idFra)
	if lim := r.URL.Query().Get("limit"); lim != "" {
		q.Set("limit", lim)
	}
	path := "/vis_relatorio_operacional/" + suffix + "?" + q.Encode()
	status, raw, ok := fetchXano(http.MethodGet, path)
	if !ok || status >= 400 {
		auxiliar.RespostaErro(w, http.StatusBadGateway, errFalhaVisdata)
		return
	}
	auxiliar.RespostaAPP(w, raw)
}

func ProxyRelatorioStream(w http.ResponseWriter, r *http.Request) {
	proxyRelatorioOperacional(w, r, "stream")
}

func ProxyRelatorioHealth(w http.ResponseWriter, r *http.Request) {
	idFra, err := assertFranqueadoQuery(r, "")
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusForbidden, err)
		return
	}
	_ = idFra
	limit := r.URL.Query().Get("limit")
	path := "/vis_relatorio_operacional/health"
	if limit != "" {
		path += "?limit=" + url.QueryEscape(limit)
	}
	status, raw, ok := fetchXano(http.MethodGet, path)
	if !ok || status >= 400 {
		auxiliar.RespostaErro(w, http.StatusBadGateway, errFalhaVisdata)
		return
	}
	auxiliar.RespostaAPP(w, raw)
}

func ProxyRelatorioMetric(w http.ResponseWriter, r *http.Request) {
	limit := r.URL.Query().Get("limit")
	path := "/vis_relatorio_operacional/metric"
	if limit != "" {
		path += "?limit=" + url.QueryEscape(limit)
	}
	status, raw, ok := fetchXano(http.MethodGet, path)
	if !ok || status >= 400 {
		auxiliar.RespostaErro(w, http.StatusBadGateway, errFalhaVisdata)
		return
	}
	auxiliar.RespostaAPP(w, raw)
}

var errFalhaVisdata = &proxyVisErr{msg: "falha ao consultar relatório"}

type proxyVisErr struct{ msg string }

func (e *proxyVisErr) Error() string { return e.msg }
