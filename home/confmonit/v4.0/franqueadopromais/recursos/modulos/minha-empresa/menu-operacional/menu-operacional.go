package menuOperacional

import (
	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"net/http"
)

var RotasMenuOperacional = []auxiliar.Rota{
	{
		URI:    "/carregar-menu-operacional",
		Metodo: http.MethodGet,
		Funcao: CarregarMenuOperacional,
		Aberto: false,
	},
}

func CarregarMenuOperacional(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina

	d.TituloSite = config.TituloSite

	d.NomeTela = "Operacional / Campo"

	d.LinkRetorno = "/carregar-menu-minha-empresa"

	d.LogoMarca = "logo2Id6.png"

	auxiliar.ExecutarTemplate(w, "menu-operacional.html", d)
}
