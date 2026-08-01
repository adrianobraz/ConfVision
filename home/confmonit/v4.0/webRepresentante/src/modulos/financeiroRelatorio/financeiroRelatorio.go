package financeiroRelatorio

import (
	"fmt"
	"html/template"
	"net/http"
	"webRepresentante/src/config"
	"webRepresentante/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:      "/financeiro/relatorio",
		Metodo:   http.MethodGet,
		Controle: relatorioPage,
		Seguro:   true,
	},
}

func relatorioPage(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/financeiroRelatorio/financeiroRelatorio.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	var d = tipos.Page{
		Titulo:       fmt.Sprintf("%s Relatorio", config.TituloSite),
		NavbarIcon:   "bi bi-file-earmark-richtext",
		NavbarTitulo: "Relatorio",
	}
	if err := page.ExecuteTemplate(w, "financeiroRelatorio.html", d); err != nil {
		fmt.Println(err)
	}

}
