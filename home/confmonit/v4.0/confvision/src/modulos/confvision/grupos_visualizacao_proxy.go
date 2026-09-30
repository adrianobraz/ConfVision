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
	"strings"

	"github.com/gorilla/mux"
)

func proxyGrupoVis(w http.ResponseWriter, r *http.Request, metodo, path string) {
	proxyXano(w, r, metodo, path)
}

func idFranqueadoOuErro(w http.ResponseWriter, r *http.Request) (string, bool) {
	cookie := sessaoCookie(r)
	if seguranca.EhAdministrator(cookie) {
		auxiliar.RespostaErro(w, http.StatusForbidden, fmt.Errorf("acesso restrito ao franqueado"))
		return "", false
	}
	idFra := strings.TrimSpace(seguranca.IdFranqueadoDoCookie(cookie))
	if idFra == "" {
		auxiliar.RespostaErro(w, http.StatusForbidden, fmt.Errorf("id_franqueado ausente na sessao"))
		return "", false
	}
	return idFra, true
}

func exigeFranqueadoGestao(w http.ResponseWriter, r *http.Request) (string, bool) {
	cookie := sessaoCookie(r)
	if seguranca.EhCliente(cookie) || seguranca.EhAdministrator(cookie) {
		auxiliar.RespostaErro(w, http.StatusForbidden, fmt.Errorf("acesso restrito ao franqueado"))
		return "", false
	}
	return idFranqueadoOuErro(w, r)
}

func ProxyListarGruposVisualizacao(w http.ResponseWriter, r *http.Request) {
	idFra, ok := idFranqueadoOuErro(w, r)
	if !ok {
		return
	}
	if seguranca.EhCliente(sessaoCookie(r)) {
		ProxyListarGruposDisponiveis(w, r)
		return
	}
	path := fmt.Sprintf("/vis_grupo_visualizacao?id_franqueado=%s", url.QueryEscape(idFra))
	proxyGrupoVis(w, r, http.MethodGet, path)
}

func ProxyListarGruposDisponiveis(w http.ResponseWriter, r *http.Request) {
	cookie := sessaoCookie(r)
	idFra := strings.TrimSpace(seguranca.IdFranqueadoDoCookie(cookie))
	idCli := strings.TrimSpace(seguranca.IdClienteDoCookie(cookie))
	if idFra == "" || idCli == "" {
		auxiliar.RespostaErro(w, http.StatusForbidden, fmt.Errorf("sessao cliente invalida"))
		return
	}
	path := fmt.Sprintf("/vis_grupo_visualizacao/disponiveis?id_franqueado=%s&id_cliente=%s",
		url.QueryEscape(idFra), url.QueryEscape(idCli))
	proxyGrupoVis(w, r, http.MethodGet, path)
}

func ProxyGetGrupoVisualizacao(w http.ResponseWriter, r *http.Request) {
	grupoID := mux.Vars(r)["id"]
	idFra, idCli, ok := sessaoGrupoLeitura(w, r)
	if !ok {
		return
	}
	path := fmt.Sprintf("/vis_grupo_visualizacao/%s?id_franqueado=%s", url.PathEscape(grupoID), url.QueryEscape(idFra))
	if idCli != "" {
		path += "&id_cliente=" + url.QueryEscape(idCli)
	}
	proxyGrupoVis(w, r, http.MethodGet, path)
}

func ProxyGetGrupoCameras(w http.ResponseWriter, r *http.Request) {
	grupoID := mux.Vars(r)["id"]
	idFra, idCli, ok := sessaoGrupoLeitura(w, r)
	if !ok {
		return
	}
	path := fmt.Sprintf("/vis_grupo_visualizacao/%s/cameras?id_franqueado=%s", url.PathEscape(grupoID), url.QueryEscape(idFra))
	if idCli != "" {
		path += "&id_cliente=" + url.QueryEscape(idCli)
	}
	proxyGrupoVis(w, r, http.MethodGet, path)
}

