package menuGestao

import (
	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"net/http"
)

var RotasMenuGestao = []auxiliar.Rota{
	{
		URI:    "/carregar-menu-gestao",
		Metodo: http.MethodGet,
		Funcao: CarregarMenuGestao,
		Aberto: false,
	},
}

func CarregarMenuGestao(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina

	d.TituloSite = config.TituloSite
	d.NomeTela = "Gestão Interna"
	d.LinkRetorno = "/carregar-menu-minha-empresa"
	d.LogoMarca = "logo2Id6.png"
	auxiliar.PreencherEhMaster(r, &d)

	auxiliar.ExecutarTemplate(w, "menu-gestao.html", d)
}
