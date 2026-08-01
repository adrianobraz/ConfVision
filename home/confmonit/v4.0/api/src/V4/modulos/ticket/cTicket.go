package ticketV4

import (
	"api/src/V4/respApp"
	"encoding/json"
	"io"
	"net/http"
)

func Insere(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var t Ticket

	if err := json.Unmarshal(body, &t); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := t.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, t.ID_Ticket)
}
func GetDadosById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var t Ticket

	if err := json.Unmarshal(body, &t); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := t.GetDadosById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, t)
}

func AlteraById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var t Ticket

	if err := json.Unmarshal(body, &t); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := t.AlteraById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func DeletaById(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var t Ticket

	if err := json.Unmarshal(body, &t); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := t.DeletaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func DeletaAllByMster(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var t Ticket

	if err := json.Unmarshal(body, &t); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := t.DeletaAllByMster(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func DeletaAllBySlave(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var t Ticket

	if err := json.Unmarshal(body, &t); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := t.DeletaAllBySlave(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.OK(w)
}

func ListarByIdMaster(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var t Ticket

	if err := json.Unmarshal(body, &t); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Ticket
	if err := t.ListarByIdMaster(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) <= 0 {
		respApp.Vazio(w)
	} else {
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func ListarByIdSlave(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var t Ticket

	if err := json.Unmarshal(body, &t); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Ticket
	if err := t.ListarByIdSlave(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) <= 0 {
		respApp.Vazio(w)
	} else {
		respApp.Dados(w, http.StatusOK, lista)
	}
}
