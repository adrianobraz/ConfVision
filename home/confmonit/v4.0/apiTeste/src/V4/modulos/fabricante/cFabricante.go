package fabricanteV4

import (
	"api/src/V4/respApp"
	"encoding/json"
	"io"
	"net/http"
)

func insere(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var f Fabricante

	if err := json.Unmarshal(body, &f); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := f.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, f.ID_Fabricante)
}

func getDadosById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var f Fabricante

	if err := json.Unmarshal(body, &f); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := f.GetDadosById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, f)
}

func alterarById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var f Fabricante

	if err := json.Unmarshal(body, &f); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := f.AlterarById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func deletaById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var f Fabricante

	if err := json.Unmarshal(body, &f); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := f.DeletaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func listar(w http.ResponseWriter, r *http.Request) {

	var f Fabricante
	var lista []Fabricante
	if err := f.Listar(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) <= 0 {
		respApp.Vazio(w)
	} else {
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func getAtivoById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var f Fabricante

	if err := json.Unmarshal(body, &f); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := f.GetAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, f.Ativo)
}

func setAtivoById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var f Fabricante

	if err := json.Unmarshal(body, &f); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := f.SetAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, f.Ativo)
}

func inverteAtivoById(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var f Fabricante

	if err := json.Unmarshal(body, &f); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := f.InverteAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}
