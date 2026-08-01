package relatorioAtendimento

import (
	"fmt"
	"html/template"
	"net/http"
	"webRepresentante/src/config"
	"webRepresentante/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:      "/relatorio/atendimento",
		Metodo:   http.MethodGet,
		Controle: relatorioAtendimentoPage,
		Seguro:   true,
	},
}

func relatorioAtendimentoPage(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/relatorioAtendimento/relatorioAtendimento.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/tabela.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	var d = tipos.Page{
		Titulo:       fmt.Sprintf("%s - Relatorio Atendimento", config.TituloSite),
		NavbarIcon:   "bi bi-file-earmark-richtext",
		NavbarTitulo: "Relatorio Atendimento",
	}
	if err := page.ExecuteTemplate(w, "relatorioAtendimento.html", d); err != nil {
		fmt.Println(err)
	}

}
