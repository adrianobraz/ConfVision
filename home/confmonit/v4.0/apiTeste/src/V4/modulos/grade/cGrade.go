package gradeV4

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
	var gra Grade

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &gra); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := gra.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, gra.ID_Grade)
}

func getDadosById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var gra Grade

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &gra); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := gra.GetDadosById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, gra)
}

func alteraById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var gra Grade

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &gra); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := gra.AlteraById(); err != nil {
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
	var gra Grade

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &gra); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := gra.DeletaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.OK(w)
}

func deletaAllByIdDispositivo(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var gra Grade

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &gra); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := gra.DeletaAllByIdDispositivo(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.OK(w)
}

func listarByIdDispositivo(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var gra Grade

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &gra); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Grade
	// Executar a operação
	if err := gra.ListarByIdDispositivo(&lista); err != nil {
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

func getAtivo(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var gra Grade

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &gra); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := gra.GetAtivo(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, gra.Ativo)
}

func setAtivo(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var gra Grade

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &gra); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := gra.SetAtivo(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, gra.Ativo)
}

func desabiltaAllAtivoByIdDispositivo(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var gra Grade

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &gra); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := gra.DeletaAllByIdDispositivo(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.OK(w)
}

func inverteAtivo(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var gra Grade

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &gra); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := gra.InverteAtivo(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, gra.Ativo)
}
