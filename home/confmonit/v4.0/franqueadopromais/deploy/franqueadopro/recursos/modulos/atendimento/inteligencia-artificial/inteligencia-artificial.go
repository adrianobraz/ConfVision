package inteligenciaArtificial

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"franqueadopro/src/xano"

	"github.com/joho/godotenv"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregar-inteligencia-artificial",
		Metodo: http.MethodGet,
		Funcao: carregarPagina,
		Aberto: false,
	},
	{
		URI:    "/iaTelefonesCarregar",
		Metodo: http.MethodPost,
		Funcao: telefonesCarregar,
		Aberto: false,
	},
	{
		URI:    "/iaTelefonesSalvar",
		Metodo: http.MethodPost,
		Funcao: telefonesSalvar,
		Aberto: false,
	},
	{
		URI:    "/iaBloqueioListar",
		Metodo: http.MethodPost,
		Funcao: bloqueioListar,
		Aberto: false,
	},
	{
		URI:    "/iaBloqueioAdicionar",
		Metodo: http.MethodPost,
		Funcao: bloqueioAdicionar,
		Aberto: false,
	},
	{
		URI:    "/iaBloqueioRemover",
		Metodo: http.MethodPost,
		Funcao: bloqueioRemover,
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
	d.NomeTela = "Inteligência Artificial"
	d.LogoMarca = "logo2Id6.png"
	d.LinkRetorno = "/carregar-menu-atendimento"
	auxiliar.ExecutarTemplate(w, "inteligencia-artificial.html", d)
}

func telefonesCarregar(w http.ResponseWriter, r *http.Request) {
	idFranqueado, err := idFranqueadoDoBody(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	path := fmt.Sprintf("/whatseventcadfranq/idfranqueado/%s", url.PathEscape(idFranqueado))
	raw, err := xano.Get(path)
	proxyXanoRaw(w, raw, err)
}

func telefonesSalvar(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var payload map[string]interface{}
	if erro = json.Unmarshal(body, &payload); erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	idFranqueado, _ := payload["idFranqueado"].(string)
	if idFranqueado == "" {
		idFranqueado, _ = payload["idfranq"].(string)
	}
	if idFranqueado == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("idFranqueado obrigatorio"))
		return
	}

	payload["idfranq"] = idFranqueado
	path := fmt.Sprintf("/whatseventcadfranq/franqueado/%s", url.PathEscape(idFranqueado))
	raw, err := xano.Put(path, payload)
	proxyXanoRaw(w, raw, err)
}

func bloqueioListar(w http.ResponseWriter, r *http.Request) {
	idFranqueado, err := idFranqueadoDoBody(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	path := fmt.Sprintf("/whatseventcadbloq/idFranqueado/%s", url.PathEscape(idFranqueado))
	raw, err := xano.Get(path)
	proxyXanoRaw(w, raw, err)
}

func bloqueioAdicionar(w http.ResponseWriter, r *http.Request) {
	proxyXanoPost(w, r, "/whatseventcadbloq")
}

func bloqueioRemover(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var payload struct {
		ID int `json:"id"`
	}
	if erro = json.Unmarshal(body, &payload); erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	if payload.ID <= 0 {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("id obrigatorio"))
		return
	}

	path := fmt.Sprintf("/whatseventcadbloq/%d", payload.ID)
	raw, err := xano.Delete(path)
	proxyXanoRaw(w, raw, err)
}

func idFranqueadoDoBody(r *http.Request) (string, error) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		return "", erro
	}

	var payload struct {
		IDFranqueado string `json:"idFranqueado"`
	}
	if len(body) > 0 {
		if erro = json.Unmarshal(body, &payload); erro != nil {
			return "", erro
		}
	}
	if payload.IDFranqueado == "" {
		return "", fmt.Errorf("idFranqueado obrigatorio")
	}
	return payload.IDFranqueado, nil
}

func proxyXanoPost(w http.ResponseWriter, r *http.Request, path string) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var payload map[string]interface{}
	if len(body) > 0 {
		if erro = json.Unmarshal(body, &payload); erro != nil {
			auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}
	}

	raw, err := xano.Post(path, payload)
	proxyXanoRaw(w, raw, err)
}

func proxyXanoRaw(w http.ResponseWriter, raw []byte, erro error) {
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, erro)
		return
	}

	var parsed interface{}
	if json.Unmarshal(raw, &parsed) == nil {
		dados := parsed
		if m, ok := parsed.(map[string]interface{}); ok {
			if inner, has := m["dados"]; has {
				dados = inner
			}
		}
		auxiliar.RespostaJSON(w, http.StatusOK, map[string]interface{}{
			"status": "OK",
			"dados":  dados,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}
