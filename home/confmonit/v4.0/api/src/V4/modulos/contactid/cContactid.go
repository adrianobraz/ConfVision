package contactidV4

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

	var ci ContactId

	if err := json.Unmarshal(body, &ci); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := ci.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.Dados(w, http.StatusOK, ci.Codigo)
}

func getDadosByCodigo(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var ci ContactId

	if err := json.Unmarshal(body, &ci); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := ci.GetDadosByCodigo(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.Dados(w, http.StatusOK, ci)
}

func GetDadosById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var ci ContactId

	if err := json.Unmarshal(body, &ci); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := ci.GetDadosById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.Dados(w, http.StatusOK, ci)
}

func alteraById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var ci ContactId

	if err := json.Unmarshal(body, &ci); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := ci.AlteraById(); err != nil {
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

	var ci ContactId

	if err := json.Unmarshal(body, &ci); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := ci.DeletaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.OK(w)
}

func listaPadrao(w http.ResponseWriter, r *http.Request) {

	var ci ContactId

	var lista []ContactId

	if err := ci.ListaPadrao(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) <= 0 {
		respApp.Vazio(w)
	} else {
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func listaByIdVinculo(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var ci ContactId

	if err := json.Unmarshal(body, &ci); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []ContactId

	if err := ci.ListaByIdVinculo(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) <= 0 {
		respApp.Vazio(w)
	} else {
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func listaByGrupo(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var ci ContactId

	if err := json.Unmarshal(body, &ci); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []ContactId

	if err := ci.ListaByGrupo(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) <= 0 {
		respApp.Vazio(w)
	} else {
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func listaGrupos(w http.ResponseWriter, r *http.Request) {

	var ci ContactId
	var lista []ListaGrupos

	if err := ci.ListaGrupos(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) <= 0 {
		respApp.Vazio(w)
	} else {
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func getNivelById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var ci ContactId

	if err := json.Unmarshal(body, &ci); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	nivel, err := ci.GetNivelById()
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, nivel)
}