func sessaoGrupoLeitura(w http.ResponseWriter, r *http.Request) (idFra, idCli string, ok bool) {
	cookie := sessaoCookie(r)
	if seguranca.EhAdministrator(cookie) {
		auxiliar.RespostaErro(w, http.StatusForbidden, fmt.Errorf("acesso restrito"))
		return "", "", false
	}
	idFra = strings.TrimSpace(seguranca.IdFranqueadoDoCookie(cookie))
	if idFra == "" {
		auxiliar.RespostaErro(w, http.StatusForbidden, fmt.Errorf("id_franqueado ausente na sessao"))
		return "", "", false
	}
	if seguranca.EhCliente(cookie) {
		idCli = strings.TrimSpace(seguranca.IdClienteDoCookie(cookie))
		if idCli == "" {
			auxiliar.RespostaErro(w, http.StatusForbidden, fmt.Errorf("id_cliente ausente na sessao"))
			return "", "", false
		}
	}
	return idFra, idCli, true
}

func ProxyCreateGrupoVisualizacao(w http.ResponseWriter, r *http.Request) {
	idFra, ok := exigeFranqueadoGestao(w, r)
	if !ok {
		return
	}
	body, err := mergeFranqueadoBody(r, idFra)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	proxyGrupoVis(w, r, http.MethodPost, "/vis_grupo_visualizacao")
}

func ProxyUpdateGrupoVisualizacao(w http.ResponseWriter, r *http.Request) {
	idFra, ok := exigeFranqueadoGestao(w, r)
	if !ok {
		return
	}
	grupoID := mux.Vars(r)["id"]
	body, err := mergeFranqueadoBody(r, idFra)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	path := fmt.Sprintf("/vis_grupo_visualizacao/%s", url.PathEscape(grupoID))
	proxyGrupoVis(w, r, http.MethodPut, path)
}

func ProxyDeleteGrupoVisualizacao(w http.ResponseWriter, r *http.Request) {
	idFra, ok := exigeFranqueadoGestao(w, r)
	if !ok {
		return
	}
	grupoID := mux.Vars(r)["id"]
	raw, _ := json.Marshal(map[string]any{"id_franqueado": idFra})
	r.Body = io.NopCloser(bytes.NewReader(raw))
	path := fmt.Sprintf("/vis_grupo_visualizacao/%s?id_franqueado=%s", url.PathEscape(grupoID), url.QueryEscape(idFra))
	proxyGrupoVis(w, r, http.MethodDelete, path)
}

func ProxySaveGrupoComposicao(w http.ResponseWriter, r *http.Request) {
	idFra, ok := exigeFranqueadoGestao(w, r)
	if !ok {
		return
	}
	grupoID := mux.Vars(r)["id"]
	body, err := mergeFranqueadoBody(r, idFra)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	path := fmt.Sprintf("/vis_grupo_visualizacao/%s/composicao", url.PathEscape(grupoID))
	proxyGrupoVis(w, r, http.MethodPut, path)
}

func ProxyReordenarGrupoCameras(w http.ResponseWriter, r *http.Request) {
	idFra, ok := exigeFranqueadoGestao(w, r)
	if !ok {
		return
	}
	grupoID := mux.Vars(r)["id"]
	body, err := mergeFranqueadoBody(r, idFra)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	path := fmt.Sprintf("/vis_grupo_visualizacao/%s/reordenar", url.PathEscape(grupoID))
	proxyGrupoVis(w, r, http.MethodPut, path)
}

func CarregarGruposVisualizacao(w http.ResponseWriter, r *http.Request) {
	cookie := sessaoCookie(r)
	if seguranca.EhCliente(cookie) || seguranca.EhAdministrator(cookie) {
		http.Redirect(w, r, "/mosaicos", http.StatusFound)
		return
	}
	carregarPagina(w, r, "grupos-visualizacao.html", paginaBase("Grupos de visualização", "/carregar-menu-confvision"))
}

func CarregarMosaicos(w http.ResponseWriter, r *http.Request) {
	carregarPagina(w, r, "mosaicos.html", paginaBase("Mosaicos", "/carregar-menu-confvision"))
}

func CarregarMosaicoView(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	d := paginaBase("Mosaico", "/mosaicos")
	d.LinkJs = id
	carregarPagina(w, r, "mosaico-view.html", d)
}
