package auxiliar

import (
	"log"
	"net/http"
	"text/template"
)

var templates *template.Template

func CarregarTemplates() {
	templates = template.Must(template.ParseGlob("recursos/templates/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/login/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/emManutencao/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/confvision/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/administrator/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/templates/confvision-layout.html"))

	// Cadastros reutilizados do webFranqueado
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/gerenciar-cliente/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/minhas-licencas/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/whitelabel/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/dominio-saas/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/configuracao/gerenciar-dispositivo/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/configuracao/gerenciar-setores-alarme/*.html"))
}

func ExecutarTemplate(w http.ResponseWriter, template string, dados interface{}) {
	if erro := templates.ExecuteTemplate(w, template, dados); erro != nil {
		http.Error(w, "Erro ao renderizar pagina", http.StatusInternalServerError)
		log.Printf("template %s: %v", template, erro)
	}
}
