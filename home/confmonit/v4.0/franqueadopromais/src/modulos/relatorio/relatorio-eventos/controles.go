package relatorioEventos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"franqueadopro/src/seguranca"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregar-relatorio-eventos",
		Metodo: http.MethodGet,
		Funcao: CarregarRelatorioEventos,
		Aberto: false,
	},
	{
		URI:    "/RelatorioEventosCarregarClientes",
		Metodo: http.MethodPost,
		Funcao: RelatorioEventosCarregarClientes,
		Aberto: false,
	},
	{
		URI:    "/relatorioEventosListar",
		Metodo: http.MethodPost,
		Funcao: relatorioEventosListar,
		Aberto: false,
	},
	{
		URI:    "/relatorioEventosCarregarDispositivos",
		Metodo: http.MethodPost,
		Funcao: relatorioEventosCarregarDispositivos,
		Aberto: false,
	},
}

func CarregarRelatorioEventos(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina

	d.TituloSite = config.TituloSite
	d.NomeTela = "Relatório dos Eventos"
	d.LogoMarca = "logo2Id6.png"
	d.LinkRetorno = "/carregar-menu-eventos"
	if r.URL.Query().Get("from") == "cd" || strings.Contains(r.Referer(), "carregar-menu-central-disparos") {
		d.LinkRetorno = "/carregar-menu-central-disparos"
	}
	auxiliar.ExecutarTemplate(w, "relatorio-eventos.html", d)
}

func RelatorioEventosCarregarClientes(w http.ResponseWriter, r *http.Request) {
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

func relatorioEventosListar(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var payload map[string]interface{}
	_ = json.Unmarshal(body, &payload)
	if payload == nil {
		payload = map[string]interface{}{}
	}

	dispId := strings.TrimSpace(fmt.Sprint(payload["dispId"]))
	apiPath := "/v4/evento/listarByDispStartEnd"

	// "Todos os clientes": usa o mesmo endpoint do dashboard (por franqueado).
	if strings.EqualFold(dispId, "TODOS") {
		idFra := strings.TrimSpace(fmt.Sprint(payload["idFranqueado"]))
		if idFra == "" || idFra == "<nil>" {
			if cookie, err := seguranca.LerCookies(r); err == nil {
				idFra = strings.TrimSpace(cookie["idFranqueado"])
			}
		}
		if idFra == "" {
			auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("idFranqueado obrigatorio para todos os clientes"))
			return
		}

		limit := jsonNumeroInt(payload["limit"], 100)
		if limit <= 0 {
			limit = 100
		}
		if limit > 100 {
			limit = 100
		}
		offset := jsonNumeroInt(payload["offset"], 0)
		if offset < 0 {
			offset = 0
		}

		out := map[string]interface{}{
			"idFranqueado": idFra,
			"dataInicio":   payload["dataInicio"],
			"dataFim":      payload["dataFim"],
			"limit":        limit,
			"offset":       offset,
		}
		body, erro = json.Marshal(out)
		if erro != nil {
			auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}
		apiPath = "/v4/evento/listarByIdFranqueadoStartEndGrupo"
	}

	url := fmt.Sprintf("%s%s", config.ApiUrl, apiPath)

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

func jsonNumeroInt(v interface{}, def int) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return def
		}
		return int(i)
	case string:
		var i int
		if _, err := fmt.Sscanf(strings.TrimSpace(n), "%d", &i); err == nil {
			return i
		}
	}
	return def
}

func relatorioEventosCarregarDispositivos(w http.ResponseWriter, r *http.Request) {
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
		return
	}

	auxiliar.RespostaAPP(w, corpo)
}
