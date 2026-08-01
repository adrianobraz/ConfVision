package templates

import (
	"net/http"
	"text/template"
)

var templates *template.Template

// CarregarTemplates carrega as paginas html na variavel templates
func CarregarTemplates() {

	// Nao encotrada
	templates = template.Must(template.ParseGlob("assets/pagina/notFound/*.html"))

	// Login
	templates = template.Must(templates.ParseGlob("assets/pagina/login/*.html"))

	// Home
	templates = template.Must(templates.ParseGlob("assets/pagina/home/*.html"))
}

func ExecutarTemplate(w http.ResponseWriter, template string, dados interface{}) {
	templates.ExecuteTemplate(w, template, dados)
}
