package atendimento

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"text/template"
	"webCliente/src/config"
	"webCliente/src/resposta"
	"webCliente/src/seguranca"
	"webCliente/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:      "/atendimento/page",
		Metodo:   http.MethodGet,
		Controle: page,
		Seguro:   true,
	},
	{
		Uri:      "/atendimento/listarProcessoToOpenByCliente",
		Metodo:   http.MethodPost,
		Controle: listarProcessoToOpenByCliente,
		Seguro:   true,
	},
	{
		Uri:      "/atendimento/atender",
		Metodo:   http.MethodPost,
		Controle: atender,
		Seguro:   true,
	},
	{
		Uri:      "/atendimento/visualizar",
		Metodo:   http.MethodPost,
		Controle: visualizar,
		Seguro:   true,
	},
	{
		Uri:      "/atendimento/manutencao",
		Metodo:   http.MethodPost,
		Controle: manutencao,
		Seguro:   true,
	},
}

func page(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/atendimento/atendimento.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/tabela.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	var d = tipos.Page{
		Titulo:       fmt.Sprintf("%s - Atendimento", config.TituloSite),
		NavbarTitulo: "Atendimento",
	}
	if err := page.ExecuteTemplate(w, "atendimento.html", d); err != nil {
		fmt.Println(err)
	}
}

func listarProcessoToOpenByCliente(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/terminal/listarProcessoToOpenByCliente", config.Api),
		bytes.NewBuffer(body),
	)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	if res.StatusCode >= 400 {
		resposta.TratarStatusCodeDeErro(w, res)
		return
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	resposta.App(w, body)
}

func atender(w http.ResponseWriter, r *http.Request)    {}
func visualizar(w http.ResponseWriter, r *http.Request) {}

func manutencao(w http.ResponseWriter, r *http.Request) {

}
