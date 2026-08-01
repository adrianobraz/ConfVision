package central

import (
	"fmt"
	"html/template"
	"net/http"
	"webCentral/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:      "/central",
		Metodo:   http.MethodGet,
		Controle: centralPage,
		Seguro:   true,
	},
}

func centralPage(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/central/central.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/botoes.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	var d = tipos.Page{
		Titulo:       "Painel Central",
		NavbarLink:   "/home",
		NavbarIcon:   "bi bi-speedometer2",
		NavbarTitulo: "Painel Central",
	}
	if err := page.ExecuteTemplate(w, "central.html", d); err != nil {
		fmt.Println(err)
	}

}
