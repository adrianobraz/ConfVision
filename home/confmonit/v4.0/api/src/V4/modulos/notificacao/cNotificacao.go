package notificacao

import (
	"api/src/V4/respApp"
	"encoding/json"
	"io"
	"net/http"
)

func criar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var n Notificacao
	if err := json.Unmarshal(body, &n); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := n.Criar(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, n)
}

func listar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var filtro struct {
		Ativo string `json:"ativo"`
	}
	_ = json.Unmarshal(body, &filtro)

	var lista []Notificacao
	var n Notificacao
	if err := n.ListarAdmin(&lista, filtro.Ativo); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) > 0 {
		respApp.Dados(w, http.StatusOK, lista)
	} else {
		respApp.Vazio(w)
	}
}

func desativar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var req DesativarReq
	if err := json.Unmarshal(body, &req); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	n := Notificacao{ID_Notificacao: req.ID_Notificacao}
	if err := n.Desativar(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func minhas(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var ctx ContextoUsuario
	if err := json.Unmarshal(body, &ctx); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Notificacao
	if err := Minhas(ctx, &lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) > 0 {
		respApp.Dados(w, http.StatusOK, lista)
	} else {
		respApp.Vazio(w)
	}
}

func contagem(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var ctx ContextoUsuario
	if err := json.Unmarshal(body, &ctx); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	total, err := Contagem(ctx)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, map[string]int{"total": total})
}

func marcarVisto(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var req LeituraReq
	if err := json.Unmarshal(body, &req); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := MarcarVisto(req); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func marcarLido(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var req LeituraReq
	if err := json.Unmarshal(body, &req); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := MarcarLido(req); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}
