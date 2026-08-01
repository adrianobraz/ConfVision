package menuAlarmes

import (
	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"net/http"
)

func CarregarMenuAlarmes(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina

	// Carrega title do site
	d.TituloSite = config.TituloSite

	d.NomeTela = "Alarmes"

	d.LogoMarca = "logo2Id6.png"

	d.LinkRetorno = "/carregar-menu-relatorio"
	// Carrega o HTML
	auxiliar.ExecutarTemplate(w, "menu-alarmes.html", d)
}
