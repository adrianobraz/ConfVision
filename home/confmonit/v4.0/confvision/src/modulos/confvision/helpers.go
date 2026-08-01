package confvision

import (
	"bytes"
	"confvision/src/auxiliar"
	"confvision/src/config"
	"io"
	"net/http"

	"github.com/joho/godotenv"
)

func paginaBase(nomeTela, linkRetorno string) auxiliar.Pagina {
	return auxiliar.Pagina{
		TituloSite:              config.TituloSite,
		NomeTela:                nomeTela,
		LinkRetorno:             linkRetorno,
		MediamtxHlsBase:         config.MediamtxHlsBase,
		MediamtxRtmpPublishBase: config.MediamtxRtmpPublishBase,
		MediamtxHlsPublic:       config.MediamtxHlsPublic,
		MediamtxRtmpPublic:      config.MediamtxRtmpPublic,
	}
}

func carregarPagina(w http.ResponseWriter, r *http.Request, template string, dados auxiliar.Pagina) {
	res, erro := godotenv.Read()
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if res["MANUTENCAO"] == "S" {
		dados.TituloSite = config.TituloSite
		auxiliar.ExecutarTemplate(w, "emManutencao.html", dados)
		return
	}

	auxiliar.ExecutarTemplate(w, template, dados)
}

func proxyConfmonitPost(w http.ResponseWriter, r *http.Request, url string) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	resp, erro := requisicaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer resp.Body.Close()

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

func proxyXano(w http.ResponseWriter, r *http.Request, metodo, path string) {
	url := config.XanoBaseUrl + path

	var body io.Reader
	if metodo == http.MethodPost || metodo == http.MethodPut {
		corpo, erro := io.ReadAll(r.Body)
		if erro != nil {
			auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}
		body = bytes.NewBuffer(corpo)
	}

	req, erro := http.NewRequest(metodo, url, body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, erro := client.Do(req)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer resp.Body.Close()

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

func proxyXanoCvg(w http.ResponseWriter, r *http.Request, metodo, path string) {
	base := config.XanoCvgBaseUrl
	if base == "" {
		http.Error(w, `{"status":"XANO_CVG_BASE_URL nao configurado"}`, http.StatusBadGateway)
		return
	}
	url := base + path

	var body io.Reader
	if metodo == http.MethodPost || metodo == http.MethodPut {
		corpo, erro := io.ReadAll(r.Body)
		if erro != nil {
			auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}
		body = bytes.NewBuffer(corpo)
	}

	req, erro := http.NewRequest(metodo, url, body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, erro := client.Do(req)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer resp.Body.Close()

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
