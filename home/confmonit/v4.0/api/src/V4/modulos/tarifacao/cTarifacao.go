package tarifacaoV4

import (
	"api/src/V4/respApp"
	"encoding/json"
	"io"
	"net/http"
)

func insereLancamento(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var t Tarifacao
	if err := json.Unmarshal(body, &t); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := t.InsereLancamento(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, t)
}

func listaLancamentosPendentes(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var t Tarifacao
	if len(body) > 0 {
		if err := json.Unmarshal(body, &t); err != nil {
			respApp.Erro(w, http.StatusBadRequest, err)
			return
		}
	}

	var lista []Tarifacao
	if err := t.ListaLancamentosPendentes(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) > 0 {
		respApp.Dados(w, http.StatusOK, lista)
	} else {
		respApp.Vazio(w)
	}
}

func deleteLancamento(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var t Tarifacao
	if err := json.Unmarshal(body, &t); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := t.DeleteById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}
