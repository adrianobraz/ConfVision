package feedback

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
		URI:    "/feedback/enviar",
		Metodo: http.MethodPost,
		Funcao: enviar,
		Aberto: false,
	},
}

func enviar(w http.ResponseWriter, r *http.Request) {
	cookie, err := seguranca.LerCookies(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return
	}
	idUsuario := strings.TrimSpace(cookie["idUsuario"])
	if idUsuario == "" {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, fmt.Errorf("sessao invalida"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	var req struct {
		Tipo       string `json:"tipo"`
		Descricao  string `json:"descricao"`
		URLPagina  string `json:"urlPagina"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	payload := map[string]string{
		"software":     "franqueadopro",
		"tipo":         strings.TrimSpace(req.Tipo),
		"descricao":    strings.TrimSpace(req.Descricao),
		"urlPagina":    strings.TrimSpace(req.URLPagina),
		"idUsuario":    idUsuario,
		"idFranqueado": strings.TrimSpace(cookie["idFranqueado"]),
	}

	bodyOut, err := json.Marshal(payload)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	url := fmt.Sprintf("%s/v4/feedback/enviar", config.ApiUrl)
	response, err := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(bodyOut))
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, response)
		return
	}
	corpo, err := io.ReadAll(response.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	auxiliar.RespostaAPP(w, corpo)
}
