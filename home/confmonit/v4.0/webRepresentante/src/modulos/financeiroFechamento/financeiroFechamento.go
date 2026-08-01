package financeiroFechamento

import (
	"fmt"
	"html/template"
	"net/http"
	"webRepresentante/src/config"
	"webRepresentante/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:      "/financeiro/fechamento",
		Metodo:   http.MethodGet,
		Controle: fechamentoPage,
		Seguro:   true,
	},
}

func fechamentoPage(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/financeiroFechamento/financeiroFechamento.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/tabela.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	var d = tipos.Page{
		Titulo:       fmt.Sprintf("%s - Fechamento", config.TituloSite),
		NavbarIcon:   "bi bi-calendar2-week",
		NavbarTitulo: "Fechamento",
	}
	if err := page.ExecuteTemplate(w, "financeiroFechamento.html", d); err != nil {
		fmt.Println(err)
	}

}
