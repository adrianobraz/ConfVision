package geradorEventos

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
		URI:    "/CarregarGeradorEventos",
		Metodo: http.MethodGet,
		Funcao: CarregarGeradorEventos,
		Aberto: false,
	},
	{
		URI:    "/geradorEventoClientesListar",
		Metodo: http.MethodPost,
		Funcao: geradorEventoClientesListar,
		Aberto: false,
	},
	{
		URI:    "/geradorEventoDispositivoListar",
		Metodo: http.MethodPost,
		Funcao: geradorEventoDispositivoListar,
		Aberto: false,
	},
	{
		URI:    "/geradorEventoContacidListar",
		Metodo: http.MethodPost,
		Funcao: geradorEventoContacidListar,
		Aberto: false,
	},
	{
		URI:    "/geradorEventoUsuarioListar",
		Metodo: http.MethodPost,
		Funcao: geradorEventoUsuarioListar,
		Aberto: false,
	},
	{
		URI:    "/geradorEventoSetorListar",
		Metodo: http.MethodPost,
		Funcao: geradorEventoSetorListar,
		Aberto: false,
	},
	{
		URI:    "/geradorEventoEnviar",
		Metodo: http.MethodPost,
		Funcao: geradorEventoEnviar,
		Aberto: false,
	},
}

func CarregarGeradorEventos(w http.ResponseWriter, r *http.Request) {
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
		var d struct {
			auxiliar.Pagina
			LinkManual string
		}

		// Carrega title do site
		d.TituloSite = config.TituloSite

		d.NomeTela = "geradorEvento"

		d.LogoMarca = "logo2Id6.png"

		d.LinkRetorno = "/carregar-menu-atendimento"

		d.LinkManual = config.Manual.GeradorEvento
		// Carrega o HTML
		auxiliar.ExecutarTemplate(w, "geradorEventos.html", d)
	}
}

func geradorEventoClientesListar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do corpo da requidição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria URL para consulta a API
	url := fmt.Sprintf("%s/v4/cliente/listarByIdFranqueado", config.ApiUrl)

	// Executa a consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se algum erro foi retornado
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera o conteudo retornado
	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	// Respode ao requisitante
	auxiliar.RespostaAPP(w, corpo)
}

func geradorEventoDispositivoListar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do corpo da requidição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria URL para consulta a API
	url := fmt.Sprintf("%s/v4/dispositivo/listarByIdCliente", config.ApiUrl)

	// Executa a consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se algum erro foi retornado
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera o conteudo retornado
	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	// Respode ao requisitante
	auxiliar.RespostaAPP(w, corpo)
}

func geradorEventoContacidListar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do corpo da requidição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria URL para consulta a API
	url := fmt.Sprintf("%s/v4/contactid/listaPadrao", config.ApiUrl)

	// Executa a consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se algum erro foi retornado
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera o conteudo retornado
	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	// Respode ao requisitante
	auxiliar.RespostaAPP(w, corpo)
}

func geradorEventoUsuarioListar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do corpo da requidição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria URL para consulta a API
	url := fmt.Sprintf("%s/v4/usuarioAlarme/listaByIdDispositivo", config.ApiUrl)

	// Executa a consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se algum erro foi retornado
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera o conteudo retornado
	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	// Respode ao requisitante
	auxiliar.RespostaAPP(w, corpo)
}

func geradorEventoSetorListar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do corpo da requidição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
 
	// Cria URL para consulta a API
	url := fmt.Sprintf("%s/v4/setor/listaByIdDispositivo", config.ApiUrl)

	// Executa a consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se algum erro foi retornado
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera o conteudo retornado
	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	// Respode ao requisitante
	auxiliar.RespostaAPP(w, corpo)
}

func geradorEventoEnviar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do corpo da requidição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Executa a consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(
		r,
		http.MethodPost,
		config.UrlReceptor,
		bytes.NewBuffer(body),
	)

	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se algum erro foi retornado
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera o conteudo retornado
	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	// Respode ao requisitante
	auxiliar.RespostaAPP(w, corpo)
}
