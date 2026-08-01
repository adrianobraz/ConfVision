package informativo

import (
	"api/src/V4/respApp"
	"encoding/json"
	"io"
	"net/http"
)

func insere(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var inf Informativo

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &inf); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := inf.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, inf.ID_Informativo)
}

func getDadosById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var inf Informativo

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &inf); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := inf.GetDadosById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, inf)
}

func alteraById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var inf Informativo

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &inf); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := inf.AlteraById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.OK(w)
}

func deletaById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var inf Informativo

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &inf); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := inf.DeletaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.OK(w)
}

func listaByIdVinculo(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var inf Informativo

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &inf); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Informativo

	// Executar a operação
	if err := inf.ListaByIdVinculo(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	if len(lista) > 0 {
		respApp.Dados(w, http.StatusOK, lista)
	} else {
		respApp.Vazio(w)
	}

}
