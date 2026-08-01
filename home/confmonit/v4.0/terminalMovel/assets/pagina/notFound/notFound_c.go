package notFound

import (
	"net/http"
	"terminal/src/templates"
)

func Pagina(w http.ResponseWriter, r *http.Request) {
	templates.ExecutarTemplate(w, "notFound.html", nil)
}
