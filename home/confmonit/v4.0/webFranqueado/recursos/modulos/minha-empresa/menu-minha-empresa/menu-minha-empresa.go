package menuMinhaEmpresa

import (
	"net/http"
	"webFranqueado/src/auxiliar"
	"webFranqueado/src/config"

	"github.com/joho/godotenv"
)

var RotasMenuMinhaEmpresa = []auxiliar.Rota{
	{
		URI:    "/carregar-menu-minha-empresa",
		Metodo: http.MethodGet,
		Funcao: CarregarMenuMinhaEmpresa,
		Aberto: false,
	},
	
}

func CarregarMenuMinhaEmpresa(w http.ResponseWriter, r *http.Request) {
	res, erro := godotenv.Read()
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if res["MANUTENCAO"] == "S" {
		var d auxiliar.Pagina

		d.TituloSite = config.TituloSite

		auxiliar.ExecutarTemplate(w, "emManutencao.html", d)
	} else {
		var d auxiliar.Pagina

		// Carrega title do site
		d.TituloSite = config.TituloSite

		d.NomeTela = "Menu Minha Empresa"

		d.LinkRetorno = "/carregar-menu-principal"

		d.LogoMarca = "logo2Id6.png"
		// Carrega o HTML
		auxiliar.ExecutarTemplate(w, "menu-minha-empresa.html", d)
	}
}
