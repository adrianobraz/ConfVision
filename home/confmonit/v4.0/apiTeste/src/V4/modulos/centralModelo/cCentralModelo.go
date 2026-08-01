package centralModeloV4

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

	var cm CentralModelo

	if err := json.Unmarshal(body, &cm); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cm.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.Dados(w, http.StatusOK, cm.ID_Modelo)
}

func getDadosById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cm CentralModelo

	if err := json.Unmarshal(body, &cm); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cm.GetDadosById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.Dados(w, http.StatusOK, cm)

}

func alterarById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cm CentralModelo

	if err := json.Unmarshal(body, &cm); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cm.AlterarById(); err != nil {
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

	var cm CentralModelo

	if err := json.Unmarshal(body, &cm); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cm.DeletaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.OK(w)
}

func deletaAllByIdFabricante(w http.ResponseWriter, r *http.Request) {}

func listarByIdFabricante(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cm CentralModelo

	if err := json.Unmarshal(body, &cm); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []CentralModelo
	if err := cm.ListarByIdFabricante(&lista); err != nil {
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

	var cm CentralModelo

	if err := json.Unmarshal(body, &cm); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cm.GetAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.Dados(w, http.StatusOK, cm.Ativo)
}

func setAtivoById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cm CentralModelo

	if err := json.Unmarshal(body, &cm); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cm.SetAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.Dados(w, http.StatusOK, cm.Ativo)
}

func invereteAtivoById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cm CentralModelo

	if err := json.Unmarshal(body, &cm); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cm.InvereteAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.Dados(w, http.StatusOK, cm.Ativo)
}

func getEletrificadorById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cm CentralModelo

	if err := json.Unmarshal(body, &cm); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cm.GetEletrificadorById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.Dados(w, http.StatusOK, cm.Eletrificador)
}

func setEletrificadorById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cm CentralModelo

	if err := json.Unmarshal(body, &cm); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cm.SetEletrificadorById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.Dados(w, http.StatusOK, cm.Eletrificador)
}

func invereteEletrificadorById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cm CentralModelo

	if err := json.Unmarshal(body, &cm); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cm.InvereteEletrificadorById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.Dados(w, http.StatusOK, cm.Eletrificador)
}
