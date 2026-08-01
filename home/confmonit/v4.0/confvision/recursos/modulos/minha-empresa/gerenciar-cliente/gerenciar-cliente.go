package gerenciarCliente

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/seguranca"

	"github.com/joho/godotenv"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/CarregarPaginaGerenciarCliente",
		Metodo: http.MethodGet,
		Funcao: carregarPaginaGerenciarCliente,
		Aberto: false,
	},
	{
		URI:    "/ClienteBuscarDados",
		Metodo: http.MethodPost,
		Funcao: ClienteBuscarDados,
		Aberto: false,
	},
	{
		URI:    "/ClienteHabilitar",
		Metodo: http.MethodPost,
		Funcao: ClienteHabilitar,
		Aberto: false,
	},
	{
		URI:    "/ClienteIncluir",
		Metodo: http.MethodPost,
		Funcao: ClienteIncluir,
		Aberto: false,
	},
	{
		URI:    "/ClienteAlterar",
		Metodo: http.MethodPost,
		Funcao: ClienteAlterar,
		Aberto: false,
	},
	{
		URI:    "/ClientePreExcluir",
		Metodo: http.MethodPost,
		Funcao: ClientePreExcluir,
		Aberto: false,
	},
	{
		URI:    "/ClienteRestauraPreExcluir",
		Metodo: http.MethodPost,
		Funcao: ClienteRestauraPreExcluir,
		Aberto: false,
	},
	{
		URI:    "/ClienteResetarSenha",
		Metodo: http.MethodPost,
		Funcao: ClienteResetarSenha,
		Aberto: false,
	},
	{
		URI:    "/ClienteListar",
		Metodo: http.MethodPost,
		Funcao: ClienteListar,
		Aberto: false,
	},
	{
		URI:    "/ClienteHabilitarEmail",
		Metodo: http.MethodPost,
		Funcao: ClienteHabilitarEmail,
		Aberto: false,
	},
	{
		URI:    "/ClienteEmailLivre",
		Metodo: http.MethodPost,
		Funcao: ClienteEmailLivre,
		Aberto: true,
	},
	{
		URI:    "/ClienteAlterarEmailPrincipal",
		Metodo: http.MethodPost,
		Funcao: ClienteAlterarEmailPrincipal,
		Aberto: true,
	},
	{
		URI:    "/ClienteDispositivoListarPorCliente",
		Metodo: http.MethodPost,
		Funcao: ClienteDispositivoListarPorCliente,
		Aberto: true,
	},
	{
		URI:    "/GerenciarClienteArmar",
		Metodo: http.MethodPost,
		Funcao: ClienteArmar,
		Aberto: true,
	},
	{
		URI:    "/clienteCarregarPacotes",
		Metodo: http.MethodPost,
		Funcao: clienteCarregarPacotes,
		Aberto: false,
	},
}

func carregarPaginaGerenciarCliente(w http.ResponseWriter, r *http.Request) {
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

		d.NomeTela = "Gerenciar Cliente"

		d.LinkRetorno = "/carregar-menu-confvision"

		d.LogoMarca = "logo2Id6.png"
		// Carrega o HTML
		auxiliar.ExecutarTemplate(w, "gerenciar-cliente.html", d)
	}
}

