package emailEventoV4

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

	var emEvt EmailEvento

	if err := json.Unmarshal(body, &emEvt); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := emEvt.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, emEvt.ID_SetupEnvioEvento)
}

func getDadosById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var emEvt EmailEvento

	if err := json.Unmarshal(body, &emEvt); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := emEvt.GetDadosById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, emEvt)
}

func getDadosByIdDispositivo(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var emEvt EmailEvento

	if err := json.Unmarshal(body, &emEvt); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := emEvt.GetDadosByIdDispositivo(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, emEvt)
}

func alteraById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var emEvt EmailEvento

	if err := json.Unmarshal(body, &emEvt); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := emEvt.AlteraById(); err != nil {
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

	var emEvt EmailEvento

	if err := json.Unmarshal(body, &emEvt); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := emEvt.DeletaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func deletaByIdDispositivo(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var emEvt EmailEvento

	if err := json.Unmarshal(body, &emEvt); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := emEvt.DeletaByIdDispositivo(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}
