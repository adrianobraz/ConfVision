package dadosMonitoramento

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"franqueadopro/src/seguranca"

	"github.com/joho/godotenv"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/CarregarPaginaDadosMonitoramento",
		Metodo: http.MethodGet,
		Funcao: CarregarPaginaDadosMonitoramento,
		Aberto: false,
	},
	{
		URI:    "/DadosMonitoramentoBuscar",
		Metodo: http.MethodPost,
		Funcao: DadosMonitoramentoBuscar,
		Aberto: false,
	},
	{
		URI:    "/DadosMonitoramentoGravarSenha",
		Metodo: http.MethodPost,
		Funcao: DadosMonitoramentoGravarSenha,
		Aberto: false,
	},
	{
		URI:    "/DadosMonitoramentoEnvioEmailAtivo",
		Metodo: http.MethodPost,
		Funcao: DadosMonitoramentoEnvioEmailAtivo,
		Aberto: false,
	},
}

func CarregarPaginaDadosMonitoramento(w http.ResponseWriter, r *http.Request) {
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

		d.NomeTela = "Dados do Monitoramento"

		d.LinkRetorno = "/carregar-menu-gestao"

		d.LogoMarca = "logo2Id6.png"
		// Carrega o HTML
		auxiliar.ExecutarTemplate(w, "dados-monitoramento.html", d)
	}
}

// Buscar dados franqueado
func DadosMonitoramentoBuscar(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// url := fmt.Sprintf("%s/FranqueadoBuscar", config.ApiUrl)
	url := fmt.Sprintf("%s/v4/franqueado/getDadosById", config.ApiUrl)

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

func DadosMonitoramentoGravarSenha(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/usuario/alterarSenhaById", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(
		r, http.MethodPost, url, bytes.NewBuffer(corpo),
	)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	corpo, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	auxiliar.RespostaAPP(w, corpo)
}

func DadosMonitoramentoEnvioEmailAtivo(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/franqueado/inverterEmailAtivoById", config.ApiUrl)
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
