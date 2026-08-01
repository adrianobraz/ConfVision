package relatorioAtendimento

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"franqueadopro/src/seguranca"
)

func CarregarRelatorioAtendimento(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina

	// Carrega title do site
	d.TituloSite = config.TituloSite

	d.NomeTela = "Relatório dos Atendimento"

	d.LogoMarca = "logo2Id6.png"

	d.LinkRetorno = "/carregar-menu-relatorio"
	if r.URL.Query().Get("from") == "cd" || strings.Contains(r.Referer(), "carregar-menu-central-disparos") {
		d.LinkRetorno = "/carregar-menu-central-disparos"
	}
	// Carrega o HTML
	auxiliar.ExecutarTemplate(w, "relatorio-atendimento.html", d)
}

func relatorioAtendimentoListar(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/processo/listarEventosByFiltro", config.ApiUrl)

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
