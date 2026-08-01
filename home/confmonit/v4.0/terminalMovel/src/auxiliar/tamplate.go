package aux

import (
	"net/http"
	"text/template"
)

var templates *template.Template

func CarregarTemplates() {
	//MODULOS
	templates = template.Must(template.ParseGlob("html/login/*.html"))
	templates = template.Must(templates.ParseGlob("html/terminal/*.html"))

}

func ExecutarTemplate(w http.ResponseWriter, template string, dados interface{}) {
	templates.ExecuteTemplate(w, template, dados)
}
