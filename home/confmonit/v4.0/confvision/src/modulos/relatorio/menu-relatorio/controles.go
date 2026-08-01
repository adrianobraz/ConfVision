package menuRelatorio

import (
	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/seguranca"
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func CarregarMenuRelatorio(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina

	// Carrega title do site
	d.TituloSite = config.TituloSite

	d.NomeTela = "Menu Relatório"

	d.LogoMarca = "logo2Id6.png"
	// Carrega o HTML
	auxiliar.ExecutarTemplate(w, "menu-relatorio.html", d)
}

func RelatorioLigacoesCarregarClientes(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/ClienteListar", config.ApiUrl)

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

	}

	auxiliar.RespostaAPP(w, corpo)
}
