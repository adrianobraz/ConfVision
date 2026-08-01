package menuCentralDisparos

import (
	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"net/http"
)

func CarregarMenuCentralDisparos(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina

	d.TituloSite = config.TituloSite
	d.NomeTela = "Central de Disparos"
	d.LogoMarca = "logo2Id6.png"
	d.LinkRetorno = "/carregar-menu-relatorio"

	auxiliar.ExecutarTemplate(w, "menu-central-disparos.html", d)
}
