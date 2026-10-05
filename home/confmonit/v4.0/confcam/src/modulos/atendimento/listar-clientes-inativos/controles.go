package listarClientesInativos

import (
	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/seguranca"
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/joho/godotenv"
)

func CarregarGerenciarClientesInativos(w http.ResponseWriter, r *http.Request) {

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

		d.NomeTela = "Lista de Clientes Inativos"

		d.LogoMarca = "logo2Id6.png"

		d.LinkRetorno = "/carregar-menu-atendimento"
		// Carrega o HTML
		auxiliar.ExecutarTemplate(w, "listar-clientes-inativos.html", d)
	}
}

func ListarClientesInativos(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/cliente/listarByIdFranqueado", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	auxiliar.RespostaAPP(w, corpo)
}
