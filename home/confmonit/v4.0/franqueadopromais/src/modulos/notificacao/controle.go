package notificacao

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
		URI:    "/notificacoes/minhas",
		Metodo: http.MethodPost,
		Funcao: minhas,
		Aberto: false,
	},
	{
		URI:    "/notificacoes/contagem",
		Metodo: http.MethodPost,
		Funcao: contagem,
		Aberto: false,
	},
	{
		URI:    "/notificacoes/marcar-visto",
		Metodo: http.MethodPost,
		Funcao: marcarVisto,
		Aberto: false,
	},
	{
		URI:    "/notificacoes/marcar-lido",
		Metodo: http.MethodPost,
		Funcao: marcarLido,
		Aberto: false,
	},
}

func contextoDaSessao(r *http.Request) (map[string]string, error) {
	cookie, err := seguranca.LerCookies(r)
	if err != nil {
		return nil, err
	}
	master := strings.TrimSpace(cookie["master"])
	if master == "" {
		master = auxiliar.MasterDoCookie(r)
	}
	if master == "" {
		master = "N"
	}
	return map[string]string{
		"software":      "franqueadopro",
		"idUsuario":     strings.TrimSpace(cookie["idUsuario"]),
		"userTipo":      "FRA",
		"userMaster":    master,
		"idFranqueado":  strings.TrimSpace(cookie["idFranqueado"]),
		"idCliente":     "",
	}, nil
}

func proxyAPI(w http.ResponseWriter, r *http.Request, path string, payload map[string]string) {
	body, err := json.Marshal(payload)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	url := fmt.Sprintf("%s%s", config.ApiUrl, path)
	response, err := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
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

func minhas(w http.ResponseWriter, r *http.Request) {
	ctx, err := contextoDaSessao(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return
	}
	if ctx["idUsuario"] == "" {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, fmt.Errorf("sessao invalida"))
		return
	}
	proxyAPI(w, r, "/v4/notificacao/minhas", ctx)
}

func contagem(w http.ResponseWriter, r *http.Request) {
	ctx, err := contextoDaSessao(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return
	}
	if ctx["idUsuario"] == "" {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, fmt.Errorf("sessao invalida"))
		return
	}
	proxyAPI(w, r, "/v4/notificacao/contagem", ctx)
}

func marcarVisto(w http.ResponseWriter, r *http.Request) {
	marcar(w, r, "/v4/notificacao/marcarVisto")
}

func marcarLido(w http.ResponseWriter, r *http.Request) {
	marcar(w, r, "/v4/notificacao/marcarLido")
}

func marcar(w http.ResponseWriter, r *http.Request, path string) {
	ctx, err := contextoDaSessao(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return
	}
	if ctx["idUsuario"] == "" {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, fmt.Errorf("sessao invalida"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	var req struct {
		IDNotificacao string `json:"idNotificacao"`
	}
	_ = json.Unmarshal(body, &req)
	if strings.TrimSpace(req.IDNotificacao) == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("idNotificacao obrigatorio"))
		return
	}

	payload := map[string]string{
		"idNotificacao": strings.TrimSpace(req.IDNotificacao),
		"software":      ctx["software"],
		"idUsuario":     ctx["idUsuario"],
		"userTipo":      ctx["userTipo"],
		"idFranqueado":  ctx["idFranqueado"],
		"idCliente":     ctx["idCliente"],
	}
	proxyAPI(w, r, path, payload)
}
