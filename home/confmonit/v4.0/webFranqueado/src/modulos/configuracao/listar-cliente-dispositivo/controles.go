package listarClienteDispositivo

import (
	"webFranqueado/src/auxiliar"
	"webFranqueado/src/config"
	"webFranqueado/src/seguranca"
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/joho/godotenv"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/listar-cliente-dispositivo",
		Metodo: http.MethodGet,
		Funcao: CarregarListarClienteDispositivo,
		Aberto: false,
	},

	{
		URI:    "/carregarTabela",
		Metodo: http.MethodPost,
		Funcao: carregarTabela,
		Aberto: false,
	},

	{
		URI:    "/GerenciarDispositivoHabilitar",
		Metodo: http.MethodPost,
		Funcao: GerenciarDispositivoHabilitar,
		Aberto: false,
	},
}

func CarregarListarClienteDispositivo(w http.ResponseWriter, r *http.Request) {
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

		d.NomeTela = "Listar Cliente Dispositivo"

		d.LogoMarca = "logo2Id6.png"

		d.LinkRetorno = "/carregar-menu-configuracoes"
		// Carrega o HTML
		auxiliar.ExecutarTemplate(w, "listar-cliente-dispositivo.html", d)
	}
}

func carregarTabela(w http.ResponseWriter, r *http.Request) {
	
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/cliente/listarComDispByIdFranqueado", config.ApiUrl)

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

	}

	auxiliar.RespostaAPP(w, corpo)
}

func GerenciarDispositivoHabilitar(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/DispositivoHabilitar", config.ApiUrl)

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
