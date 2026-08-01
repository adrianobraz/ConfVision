package pacotev4

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
	var p Pacote

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &p); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := p.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, p.ID_Pacote)
}

func getDadosById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var p Pacote

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &p); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := p.GetDadosById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, p)
}

func alteraById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var p Pacote

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &p); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := p.AlteraById(); err != nil {
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
	var p Pacote

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &p); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := p.DeletaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.OK(w)
}

func deletaAllByVinculo(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var p Pacote

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &p); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := p.DeletaAllByVinculo(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.OK(w)
}

func listaByVinculo(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var p Pacote

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &p); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Pacote
	// Executar a operação
	if err := p.ListaByVinculo(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) <= 0 {
		respApp.Vazio(w)
	} else {
		// Envia resposta para o APP
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func getAtivoById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var p Pacote

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &p); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := p.GetAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, p.Ativo)
}

func setAtivoById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var p Pacote

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &p); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := p.SetAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, p.Ativo)
}

func inverteAtivoById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var p Pacote

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &p); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := p.InverteAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, p.Ativo)
}