func ClienteBuscarDados(w http.ResponseWriter, r *http.Request) {

	// Recupera o dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/cliente/getDadosById", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		fmt.Println(erro)
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

func ClienteHabilitar(w http.ResponseWriter, r *http.Request) {

	// Recupera o dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/cliente/inverterAtivoById", config.ApiUrl)

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

func ClienteAlterar(w http.ResponseWriter, r *http.Request) {

	// Recupera o dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/cliente/alteraById", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, body)
}

func ClientePreExcluir(w http.ResponseWriter, r *http.Request) {

	// Recupera o dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/cliente/preDeleteById", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, body)
}

func ClienteRestauraPreExcluir(w http.ResponseWriter, r *http.Request) {

	// Recupera o dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/cliente/restauraPreDeleteById", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, body)
}

func ClienteResetarSenha(w http.ResponseWriter, r *http.Request) {

	// Recupera o dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/cliente/resetarSenha", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, body)
}

func ClienteAlterarEmailPrincipal(w http.ResponseWriter, r *http.Request) {

	// Recupera o dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/cliente/setEmail1ById", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, body)
}

// clienteListarReq corpo enviado pelo front do ConfVision.
type clienteListarReq struct {
	IDFranqueado string `json:"idFranqueado"`
	Termo        string `json:"termo"`
	FiltroStatus string `json:"filtroStatus"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
}

// clienteListarItem campos usados no filtro local (ativo vem de listaBloqueio na API).
type clienteListarItem struct {
	Ativo            string `json:"ativo"`
	DataCancelamento string `json:"dataCancelamento"`
}

func clientePassaFiltroStatus(ativo, dataCancelamento, filtro string) bool {
	canc := strings.TrimSpace(dataCancelamento) != ""
	switch filtro {
	case "ativo":
		return ativo == "S" && !canc
	case "desativado":
		return ativo == "N" && !canc
	case "cancelado":
		return canc
	default:
		return true
	}
}

func ClienteListar(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var req clienteListarReq
	if erro := json.Unmarshal(body, &req); erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	filtroStatus := strings.ToLower(strings.TrimSpace(req.FiltroStatus))

	// Nao envia filtroStatus para a API: la o campo usa cliente.Ativo (coluna),
	// enquanto o status real do cliente vem de listaBloqueio (campo ativo na resposta).
	apiPayload := map[string]interface{}{
		"idFranqueado": req.IDFranqueado,
		"termo":        req.Termo,
	}
	if filtroStatus == "" {
		apiPayload["limit"] = req.Limit
		apiPayload["offset"] = req.Offset
	} else {
		// Busca completa; filtra e pagina no ConfVision.
		apiPayload["limit"] = 0
		apiPayload["offset"] = 0
	}

	payloadBytes, erro := json.Marshal(apiPayload)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/cliente/listarByIdFranqueado", config.ApiUrl)
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	respBody, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if filtroStatus == "" {
		auxiliar.RespostaAPP(w, respBody)
		return
	}

	var apiResp struct {
		Status string            `json:"status"`
		Dados  []json.RawMessage `json:"dados"`
	}
	if erro := json.Unmarshal(respBody, &apiResp); erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	filtrados := make([]json.RawMessage, 0, len(apiResp.Dados))
	if apiResp.Status != "Vazio" {
		for _, raw := range apiResp.Dados {
			var item clienteListarItem
			if erro := json.Unmarshal(raw, &item); erro != nil {
				continue
			}
			if clientePassaFiltroStatus(item.Ativo, item.DataCancelamento, filtroStatus) {
				filtrados = append(filtrados, raw)
			}
		}
	}

	total := len(filtrados)
	limit := req.Limit
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 50
	}

	pagina := []json.RawMessage{}
	if offset < total {
		fim := offset + limit
		if fim > total {
			fim = total
		}
		pagina = filtrados[offset:fim]
	}

	hasMore := offset+len(pagina) < total
	auxiliar.RespostaJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "OK",
		"dados":   pagina,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
		"hasMore": hasMore,
	})
}

func ClienteHabilitarEmail(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/cliente/inverterEmailAtivoById", config.ApiUrl)

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

func ClienteEmailLivre(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/cliente/getEmail1LivreByEmail1", config.ApiUrl)

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

func ClienteArmar(w http.ResponseWriter, r *http.Request) {
	// Recupera o dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/armar", config.UrlComando)

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

func ClienteDispositivoListarPorCliente(w http.ResponseWriter, r *http.Request) {
	// Recupera o dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/dispositivo/listarByIdCliente", config.ApiUrl)

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

func ClienteIncluir(w http.ResponseWriter, r *http.Request) {

	// Recupera o dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/cliente/insere", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, body)
}

func clienteCarregarPacotes(w http.ResponseWriter, r *http.Request) {

	// Recupera o dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/pacote/listaByVinculo", config.ApiUrl)

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
