package setupRelatorioV4

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

	var setRel SetupRelatorio

	if err := json.Unmarshal(body, &setRel); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := setRel.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, setRel.ID_SetupRelatorio)
}

func getDadosById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var setRel SetupRelatorio

	if err := json.Unmarshal(body, &setRel); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := setRel.GetDadosById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, setRel)
}

func getDadosByIdCliente(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var setRel SetupRelatorio

	if err := json.Unmarshal(body, &setRel); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := setRel.GetDadosByIdCliente(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, setRel)
}

func alteraById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var setRel SetupRelatorio

	if err := json.Unmarshal(body, &setRel); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := setRel.AlteraById(); err != nil {
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

	var setRel SetupRelatorio

	if err := json.Unmarshal(body, &setRel); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := setRel.DeletaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func deletaByIdCliente(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var setRel SetupRelatorio

	if err := json.Unmarshal(body, &setRel); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := setRel.DeletaByIdCliente(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}
