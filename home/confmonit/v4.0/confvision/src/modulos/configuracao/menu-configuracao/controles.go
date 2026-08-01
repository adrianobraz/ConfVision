package menuConfiguracao

import (
	"confvision/src/auxiliar"
	"confvision/src/config"
	"net/http"

	"github.com/joho/godotenv"
)

var RotasMenuConfiguracoes = []auxiliar.Rota{
	{
		URI:    "/carregar-menu-configuracoes",
		Metodo: http.MethodGet,
		Funcao: CarregarMenuConfiguracoes,
		Aberto: false,
	},
}

func CarregarMenuConfiguracoes(w http.ResponseWriter, r *http.Request) {
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

		d.NomeTela = "Menu Configurações"

		d.LogoMarca = "logo2Id6.png"
		// Carrega o HTML
		auxiliar.ExecutarTemplate(w, "menu-configuracao.html", d)
	}
}
