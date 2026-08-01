package mapas

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"text/template"
	"webAmbiente/src/config"
	"webAmbiente/src/auxiliar"
	"webAmbiente/src/resposta"
	"webAmbiente/src/seguranca"
	"webAmbiente/src/tipos"
	"webAmbiente/src/xano"
)

var Rotas = []tipos.Rota{
	{
		Uri:      "/mapas/page",
		Metodo:   http.MethodGet,
		Controle: page,
		Seguro:   true,
	},
	{
		Uri:      "/mapas/listar",
		Metodo:   http.MethodPost,
		Controle: listar,
		Seguro:   true,
	},
}

func page(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/mapas/mapas.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/botoes.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		_ = err
	}

	d := tipos.Page{
		Titulo:       fmt.Sprintf("%s - Mapas", config.TituloSite),
		NavbarTitulo: "Mapas do cliente",
	}
	_ = page.ExecuteTemplate(w, "mapas.html", d)
}

func listar(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		IdCliente    string `json:"idCliente"`
		IdFranqueado string `json:"idFranqueado"`
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

	idFranqueado, err := seguranca.ResolverIdFranqueado(r, req.IdFranqueado, req.IdCliente)
	if err != nil {
		resposta.Erro(w, http.StatusUnauthorized, err)
		return
	}
	// Franqueado correto do cliente (evita mapa vazio quando sessionStorage envia franqueado antigo)
	if fra, e := auxiliar.FranqueadoDoCliente(req.IdCliente); e == nil && fra != "" {
		idFranqueado = fra
	}

	lista, err := xano.ListarMapaAmbiente(idFranqueado, req.IdCliente)
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
