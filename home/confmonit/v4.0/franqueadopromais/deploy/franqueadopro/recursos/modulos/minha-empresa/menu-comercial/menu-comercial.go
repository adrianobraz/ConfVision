package menuComercial

import (
	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"net/http"
)

var RotasMenuComercial = []auxiliar.Rota{
	{
		URI:    "/carregar-menu-comercial",
		Metodo: http.MethodGet,
		Funcao: CarregarMenuComercial,
		Aberto: false,
	},
}

func CarregarMenuComercial(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina

	d.TituloSite = config.TituloSite

	d.NomeTela = "Comercial"

	d.LinkRetorno = "/carregar-menu-minha-empresa"

	d.LogoMarca = "logo2Id6.png"

	auxiliar.ExecutarTemplate(w, "menu-comercial.html", d)
}
