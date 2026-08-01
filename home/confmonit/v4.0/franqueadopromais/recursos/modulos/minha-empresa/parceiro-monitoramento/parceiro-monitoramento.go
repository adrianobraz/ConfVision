package parceiroMonitoramento

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"

	"github.com/joho/godotenv"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregar-parceiro-monitoramento",
		Metodo: http.MethodGet,
		Funcao: carregarPagina,
		Aberto: false,
	},
	{
		URI:    "/pmParceirosListar",
		Metodo: http.MethodPost,
		Funcao: parceirosListar,
		Aberto: false,
	},
	{
		URI:    "/pmVinculosListar",
		Metodo: http.MethodPost,
		Funcao: vinculosListar,
		Aberto: false,
	},
	{
		URI:    "/pmVinculoSalvar",
		Metodo: http.MethodPost,
		Funcao: vinculoSalvar,
		Aberto: false,
	},
	{
		URI:    "/pmVinculoRemover",
		Metodo: http.MethodPost,
		Funcao: vinculoRemover,
		Aberto: false,
	},
	{
		URI:    "/pmFaturasListar",
		Metodo: http.MethodPost,
		Funcao: faturasListar,
		Aberto: false,
	},
}

func carregarPagina(w http.ResponseWriter, r *http.Request) {
	res, erro := godotenv.Read()
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	if res["MANUTENCAO"] == "S" {
		var d auxiliar.Pagina
		d.TituloSite = config.TituloSite
		auxiliar.ExecutarTemplate(w, "emManutencao.html", d)
		return
	}
	var d auxiliar.Pagina
	d.TituloSite = config.TituloSite
	d.NomeTela = "Parceiro de Monitoramento"
	d.LogoMarca = "logo2Id6.png"
	d.LinkRetorno = "/carregar-menu-atendimento"
	auxiliar.ExecutarTemplate(w, "parceiro-monitoramento.html", d)
}

func parceirosListar(w http.ResponseWriter, r *http.Request) {
	idFranqueado, err := idFranqueadoDoBody(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	path := "/internal/parceiros/catalogo?idFranqueado=" + url.QueryEscape(idFranqueado)
	raw, status, err := csRequest(http.MethodGet, path, nil)
	proxyCS(w, raw, status, err)
}

func vinculosListar(w http.ResponseWriter, r *http.Request) {
	idFranqueado, err := idFranqueadoDoBody(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	path := "/internal/vinculos?idFranqueado=" + url.QueryEscape(idFranqueado)
	raw, status, err := csRequest(http.MethodGet, path, nil)
	proxyCS(w, raw, status, err)
}

func vinculoSalvar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	raw, status, err := csRequest(http.MethodPost, "/internal/vinculo", body)
	proxyCS(w, raw, status, err)
}

func vinculoRemover(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	raw, status, err := csRequest(http.MethodPost, "/internal/vinculo/desativar", body)
	proxyCS(w, raw, status, err)
}

func faturasListar(w http.ResponseWriter, r *http.Request) {
	idFranqueado, err := idFranqueadoDoBody(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	path := "/internal/faturas/franqueado?idFranqueado=" + url.QueryEscape(idFranqueado)
	raw, status, err := csRequest(http.MethodGet, path, nil)
	proxyCS(w, raw, status, err)
}

func idFranqueadoDoBody(r *http.Request) (string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	var payload struct {
		IDFranqueado string `json:"idFranqueado"`
	}
	if len(body) > 0 {
		if err = json.Unmarshal(body, &payload); err != nil {
			return "", err
		}
	}
	if strings.TrimSpace(payload.IDFranqueado) == "" {
		return "", fmt.Errorf("idFranqueado obrigatorio")
	}
	return payload.IDFranqueado, nil
}

func csRequest(method, path string, body []byte) ([]byte, int, error) {
	base := strings.TrimRight(config.ConfServiceURL, "/")
	if base == "" {
		return nil, 0, fmt.Errorf("CONFSERVICE_URL nao configurado")
	}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, base+path, reader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if config.ConfServiceKey != "" {
		req.Header.Set("X-Api-Key", config.ConfServiceKey)
	}
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	return raw, resp.StatusCode, err
}

func proxyCS(w http.ResponseWriter, raw []byte, status int, err error) {
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		if status >= 400 {
			auxiliar.RespostaErro(w, status, fmt.Errorf("confservice erro"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(raw)
		return
	}
	if status >= 400 {
		msg := "confservice erro"
		if e, ok := parsed["erro"].(string); ok && e != "" {
			msg = e
		}
		auxiliar.RespostaErro(w, status, fmt.Errorf("%s", msg))
		return
	}
	dados := interface{}(parsed)
	if v, ok := parsed["dados"]; ok {
		dados = v
	}
	auxiliar.RespostaJSON(w, http.StatusOK, map[string]interface{}{
		"status": "OK",
		"dados":  dados,
	})
}
