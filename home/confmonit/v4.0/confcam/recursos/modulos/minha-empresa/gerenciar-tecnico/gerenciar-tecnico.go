package gerenciarTecnico

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/seguranca"

	"github.com/joho/godotenv"
)

var RotasGerenciarTecnico = []auxiliar.Rota{
	{
		URI:    "/CarregarPaginaGerenciarTecnico",
		Metodo: http.MethodGet,
		Funcao: CarregarPaginaGerenciarTecnico,
		Aberto: false,
	},

	{
		URI:    "/TecnicoCarregarTabela",
		Metodo: http.MethodPost,
		Funcao: TecnicoCarregarTabela,
		Aberto: false,
	},
	{
		URI:    "/TecnicoResetarSenha",
		Metodo: http.MethodPost,
		Funcao: TecnicoResetarSenha,
		Aberto: false,
	},
	{
		URI:    "/TecnicoHabilitar",
		Metodo: http.MethodPost,
		Funcao: TecnicoHabilitar,
		Aberto: false,
	},
	{
		URI:    "/TecnicoInserir",
		Metodo: http.MethodPost,
		Funcao: TecnicoInserir,
		Aberto: false,
	},
	{
		URI:    "/TecnicoBuscar",
		Metodo: http.MethodPost,
		Funcao: TecnicoBuscar,
		Aberto: false,
	},
	{
		URI:    "/TecnicoAlterar",
		Metodo: http.MethodPost,
		Funcao: TecnicoAlterar,
		Aberto: false,
	},
	{
		URI:    "/TecnicoExcluir",
		Metodo: http.MethodPost,
		Funcao: TecnicoExcluir,
		Aberto: false,
	},
}

func CarregarPaginaGerenciarTecnico(w http.ResponseWriter, r *http.Request) {
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

		d.NomeTela = "Gerenciar Tecnico"

		d.LogoMarca = "logo2Id6.png"

		d.LinkRetorno = "/carregar-menu-minha-empresa"
		// Carrega o HTML
		auxiliar.ExecutarTemplate(w, "gerenciar-tecnico.html", d)
	}
}

func TecnicoCarregarTabela(w http.ResponseWriter, r *http.Request) {

	// Recebe o conteudo do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria url para consultar a API
	url := fmt.Sprintf("%s/v4/tecnico/listaByIdVinculo", config.ApiUrl)

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

func TecnicoResetarSenha(w http.ResponseWriter, r *http.Request) {

	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/tecnico/resetSenhaById", config.ApiUrl)

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

func TecnicoHabilitar(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/tecnico/inverteAtivoById", config.ApiUrl)

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

func TecnicoInserir(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/tecnico/insere", config.ApiUrl)
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(corpo))
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

func TecnicoBuscar(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/tecnico/getDadosById", config.ApiUrl)

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

func TecnicoAlterar(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	// fmt.Println(string(corpo))
	url := fmt.Sprintf("%s/v4/tecnico/alteraById", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(corpo))
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

func TecnicoExcluir(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/tecnico/deletaById", config.ApiUrl)

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
