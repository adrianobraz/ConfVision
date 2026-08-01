package home

import (
	"net/http"
	"terminal/src/templates"
	"terminal/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:    "/home",
		Metodo: http.MethodGet,
		Funcao: paginaHome,
	},
}

func paginaHome(w http.ResponseWriter, r *http.Request) {
	templates.ExecutarTemplate(w, "home.html", nil)
}
