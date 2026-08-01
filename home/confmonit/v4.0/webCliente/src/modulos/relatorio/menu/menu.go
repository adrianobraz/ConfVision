package relatorioMenu

import (
	"fmt"
	"html/template"
	"net/http"
	"webCliente/src/config"
	"webCliente/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:      "/relatorio/menu/page",
		Metodo:   http.MethodGet,
		Controle: page,
		Seguro:   true,
	},
}

func page(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/relatorio/menu/menu.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/botoes.html",
		"public/templates/components/tabela.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	var d = tipos.Page{
		Titulo:       fmt.Sprintf("%s - Menu Relatório", config.TituloSite),
		NavbarTitulo: "Menu Relatório",
	}
	if err := page.ExecuteTemplate(w, "menu.html", d); err != nil {
		fmt.Println(err)
	}
}
