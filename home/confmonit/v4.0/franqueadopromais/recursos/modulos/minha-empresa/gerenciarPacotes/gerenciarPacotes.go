package gerenciarPacotes

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
		URI:    "/gerenciarPacotesPage",
		Metodo: http.MethodGet,
		Funcao: gerenciarPacotesPage,
		Aberto: false,
	},
	{
		URI:    "/pacote/listar",
		Metodo: http.MethodPost,
		Funcao: listar,
		Aberto: true,
	},
	{
		URI:    "/pacote/buscar",
		Metodo: http.MethodPost,
		Funcao: buscar,
		Aberto: true,
	},
	{
		URI:    "/pacote/habilitar",
		Metodo: http.MethodPost,
		Funcao: habilitar,
		Aberto: true,
	},
	{
		URI:    "/pacote/inserir",
		Metodo: http.MethodPost,
		Funcao: inserir,
		Aberto: true,
	},
	{
		URI:    "/pacote/alterar",
		Metodo: http.MethodPost,
		Funcao: alterar,
		Aberto: true,
	},
}

func gerenciarPacotesPage(w http.ResponseWriter, r *http.Request) {
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

		d.NomeTela = "Gerenciar Pacote"

		d.LogoMarca = "logo2Id6.png"

		d.LinkRetorno = "/carregar-menu-comercial"
		// Carrega o HTML
		auxiliar.ExecutarTemplate(w, "gerenciarPacotes.html", d)
	}
}

// ok
func listar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.RequisiacaoAutenticada(
		r,
		http.MethodPost,
		fmt.Sprintf("%s/v4/pacote/listaByVinculo", config.ApiUrl),
		bytes.NewBuffer(body),
	)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	if res.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, res)
		return
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	auxiliar.RespostaAPP(w, body)
}

// ok
func buscar(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.RequisiacaoAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/pacote/getDadosById", config.ApiUrl),
		bytes.NewBuffer(body),
	)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	if res.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, res)
		return
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	auxiliar.RespostaAPP(w, body)
}

func habilitar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.RequisiacaoAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/pacote/inverteAtivoById", config.ApiUrl),
		bytes.NewBuffer(body),
	)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	if res.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, res)
		return
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	auxiliar.RespostaAPP(w, body)
}

// ok
func inserir(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.RequisiacaoAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/pacote/insere", config.ApiUrl),
		bytes.NewBuffer(body),
	)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	if res.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, res)
		return
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	auxiliar.RespostaAPP(w, body)
}

// ok
func alterar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.RequisiacaoAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/pacote/alteraById", config.ApiUrl),
		bytes.NewBuffer(body),
	)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	if res.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, res)
		return
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	auxiliar.RespostaAPP(w, body)
}
