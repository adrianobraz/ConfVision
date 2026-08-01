package menuAtendimento

import (
	"confvision/src/auxiliar"
	"confvision/src/config"
	"net/http"
)

func CarregarMenuAtendimento(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina

	// Carrega title do site
	d.TituloSite = config.TituloSite

	d.NomeTela = "Menu Atendimento"

	d.LogoMarca = "logo2Id6.png"

	d.LinkRetorno = "/carregar-menu-principal"
	// Carrega o HTML
	auxiliar.ExecutarTemplate(w, "menu-atendimento.html", d)
}
